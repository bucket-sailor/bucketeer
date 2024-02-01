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

package api

import (
	"crypto/rand"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/bucket-sailor/bucketeer/internal/api/v1alpha1"
	"github.com/bucket-sailor/writablefs"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/labstack/echo/v4"
	"k8s.io/utils/ptr"
)

const (
	listCacheMaxSize = 100
	listCacheTTL     = 5 * time.Minute
)

// Server is yet another remote filesystem API.
type Server struct {
	logger    *slog.Logger
	fsys      writablefs.FS
	listCache *expirable.LRU[string, []v1alpha1.FileInfo]
}

func NewServer(logger *slog.Logger, fsys writablefs.FS) *Server {
	return &Server{
		logger:    logger,
		fsys:      fsys,
		listCache: expirable.NewLRU[string, []v1alpha1.FileInfo](listCacheMaxSize, nil, listCacheTTL),
	}
}

func (s *Server) Register(e *echo.Echo) {
	g := e.Group("/api/v1alpha1")
	g.GET("/fs/info", s.Info)
	g.GET("/fs/list", s.List)
	g.POST("/fs/mkdir", s.Mkdir)
	g.POST("/fs/remove", s.Remove)
	g.POST("/fs/rename", s.Rename)
}

func (s *Server) Info(c echo.Context) error {
	fi, err := s.fsys.Stat(c.QueryParam("path"))
	if err != nil {
		if errors.Is(err, writablefs.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, v1alpha1.ErrorResponse{Message: err.Error()})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, &v1alpha1.FileInfo{
		Name:         fi.Name(),
		IsDir:        fi.IsDir(),
		Size:         ptr.To(fi.Size()),
		LastModified: ptr.To(v1alpha1.Time(fi.ModTime())),
	})
}

func (s *Server) List(c echo.Context) error {
	var err error
	var startIndex, stopIndex int64

	startIndexParam := c.QueryParam("startIndex")
	if startIndexParam != "" {
		startIndex, err = strconv.ParseInt(startIndexParam, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, v1alpha1.ErrorResponse{Message: err.Error()})
		}
	}

	stopIndexParam := c.QueryParam("stopIndex")
	if stopIndexParam != "" {
		stopIndex, err = strconv.ParseInt(stopIndexParam, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, v1alpha1.ErrorResponse{Message: err.Error()})
		}
	}

	populateCache := func(id string) ([]v1alpha1.FileInfo, error) {
		entries, err := s.fsys.ReadDir(c.QueryParam("path"))
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
		}

		files := make([]v1alpha1.FileInfo, len(entries))
		for i, entry := range entries {
			fi, err := toFileInfo(entry)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
			}

			files[i] = *fi
		}

		s.listCache.Add(id, files)

		return files, nil
	}

	var files []v1alpha1.FileInfo

	id := c.QueryParam("id")
	if id == "" {
		id = generateID(8)

		files, err = populateCache(id)
		if err != nil {
			return err
		}
	} else {
		var ok bool
		files, ok = s.listCache.Get(id)
		if !ok {
			files, err = populateCache(id)
			if err != nil {
				return err
			}
		}
	}

	if startIndex == 0 && stopIndex == 0 {
		return c.JSON(http.StatusOK, &v1alpha1.ListResponse{
			ID:    id,
			Files: files,
		})
	}

	if startIndex >= stopIndex {
		return echo.NewHTTPError(http.StatusBadRequest, v1alpha1.ErrorResponse{Message: "start index must be less than stop index"})
	}

	startIndex = min(max(startIndex, 0), int64(len(files)))
	stopIndex = min(stopIndex, int64(len(files)))

	return c.JSON(http.StatusOK, &v1alpha1.ListResponse{
		ID:    id,
		Files: files[startIndex:stopIndex],
	})
}

func (s *Server) Mkdir(c echo.Context) error {
	if err := s.fsys.MkdirAll(c.FormValue("path")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}

func (s *Server) Remove(c echo.Context) error {
	if err := s.fsys.RemoveAll(c.FormValue("path")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) Rename(c echo.Context) error {
	if err := s.fsys.Rename(c.FormValue("oldPath"), c.FormValue("newPath")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func toFileInfo(entry writablefs.DirEntry) (*v1alpha1.FileInfo, error) {
	resp := &v1alpha1.FileInfo{
		Name:  entry.Name(),
		IsDir: entry.IsDir(),
	}

	if fi, err := entry.Info(); err == nil {
		resp.Size = ptr.To(fi.Size())
		resp.LastModified = ptr.To(v1alpha1.Time(fi.ModTime()))
	}

	return resp, nil
}

func generateID(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	for i := range b {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}

		b[i] = letters[r.Int64()]
	}

	return string(b)
}
