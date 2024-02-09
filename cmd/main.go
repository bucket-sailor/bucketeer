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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/bucket-sailor/bucketeer/internal/constants"
	"github.com/bucket-sailor/bucketeer/internal/download"
	"github.com/bucket-sailor/bucketeer/internal/filesystem"
	"github.com/bucket-sailor/bucketeer/internal/telemetry"
	"github.com/bucket-sailor/bucketeer/internal/upload"
	"github.com/bucket-sailor/bucketeer/web"
	"github.com/bucket-sailor/writablefs/dirfs"
	"github.com/bucket-sailor/writablefs/s3fs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-isatty"
	"github.com/minio/minio-go/v7/pkg/credentials"
	slogecho "github.com/samber/slog-echo"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/http2"
)

const (
	banner = `    ____             __        __                
   / __ )__  _______/ /_____  / /____  ___  _____
  / __  / / / / ___/ //_/ _ \/ __/ _ \/ _ \/ ___/
 / /_/ / /_/ / /__/ ,< /  __/ /_/  __/  __/ /    
/_____/\__,_/\___/_/|_|\___/\__/\___/\___/_/     
The Ultimate S3 Bucket Explorer.`
)

func main() {
	var logWriter io.WriteCloser = os.Stderr
	logger := slog.New(slog.NewTextHandler(logWriter, nil))

	beforeAll := func(c *cli.Context) error {
		logFilePath := c.String("log-file")

		if logFilePath != "" {
			err := os.MkdirAll(filepath.Dir(logFilePath), 0o755)
			if err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}

			logWriter, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return fmt.Errorf("failed to open log file: %w", err)
			}
		}

		logger = slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{
			Level: (*slog.Level)(c.Generic("log-level").(*logLevelFlag)),
		}))

		return nil
	}

	afterAll := func(c *cli.Context) error {
		logFilePath := c.String("log-file")

		if logFilePath != "" {
			if err := logWriter.(io.Closer).Close(); err != nil {
				return fmt.Errorf("failed to close log file: %w", err)
			}

			// For any shutdown logs.
			logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		}

		return nil
	}

	// Attempt to detect if we're running in a headless environment.
	headless := !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd())

	var defaultLogFilePath string
	if !headless {
		var err error
		defaultLogFilePath, err = xdg.DataFile("bucketeer/bucketeer.log")
		if err != nil {
			logger.Error("Failed to get defaut log path", "error", err)

			os.Exit(1)
		}
	}

	sharedFlags := []cli.Flag{
		&cli.GenericFlag{
			Name:    "log-level",
			Usage:   "Set the log level",
			EnvVars: []string{"BUCKETEER_LOG_LEVEL"},
			Value:   fromLogLevel(slog.LevelInfo),
		},
		&cli.StringFlag{
			Name:    "log-file",
			Usage:   "The path to the log file, if not set logs will be written to stderr",
			EnvVars: []string{"BUCKETEER_LOG_FILE"},
			Value:   defaultLogFilePath,
		},
		&cli.BoolFlag{
			Name:    "headless",
			Usage:   "Run in headless mode",
			EnvVars: []string{"BUCKETEER_HEADLESS"},
			Value:   headless,
		},
	}

	app := &cli.App{
		Name:      "bucketeer",
		Usage:     "The ultimate S3 bucket explorer",
		ArgsUsage: "<bucket name>",
		Version:   constants.Version,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Usage:   "The address to listen on",
				Aliases: []string{"l"},
				EnvVars: []string{"BUCKETEER_LISTEN_ADDR"},
				Value:   "localhost:16321",
			},
			&cli.BoolFlag{
				Name:    "disable-cors",
				Usage:   "Disable CORS protection",
				EnvVars: []string{"BUCKETEER_DISABLE_CORS"},
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
		After:  afterAll,
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

			opts := s3fs.Options{
				EndpointURL:     c.String("endpoint-url"),
				Region:          c.String("region"),
				TLSClientConfig: tlsClientConfig,
				Credentials:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
				BucketName:      bucketName,
			}

			fsys, err := s3fs.New(c.Context, logger, opts)
			if err != nil {
				return fmt.Errorf("failed to open s3 filesystem: %w", err)
			}

			telemetryReporter := telemetry.NewRemoteReporter(
				c.Context, logger, http.DefaultClient, constants.TelemetryURL)
			defer telemetryReporter.Close()

			err = telemetryReporter.ReportStart(c.Context, c.String("endpoint-url"))
			if err != nil {
				logger.Warn("Failed to report application start", "error", err)
			}

			e := echo.New()
			e.HideBanner = true

			e.Use(slogecho.New(logger))

			recoverConfig := middleware.DefaultRecoverConfig
			recoverConfig.StackSize = 8000000 // Capture as much as practical

			panicReporter := telemetry.NewPanicReporter(logger, telemetryReporter)
			recoverConfig.LogErrorFunc = panicReporter.OnPanic

			e.Use(middleware.RecoverWithConfig(recoverConfig))

			// For local development.
			if c.Bool("disable-cors") {
				e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
					AllowOrigins: []string{"http://localhost:*"},
					AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete},
					AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "Content-Range", "Connect-Protocol-Version"},
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

			// Handle filesystem operations.
			filesystemServerPath, filesystemServer := filesystem.NewServer(logger, fsys)
			e.Any(filesystemServerPath+"*", echo.WrapHandler(filesystemServer))

			// Handle file uploads / downloads.
			cacheDir, err := os.MkdirTemp("", "bucketeer-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(cacheDir)

			// S3 doesn't support partial file writes, so we need to stage files locally before
			// uploading them.
			cacheFS, err := dirfs.New(cacheDir)
			if err != nil {
				return err
			}

			uploadServerPath, uploadServer := upload.NewServer(logger, fsys, cacheFS)
			e.Any(uploadServerPath+"*", echo.WrapHandler(uploadServer))

			chunkServerPath, chunkServer := upload.NewChunkServer(logger, fsys, cacheFS)
			e.Any(chunkServerPath, echo.WrapHandler(chunkServer))

			downloadServerPath, downloadServer := download.NewServer(logger, fsys)
			e.Any(downloadServerPath+"*", echo.WrapHandler(downloadServer))

			// Allow the browser to report telemetry / errors.
			telemetryProxyServerPath, telemetryProxyServer := telemetry.NewProxyServer(logger, telemetryReporter)
			e.Any(telemetryProxyServerPath+"*", echo.WrapHandler(telemetryProxyServer))

			// Catch any shutdown signals.
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigCh

				logger.Info("Shutting down server")

				if err := e.Shutdown(c.Context); err != nil {
					logger.Error("Failed to shutdown server", "error", err)
				}
			}()

			if !headless {
				fmt.Println(banner)
			}

			if err := e.StartH2CServer(c.String("listen"), &http2.Server{}); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("failed to start server: %w", err)
			}

			return nil
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
