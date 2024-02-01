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

package files

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/bucket-sailor/bucketeer/internal/api/v1alpha1"
	"github.com/bucket-sailor/windlass"
	"github.com/bucket-sailor/writablefs"
	"github.com/bucket-sailor/writablefs/dir"

	"github.com/labstack/echo/v4"
)

type Server struct {
	logger       *slog.Logger
	fsys         writablefs.FS
	uploadServer *windlass.Server
	stagingFS    writablefs.FS
	stagingDir   string
}

func NewServer(logger *slog.Logger, fsys writablefs.FS) (*Server, error) {
	stagingDir, err := os.MkdirTemp("", "bucketeer-*")
	if err != nil {
		return nil, err
	}

	stagingFS, err := dir.NewFS(stagingDir)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:       logger,
		fsys:         fsys,
		uploadServer: windlass.NewServer(logger, fsys, windlass.WithStagingFS(stagingFS)),
		stagingFS:    stagingFS,
		stagingDir:   stagingDir,
	}, nil
}

func (s *Server) Close() error {
	return os.RemoveAll(s.stagingDir)
}

func (s *Server) Register(e *echo.Echo) {
	e.POST("/files", echo.WrapHandler(http.StripPrefix("/files", s.uploadServer)))
	e.PATCH("/files/*", echo.WrapHandler(http.StripPrefix("/files/", s.uploadServer)))
	e.GET("/files/:path", s.download)
}

func (s *Server) download(c echo.Context) error {
	path, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, v1alpha1.ErrorResponse{Message: err.Error()})
	}

	fi, err := s.fsys.Stat(path)
	if err != nil {
		// Try listing to see if it's a directory.
		// Some filesystems (e.g. S3) don't support Stat() but do support ReadDir() for directories.
		if _, err := s.fsys.ReadDir(path); err == nil {
			return s.downloadDirectory(c, path)
		}

		if errors.Is(err, writablefs.ErrNotExist) {
			return c.NoContent(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
	}

	if fi.IsDir() {
		return s.downloadDirectory(c, path)
	}

	f, err := s.fsys.OpenFile(path, writablefs.O_RDONLY)
	if err != nil {
		if errors.Is(err, writablefs.ErrNotExist) {
			return c.NoContent(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, v1alpha1.ErrorResponse{Message: err.Error()})
	}
	defer f.Close()

	// Don't attempt to preview the file in the browser.
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fi.Name()))

	// ServeContent() will take care of things like range requests, etc.
	http.ServeContent(c.Response().Writer, c.Request(), fi.Name(), fi.ModTime(), f)
	return nil
}

func (s *Server) downloadDirectory(c echo.Context, name string) error {
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+name+".zip")
	c.Response().Header().Set("Content-Type", "application/zip")

	pr, pw := io.Pipe()
	defer pr.Close()

	go func() {
		defer pw.Close()

		zw := zip.NewWriter(pw)
		defer zw.Close()

		err := writablefs.WalkDir(s.fsys, name, func(path string, d writablefs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				fi, err := d.Info()
				if err != nil {
					return err
				}

				header, err := zip.FileInfoHeader(fi)
				if err != nil {
					return err
				}
				header.Name = path
				header.Method = zip.Deflate

				zfw, err := zw.CreateHeader(header)
				if err != nil {
					return err
				}

				f, err := s.fsys.OpenFile(path, writablefs.O_RDONLY)
				if err != nil {
					return err
				}
				defer f.Close()

				_, err = io.Copy(zfw, f)
				return err
			}
			return nil
		})
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	_, err := io.Copy(c.Response().Writer, pr)
	return err
}
