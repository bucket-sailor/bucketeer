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

package files_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bucket-sailor/windlass"
	"github.com/bucket-sailor/writablefs/dir"
	"github.com/dpeckett/bucketeer/internal/files"
	"github.com/dpeckett/bucketeer/internal/utils/testutils"
	"github.com/labstack/echo/v4"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestFilesServer(t *testing.T) {
	logger := slogt.New(t)

	fs, err := dir.FS(t.TempDir())
	require.NoError(t, err)

	s := files.NewServer(logger, fs)

	e := echo.New()
	e.HideBanner = true

	t.Cleanup(func() {
		require.NoError(t, e.Close())
	})

	s.Register(e)

	go func() {
		if err := e.Start(":0"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", "error", err)
		}
	}()

	require.NoError(t, testutils.WaitForServerReady(e, 5*time.Second))

	url := fmt.Sprintf("http://%s/files", e.Listener.Addr().String())

	cli, err := windlass.NewClient(logger, url, nil)
	require.NoError(t, err)

	t.Run("Upload and Download File", func(t *testing.T) {
		size := int64(10000000) // 10MB

		buf := make([]byte, size)
		_, err = io.ReadFull(rand.Reader, buf)
		require.NoError(t, err)

		ctx := context.Background()
		err = cli.Upload(ctx, "test.bin", bytes.NewReader(buf), size)
		require.NoError(t, err)

		chunk, err := partialDownload(url+"/test.bin", 1000, 1999)
		require.NoError(t, err)

		require.Equal(t, buf[1000:2000], chunk)
	})

	t.Run("Download Directory", func(t *testing.T) {
		err := fs.MkdirAll("testdir")
		require.NoError(t, err)

		err = fs.MkdirAll("testdir/subdir1")
		require.NoError(t, err)

		f, err := fs.OpenFile("testdir/subdir1/foo.txt", os.O_CREATE|os.O_WRONLY)
		require.NoError(t, err)

		_, err = f.Write([]byte("test1"))
		require.NoError(t, err)

		require.NoError(t, f.Close())

		err = fs.MkdirAll("testdir/subdir2")
		require.NoError(t, err)

		f, err = fs.OpenFile("testdir/subdir2/bar.txt", os.O_CREATE|os.O_WRONLY)
		require.NoError(t, err)

		_, err = f.Write([]byte("test2"))
		require.NoError(t, err)

		require.NoError(t, f.Close())

		downloaded, err := partialDownload(url+"/testdir", 0, 0)
		require.NoError(t, err)

		zr, err := zip.NewReader(bytes.NewReader(downloaded), int64(len(downloaded)))
		require.NoError(t, err)

		require.Len(t, zr.File, 2)
	})
}

func partialDownload(url string, startByte, endByte int64) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if startByte > 0 || endByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
