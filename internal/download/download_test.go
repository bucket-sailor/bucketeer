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

package download_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/bucket-sailor/bucketeer/internal/download"
	"github.com/bucket-sailor/bucketeer/internal/util"
	"github.com/bucket-sailor/writablefs"
	"github.com/bucket-sailor/writablefs/dirfs"
	"github.com/labstack/echo/v4"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

func TestDownload(t *testing.T) {
	logger := slogt.New(t)

	testDir := t.TempDir()

	fsys, err := dirfs.New(testDir)
	require.NoError(t, err)

	// Create a file to download
	err = fsys.MkdirAll("test/folder")
	require.NoError(t, err)

	f, err := fsys.OpenFile("test/folder/file.bin", writablefs.FlagReadWrite|writablefs.FlagCreate)
	require.NoError(t, err)

	size := int64(10000000)
	_, err = io.CopyN(f, rand.Reader, size)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	e := echo.New()
	e.HideBanner = true

	downloadServerPath, downloadServer := download.NewServer(logger, fsys)
	e.Any(downloadServerPath+"*", echo.WrapHandler(downloadServer))

	go func() {
		if err := e.StartH2CServer(":0", &http2.Server{}); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", "error", err)
		}
	}()
	t.Cleanup(func() {
		require.NoError(t, e.Close())
	})

	err = util.WaitForServerReady(e, 10*time.Second)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s", e.Listener.Addr().String())

	t.Run("Download File", func(t *testing.T) {
		expectedSum, err := fileChecksum(fsys, "test/folder/file.bin")
		require.NoError(t, err)

		// Download the file
		h := sha256.New()

		err = downloadFile(context.Background(), baseURL, "test/folder/file.bin", h)
		require.NoError(t, err)

		actualSum := hex.EncodeToString(h.Sum(nil))

		assert.Equal(t, expectedSum, actualSum)
	})

	t.Run("Download Directory", func(t *testing.T) {
		var buf bytes.Buffer

		err = downloadFile(context.Background(), baseURL, "test/", &buf)
		require.NoError(t, err)

		// Check the contents of the zip file
		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		require.NoError(t, err)

		assert.Len(t, r.File, 1)
		assert.Equal(t, "test/folder/file.bin", r.File[0].Name)
	})
}

func downloadFile(ctx context.Context, baseURL, path string, w io.Writer) error {
	downloadURL := fmt.Sprintf("%s/files/download/%s", baseURL, url.QueryEscape(path))

	resp, err := http.DefaultClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func fileChecksum(fsys writablefs.FS, path string) (string, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
