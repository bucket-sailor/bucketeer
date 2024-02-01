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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bucket-sailor/bucketeer/internal/api"
	"github.com/bucket-sailor/bucketeer/internal/files"
	"github.com/bucket-sailor/bucketeer/internal/util"
	"github.com/bucket-sailor/bucketeer/web"
	"github.com/bucket-sailor/writablefs/s3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/browser"
	slogecho "github.com/samber/slog-echo"
	"github.com/urfave/cli/v2"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	beforeAll := func(c *cli.Context) error {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: (*slog.Level)(c.Generic("log-level").(*logLevelFlag)),
		}))

		return nil
	}

	sharedFlags := []cli.Flag{
		&cli.GenericFlag{
			Name:  "log-level",
			Usage: "Set the log level",
			Value: fromLogLevel(slog.LevelInfo),
		},
		&cli.BoolFlag{
			Name:  "non-interactive",
			Usage: "Whether to run in non-interactive mode",
		},
	}

	app := &cli.App{
		Name:      "bucketeer",
		Usage:     "The ultimate S3 bucket explorer",
		ArgsUsage: "<bucket name>",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Usage:   "The address to listen on",
				Aliases: []string{"l"},
				EnvVars: []string{"LISTEN_ADDR"},
				Value:   ":0",
			},
			&cli.BoolFlag{
				Name:  "disable-cors",
				Usage: "Disable CORS protection (for local development)",
			},
			&cli.StringFlag{
				Name:    "endpoint-url",
				Usage:   "The URL of your S3 server",
				EnvVars: []string{"AWS_ENDPOINT_URL_S3"},
				Value:   "https://s3.amazonaws.com",
			},
			&cli.StringFlag{
				Name:    "access-key-id",
				Usage:   "Your S3 access key ID",
				EnvVars: []string{"AWS_ACCESS_KEY_ID"},
			},
			&cli.StringFlag{
				Name:    "secret-access-key",
				Usage:   "Your S3 secret access key",
				EnvVars: []string{"AWS_SECRET_ACCESS_KEY"},
			},
			&cli.StringFlag{
				Name:    "region",
				Usage:   "The region of your S3 server",
				EnvVars: []string{"AWS_DEFAULT_REGION"},
			},
			&cli.StringFlag{
				Name:    "ca-bundle",
				Usage:   "The path to the CA bundle to use for TLS verification",
				EnvVars: []string{"AWS_CA_BUNDLE"},
			},
			&cli.BoolFlag{
				Name:    "no-verify-ssl",
				Usage:   "Whether the TLS client should skip TLS verification",
				EnvVars: []string{"AWS_NO_VERIFY_SSL"},
			},
		}, sharedFlags...),
		Before: beforeAll,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				_ = cli.ShowAppHelp(c)

				return fmt.Errorf("bucket name argument is required")
			}

			bucketName := c.Args().Get(0)

			accessKeyID := c.String("access-key-id")
			secretAccessKey := c.String("secret-access-key")

			// If the access key ID or secret access key are not set, try to get them from the
			// AWS credentials file.
			if accessKeyID == "" || secretAccessKey == "" {
				logger.Info("Attempting to get credentials from AWS credentials file")

				creds := credentials.NewFileAWSCredentials("", "")
				credValues, err := creds.Get()
				if err != nil {
					return fmt.Errorf("missing s3 credentials: %w", err)
				}

				accessKeyID = credValues.AccessKeyID
				secretAccessKey = credValues.SecretAccessKey
			}

			var tlsClientConfig *tls.Config
			if c.String("ca-bundle") != "" || c.Bool("no-verify-ssl") {
				tlsClientConfig = &tls.Config{
					InsecureSkipVerify: c.Bool("no-verify-ssl"),
				}

				caBundlePath := c.String("ca-bundle")
				if caBundlePath != "" {
					caBundle, err := os.ReadFile(caBundlePath)
					if err != nil {
						return fmt.Errorf("failed to read ca bundle: %w", err)
					}

					caCertPool := x509.NewCertPool()
					if !caCertPool.AppendCertsFromPEM(caBundle) {
						return fmt.Errorf("failed to append ca bundle to certificate pool")
					}

					tlsClientConfig.RootCAs = caCertPool
				}
			}

			opts := s3.Options{
				EndpointURL:     c.String("endpoint-url"),
				Region:          c.String("region"),
				TLSClientConfig: tlsClientConfig,
				Credentials:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
				BucketName:      bucketName,
			}

			fsys, err := s3.NewFS(c.Context, logger, opts)
			if err != nil {
				return fmt.Errorf("failed to create s3 filesystem: %w", err)
			}

			e := echo.New()

			e.Use(slogecho.New(logger))
			e.Use(middleware.Recover())

			// Allow disabling CORS protection for development.
			if c.Bool("disable-cors") {
				e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
					AllowOrigins: []string{"http://localhost:*"},
					AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete},
					AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "Content-Range"},
				}))
			}

			// The rendered React app.
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

			// Handle /api requests.
			apiServer := api.NewServer(logger, fsys)
			apiServer.Register(e)

			// Handle bulk file transfers.
			filesServer, err := files.NewServer(logger, fsys)
			if err != nil {
				return fmt.Errorf("failed to create files server: %w", err)
			}
			defer filesServer.Close()

			filesServer.Register(e)

			if !c.Bool("non-interactive") {
				go func() {
					if err := util.WaitForServerReady(e, 10*time.Second); err != nil {
						logger.Error("Failed to start server", "error", err)

						// shutdown the server
						_ = e.Shutdown(c.Context)
						return
					}

					// parse the host and port from the listener
					_, port, err := net.SplitHostPort(e.Listener.Addr().String())
					if err != nil {
						logger.Warn("Failed to parse listener address", "error", err)

						return
					}

					browseURL := "http://localhost:" + port + "/browse/"

					logger.Info("Opening in the users default browser", "url", browseURL)

					if err := browser.OpenURL(browseURL); err != nil {
						logger.Warn("Failed to open browser", "error", err)
					}
				}()
			}

			return e.Start(c.String("listen"))
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
