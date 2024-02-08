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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime"

	"connectrpc.com/connect"
	"github.com/bucket-sailor/bucketeer/internal/gen/upload/v1alpha1"
	"github.com/bucket-sailor/bucketeer/internal/gen/upload/v1alpha1/v1alpha1connect"
	"github.com/bucket-sailor/queue"
	"github.com/bucket-sailor/writablefs"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	cacheDir      = ".bucketeer"
	xAttrChecksum = "bucketeer.checksum"
	xAttrPath     = "bucketeer.path"
	xAttrComplete = "bucketeer.complete"
	xAttrError    = "bucketeer.error"
)

type Server struct {
	http.Handler
	logger  *slog.Logger
	fsys    writablefs.FS
	cacheFS writablefs.FS
	// completionQueue is a queue for processing completions.
	// We process these outside the request handler as they may
	// take a some time to complete.
	completionQueue *queue.Queue
}

func NewServer(logger *slog.Logger, fsys, cacheFS writablefs.FS) (string, http.Handler) {
	s := &Server{
		logger:          logger.WithGroup("upload"),
		fsys:            fsys,
		cacheFS:         cacheFS,
		completionQueue: queue.NewQueue(runtime.NumCPU()),
	}

	var path string
	path, s.Handler = v1alpha1connect.NewUploadHandler(s)

	s.Handler = http.StripPrefix("/api", s.Handler)

	return "/api" + path, s
}

func (s *Server) New(ctx context.Context, req *connect.Request[v1alpha1.NewRequest]) (*connect.Response[wrapperspb.StringValue], error) {
	if req.Msg.Size == 0 || req.Msg.Path == "" || req.Msg.Checksum == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing required arguments"))
	}

	uploadID := uuid.New().String()

	cachePath := filepath.Join(cacheDir, uploadID)
	f, err := s.cacheFS.OpenFile(cachePath, writablefs.FlagWriteOnly|writablefs.FlagCreate)
	if err != nil {
		// Create the cache directory if it doesn't exist.
		if errors.Is(err, writablefs.ErrNotExist) {
			if err := s.cacheFS.MkdirAll(cacheDir); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error creating cache directory: %w", err))
			}

			f, err = s.cacheFS.OpenFile(cachePath, writablefs.FlagWriteOnly|writablefs.FlagCreate)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error creating cache file: %w", err))
			}
		} else {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error opening cache file: %w", err))
		}
	}
	defer f.Close()

	if err := f.Truncate(req.Msg.Size); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error truncating cache file: %w", err))
	}

	xattrs, err := f.XAttrs()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error getting xattrs: %w", err))
	}

	if err := xattrs.Set(xAttrChecksum, []byte(req.Msg.Checksum)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error setting checksum xattr: %w", err))
	}

	if err := xattrs.Set(xAttrPath, []byte(req.Msg.Path)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error setting path xattr: %w", err))
	}

	if err := xattrs.Sync(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error syncing xattrs: %w", err))
	}

	return &connect.Response[wrapperspb.StringValue]{
		Msg: &wrapperspb.StringValue{Value: uploadID},
	}, nil
}

func (s *Server) Abort(ctx context.Context, req *connect.Request[wrapperspb.StringValue]) (*connect.Response[emptypb.Empty], error) {
	uploadID := req.Msg.Value
	if uploadID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing required argument"))
	}

	if _, err := uuid.Parse(uploadID); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid upload ID: %w", err))
	}

	cachePath := filepath.Join(cacheDir, uploadID)

	if err := s.cacheFS.RemoveAll(cachePath); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error removing cache file: %w", err))
	}

	return &connect.Response[emptypb.Empty]{}, nil
}

func (s *Server) Complete(ctx context.Context, req *connect.Request[wrapperspb.StringValue]) (*connect.Response[emptypb.Empty], error) {
	uploadID := req.Msg.Value

	if _, err := uuid.Parse(uploadID); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid upload ID: %w", err))
	}

	cachePath := filepath.Join(cacheDir, uploadID)

	s.completionQueue.Add(func() error {
		completeFn := func() error {
			f, err := s.cacheFS.OpenFile(cachePath, writablefs.FlagReadWrite)
			if err != nil {
				return err
			}
			defer f.Close()

			xattrs, err := f.XAttrs()
			if err != nil {
				return fmt.Errorf("error getting xattrs: %w", err)
			}

			expectedChecksum, err := xattrs.Get(xAttrChecksum)
			if err != nil {
				return fmt.Errorf("error getting checksum xattr: %w", err)
			}

			dstPath, err := xattrs.Get(xAttrPath)
			if err != nil {
				return fmt.Errorf("error getting path xattr: %w", err)
			}

			if err := verifyChecksum(f, string(expectedChecksum)); err != nil {
				return fmt.Errorf("checksum mismatch: %w", err)
			}

			if err := s.fsys.MkdirAll(filepath.Dir(string(dstPath))); err != nil {
				return err
			}

			if err := copyFile(s.cacheFS, cachePath, s.fsys, string(dstPath)); err != nil {
				return err
			}

			// Truncate the cache file to 0 bytes now that the upload is complete.
			if err := f.Truncate(0); err != nil {
				return err
			}

			return nil
		}

		var completionErr error

		defer func() {
			// Set the complete xattr and any error xattr.
			f, err := s.cacheFS.OpenFile(cachePath, writablefs.FlagWriteOnly)
			if err != nil {
				s.logger.Error("Error creating cache file", "error", err)
			}
			defer f.Close()

			xattrs, err := f.XAttrs()
			if err != nil {
				s.logger.Error("Error opening xattrs", "error", err)
			}

			if completionErr != nil {
				if err := xattrs.Set(xAttrError, []byte(completionErr.Error())); err != nil {
					s.logger.Error("Error setting error xattr", "error", err)
				}
			}

			if err := xattrs.Set(xAttrComplete, []byte("true")); err != nil {
				s.logger.Error("Error setting complete xattr", "error", err)
			}

			if err := xattrs.Sync(); err != nil {
				s.logger.Error("Error syncing xattrs", "error", err)
			}
		}()

		completionErr = completeFn()
		if completionErr != nil {
			s.logger.Error("Error completing upload", "error", completionErr)
		}

		return completionErr
	})

	return &connect.Response[emptypb.Empty]{}, nil
}

func (s *Server) PollForCompletion(ctx context.Context, req *connect.Request[wrapperspb.StringValue]) (*connect.Response[v1alpha1.CompleteResponse], error) {
	uploadID := req.Msg.Value

	if _, err := uuid.Parse(uploadID); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid upload ID: %w", err))
	}

	cachePath := filepath.Join(cacheDir, uploadID)

	f, err := s.cacheFS.OpenFile(cachePath, writablefs.FlagReadOnly)
	if err != nil {
		if errors.Is(err, writablefs.ErrNotExist) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("upload not found"))
		}

		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error opening cache file: %w", err))
	}
	defer f.Close()

	xattrs, err := f.XAttrs()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error opening xattrs: %w", err))
	}

	complete, err := xattrs.Get(xAttrComplete)
	if err != nil && !errors.Is(err, writablefs.ErrNoSuchAttr) {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error getting complete xattr: %w", err))
	}

	if complete == nil || string(complete) != "true" {
		return &connect.Response[v1alpha1.CompleteResponse]{
			Msg: &v1alpha1.CompleteResponse{
				Status: *v1alpha1.CompletionStatus_PENDING.Enum(),
			},
		}, nil
	}

	errorAttr, err := xattrs.Get(xAttrError)
	if err != nil && !errors.Is(err, writablefs.ErrNoSuchAttr) {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("error getting error xattr: %w", err))
	}

	if errorAttr != nil {
		return &connect.Response[v1alpha1.CompleteResponse]{
			Msg: &v1alpha1.CompleteResponse{
				Status: *v1alpha1.CompletionStatus_FAILED.Enum(),
				Error:  string(errorAttr),
			},
		}, nil
	}

	return &connect.Response[v1alpha1.CompleteResponse]{
		Msg: &v1alpha1.CompleteResponse{
			Status: *v1alpha1.CompletionStatus_COMPLETED.Enum(),
		},
	}, nil
}

func copyFile(srcFS writablefs.FS, srcPath string, dstFS writablefs.FS, dstPath string) error {
	src, err := srcFS.OpenFile(srcPath, writablefs.FlagReadOnly)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := dstFS.OpenFile(dstPath, writablefs.FlagWriteOnly|writablefs.FlagCreate)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
