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

package upload_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bucket-sailor/bucketeer/internal/upload"
	"github.com/bucket-sailor/bucketeer/internal/util"
	"github.com/bucket-sailor/writablefs/dirfs"
	"github.com/labstack/echo/v4"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

func TestUpload(t *testing.T) {
	logger := slogt.New(t)

	testDir := t.TempDir()

	serverDir := filepath.Join(testDir, "server")
	cacheDir := filepath.Join(testDir, "cache")

	fsys, err := dirfs.New(serverDir)
	require.NoError(t, err)

	cacheFS, err := dirfs.New(cacheDir)
	require.NoError(t, err)

	e := echo.New()
	e.HideBanner = true

	uploadServerPath, uploadServer := upload.NewServer(logger, fsys, cacheFS)
	e.Any(uploadServerPath+"*", echo.WrapHandler(uploadServer))

	chunkServerPath, chunkServer := upload.NewChunkServer(logger, fsys, cacheFS)
	e.Any(chunkServerPath, echo.WrapHandler(chunkServer))

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

	f, err := os.Create(filepath.Join(t.TempDir(), "test.bin"))
	require.NoError(t, err)

	size := int64(100000000)
	_, err = io.CopyN(f, rand.Reader, size)
	require.NoError(t, err)

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s", e.Listener.Addr().String())

	c, err := upload.NewClient(logger, baseURL, &upload.ClientOptions{
		NumConnections: 1,
		ChunkSizeBytes: size,
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = c.Upload(ctx, filepath.Join(t.Name(), "test.bin"), f, size)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(serverDir, t.Name(), "test.bin"))

	expectedSum, err := fileChecksum(f.Name())
	require.NoError(t, err)

	actualSum, err := fileChecksum(filepath.Join(serverDir, t.Name(), "test.bin"))
	require.NoError(t, err)

	assert.Equal(t, expectedSum, actualSum)
}

func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
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
