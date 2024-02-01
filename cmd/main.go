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

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/bucket-sailor/writablefs/s3"
	"github.com/dpeckett/bucketeer/internal/api"
	"github.com/dpeckett/bucketeer/internal/files"
	"github.com/dpeckett/bucketeer/web"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/urfave/cli/v2"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	app := &cli.App{
		Name:  "bucketeer",
		Usage: "The ultimate S3 bucket explorer",
		Flags: []cli.Flag{
			&cli.GenericFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "Set the log level",
				Value:   fromLogLevel(slog.LevelInfo),
			},
		},
		Before: func(c *cli.Context) error {
			logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: (*slog.Level)(c.Generic("log-level").(*logLevelFlag)),
			}))

			return nil
		},
		Action: func(c *cli.Context) error {
			e := echo.New()

			e.Use(middleware.Logger())
			e.Use(middleware.Recover())

			// For testing allow all local origins.
			e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: []string{"http://localhost:*"},
				AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete},
				AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "Content-Range"},
			}))

			webFS, err := web.GetFS()
			if err != nil {
				return fmt.Errorf("failed to get web filesystem: %w", err)
			}

			webFSServer := http.FileServer(webFS)

			e.GET("/", func(c echo.Context) error {
				return c.Redirect(http.StatusMovedPermanently, "/browse/")
			})

			// React.
			e.GET("/browse/*", func(c echo.Context) error {
				c.Request().URL.Path = "/"
				webFSServer.ServeHTTP(c.Response(), c.Request())
				return nil
			})

			// Assets etc.
			e.GET("/*", echo.WrapHandler(webFSServer))

			// TODO: support getting credentials from the environment etc.

			opts := s3.Options{
				Endpoint:    "localhost:32842",
				Credentials: credentials.NewStaticV4("admin", "admin", ""),
				BucketName:  "test",
			}

			fsys, err := s3.NewFS(c.Context, logger, opts)
			if err != nil {
				return fmt.Errorf("failed to create s3 filesystem: %w", err)
			}

			apiServer := api.NewServer(logger, fsys)
			apiServer.Register(e)

			filesServer, err := files.NewServer(logger, fsys)
			if err != nil {
				return fmt.Errorf("failed to create files server: %w", err)
			}
			defer filesServer.Close()

			filesServer.Register(e)

			return e.Start(":8082")
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("Failed to run application", "error", err)
		os.Exit(1)
	}
}

type logLevelFlag slog.Level

func fromLogLevel(l slog.Level) *logLevelFlag {
	f := logLevelFlag(l)
	return &f
}

func (f *logLevelFlag) Set(value string) error {
	return (*slog.Level)(f).UnmarshalText([]byte(value))
}

func (f *logLevelFlag) String() string {
	return (*slog.Level)(f).String()
}
