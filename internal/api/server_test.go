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

package api_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bucket-sailor/writablefs/dir"
	"github.com/dpeckett/bucketeer/internal/api"
	"github.com/dpeckett/bucketeer/internal/api/v1alpha1"
	"github.com/dpeckett/bucketeer/internal/utils/testutils"
	"github.com/labstack/echo/v4"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIServer(t *testing.T) {
	logger := slogt.New(t)

	testDir := t.TempDir()

	fs, err := dir.FS(testDir)
	require.NoError(t, err)

	s := api.NewServer(logger, fs)

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

	url := fmt.Sprintf("http://%s/api/v1alpha1/fs", e.Listener.Addr().String())

	c := newClient(url)

	err = os.MkdirAll(filepath.Join(testDir, "testdir"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(testDir, "testdir/testfile"), []byte("foo"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(testDir, "testdir/testfile2"), []byte("bar"), 0o644)
	require.NoError(t, err)

	t.Run("Info", func(t *testing.T) {
		fi, err := c.info("testdir")
		require.NoError(t, err)

		assert.Equal(t, "testdir", fi.Name)
		assert.True(t, fi.IsDir)
		assert.NotNil(t, fi.Size)
		assert.NotZero(t, fi.LastModified)
	})

	t.Run("List", func(t *testing.T) {
		lr, err := c.list("testdir")
		require.NoError(t, err)

		assert.Len(t, lr.Files, 2)
		assert.Equal(t, "testfile", lr.Files[0].Name)
		assert.Equal(t, "testfile2", lr.Files[1].Name)
	})

	t.Run("Mkdir", func(t *testing.T) {
		err := c.mkdir("testdir2")
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(testDir, "testdir2"))
	})

	t.Run("Remove", func(t *testing.T) {
		err := c.mkdir("testdir3")
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(testDir, "testdir3"))

		err = c.remove("testdir3")
		require.NoError(t, err)

		assert.NoDirExists(t, filepath.Join(testDir, "testdir3"))
	})

	t.Run("Rename", func(t *testing.T) {
		err := c.mkdir("testdir4")
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(testDir, "testdir4"))

		err = c.rename("testdir4", "testdir5")
		require.NoError(t, err)

		assert.NoDirExists(t, filepath.Join(testDir, "testdir4"))
		assert.DirExists(t, filepath.Join(testDir, "testdir5"))
	})
}

type client struct {
	*http.Client
	url string
}

func newClient(url string) *client {
	return &client{
		Client: &http.Client{},
		url:    url,
	}
}

func (c *client) info(path string) (*v1alpha1.FileInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/info?path="+url.QueryEscape(path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("unexpected status code %d", resp.StatusCode)
		var e v1alpha1.ErrorResponse
		if nestedErr := json.NewDecoder(resp.Body).Decode(&e); nestedErr != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%w: %q", err, e.Message)
	}

	var fi v1alpha1.FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&fi); err != nil {
		return nil, err
	}

	return &fi, nil
}

func (c *client) list(path string) (*v1alpha1.ListResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/list?path="+url.QueryEscape(path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("unexpected status code %d", resp.StatusCode)
		var e v1alpha1.ErrorResponse
		if nestedErr := json.NewDecoder(resp.Body).Decode(&e); nestedErr != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%w: %q", err, e.Message)
	}

	var lr v1alpha1.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}

	return &lr, nil
}

func (c *client) mkdir(path string) error {
	values := url.Values{}
	values.Set("path", path)

	req, err := http.NewRequest(http.MethodPost, c.url+"/mkdir", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("unexpected status code %d", resp.StatusCode)
		var e v1alpha1.ErrorResponse
		if nestedErr := json.NewDecoder(resp.Body).Decode(&e); nestedErr != nil {
			return err
		}

		return fmt.Errorf("%w: %q", err, e.Message)
	}

	return nil
}

func (c *client) remove(path string) error {
	values := url.Values{}
	values.Set("path", path)

	req, err := http.NewRequest(http.MethodPost, c.url+"/remove", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("unexpected status code %d", resp.StatusCode)
		var e v1alpha1.ErrorResponse
		if nestedErr := json.NewDecoder(resp.Body).Decode(&e); nestedErr != nil {
			return err
		}

		return fmt.Errorf("%w: %q", err, e.Message)
	}

	return nil
}

func (c *client) rename(oldPath, newPath string) error {
	values := url.Values{}
	values.Set("oldPath", oldPath)
	values.Set("newPath", newPath)

	req, err := http.NewRequest(http.MethodPost, c.url+"/rename", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("unexpected status code %d", resp.StatusCode)
		var e v1alpha1.ErrorResponse
		if nestedErr := json.NewDecoder(resp.Body).Decode(&e); nestedErr != nil {
			return err
		}

		return fmt.Errorf("%w: %q", err, e.Message)
	}

	return nil
}
