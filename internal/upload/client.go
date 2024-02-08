/* SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * Copyright 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package upload

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/avast/retry-go/v4"
	"github.com/bucket-sailor/bucketeer/internal/gen/upload/v1alpha1"
	"github.com/bucket-sailor/bucketeer/internal/gen/upload/v1alpha1/v1alpha1connect"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/jinzhu/copier"
	"github.com/rogpeppe/go-internal/par"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ClientOptions are options for configuring the behavior of the upload client.
type ClientOptions struct {
	// NumConnections is the number of concurrent connections to use when uploading chunks.
	NumConnections int
	// ChunkSizeBytes is the size of each chunk.
	ChunkSizeBytes int64
	// MaxRetryAttempts is the maximum number of retry attempts to make before giving up.
	MaxRetryAttempts int
	// TLSClientConfig is the optional TLS configuration to use when making requests.
	TLSClientConfig *tls.Config
}

type Client struct {
	logger     *slog.Logger
	baseURL    string
	httpClient *http.Client
	apiClient  v1alpha1connect.UploadClient
	opts       *ClientOptions
}

func NewClient(logger *slog.Logger, baseURL string, opts *ClientOptions) (*Client, error) {
	baseOpts := ClientOptions{
		NumConnections:   1,
		ChunkSizeBytes:   16000000, // 16MB
		MaxRetryAttempts: 3,
	}

	if opts != nil {
		if err := copier.CopyWithOption(&baseOpts, opts, copier.Option{IgnoreEmpty: true}); err != nil {
			return nil, fmt.Errorf("failed to copy options: %w", err)
		}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseOpts.TLSClientConfig != nil {
		transport.TLSClientConfig = baseOpts.TLSClientConfig
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	return &Client{
		logger:     logger.WithGroup("upload"),
		baseURL:    baseURL,
		httpClient: httpClient,
		apiClient:  v1alpha1connect.NewUploadClient(httpClient, baseURL+"/api/"),
		opts:       &baseOpts,
	}, nil
}

// Upload uploads a file to the server, you must provide a ReaderAt so that chunks can be read concurrently.
func (c *Client) Upload(ctx context.Context, path string, r io.ReaderAt, size int64) error {
	expectedChecksum, err := checksum(io.NewSectionReader(r, 0, size), algorithmXXH64)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	uploadIDResp, err := c.apiClient.New(ctx, connect.NewRequest(&v1alpha1.NewRequest{
		Path:     path,
		Size:     size,
		Checksum: expectedChecksum,
	}))
	if err != nil {
		return fmt.Errorf("failed to create new upload: %w", err)
	}

	uploadID := uploadIDResp.Msg.Value

	if _, err := uuid.Parse(uploadID); err != nil {
		return fmt.Errorf("server returned invalid upload ID: %s", uploadID)
	}

	type chunk struct {
		start int64
		end   int64
	}

	var work par.Work
	for i := int64(0); i < size; i += c.opts.ChunkSizeBytes {
		start := i
		end := start + c.opts.ChunkSizeBytes - 1
		if end >= size {
			end = size - 1
		}

		work.Add(&chunk{
			start: start,
			end:   end,
		})
	}

	errs := make(chan error, c.opts.NumConnections)
	defer close(errs)

	var resultMu sync.Mutex
	var result *multierror.Error

	work.Do(c.opts.NumConnections, func(item any) {
		chk := item.(*chunk)
		if err := c.uploadChunk(ctx, uploadID, r, chk.start, chk.end, size); err != nil {
			resultMu.Lock()
			result = multierror.Append(result, err)
			resultMu.Unlock()
		}
	})

	if err := result.ErrorOrNil(); err != nil {
		return err
	}

	_, err = c.apiClient.Complete(ctx, connect.NewRequest(&wrapperspb.StringValue{Value: uploadID}))
	if err != nil {
		return fmt.Errorf("failed to complete upload: %w", err)
	}

	return c.waitForCompletion(ctx, uploadID)
}

func (c *Client) uploadChunk(ctx context.Context, uploadID string, r io.ReaderAt, start, end, size int64) error {
	return retry.Do(
		func() error {
			pr, pw := io.Pipe()
			multipartWriter := multipart.NewWriter(pw)

			go func() {
				defer pw.Close()

				h := textproto.MIMEHeader{}
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, uploadID))
				h.Set("Content-Type", "application/octet-stream")
				h.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))

				fileWriter, err := multipartWriter.CreatePart(h)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to create file writer: %w", err))
					return
				}

				if _, err = io.Copy(fileWriter, io.NewSectionReader(r, start, end-start+1)); err != nil {
					pw.CloseWithError(fmt.Errorf("failed to copy file data: %w", err))
					return
				}

				if err := multipartWriter.Close(); err != nil {
					pw.CloseWithError(fmt.Errorf("failed to close multipart writer: %w", err))
				}
			}()

			req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/files/upload", pr)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNoContent {
				message, err := io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("failed to upload chunk, status code: %d", resp.StatusCode)
				}

				if resp.StatusCode >= 400 && resp.StatusCode < 500 {
					return retry.Unrecoverable(fmt.Errorf("failed to upload chunk: %s", string(message)))
				}

				return fmt.Errorf("failed to upload chunk: %s", string(message))
			}

			return nil
		},
		retry.Context(ctx),
		retry.Attempts(uint(c.opts.MaxRetryAttempts)),
		retry.OnRetry(func(_ uint, err error) {
			c.logger.Warn("Retrying uploading chunk", "error", err)
		}),
	)
}

func (c *Client) waitForCompletion(ctx context.Context, uploadID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var pollErrors int
	return retry.Do(
		func() error {
			completeResp, err := c.apiClient.PollForCompletion(ctx, connect.NewRequest(&wrapperspb.StringValue{Value: uploadID}))
			if err != nil {
				if pollErrors > c.opts.MaxRetryAttempts {
					return retry.Unrecoverable(fmt.Errorf("failed to poll for completion: %w", err))
				}

				pollErrors++

				return err
			}
			pollErrors = 0

			switch completeResp.Msg.Status {
			case v1alpha1.CompletionStatus_COMPLETED:
				return nil
			case v1alpha1.CompletionStatus_FAILED:
				return retry.Unrecoverable(fmt.Errorf("upload failed: %s", completeResp.Msg.Error))
			default:
				return fmt.Errorf("upload not completed yet") // retry
			}
		},
		retry.Context(ctx),
		retry.Attempts(0),
		retry.MaxDelay(10*time.Second),
		retry.OnRetry(func(n uint, err error) {
			c.logger.Debug("Polling for completion", "error", err)
		}),
	)
}
