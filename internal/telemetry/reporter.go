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

package telemetry

import (
	"context"
	"io"
	"log/slog"
	"os"

	"connectrpc.com/connect"
	"github.com/bucket-sailor/bucketeer/internal/constants"
	"github.com/bucket-sailor/bucketeer/internal/gen/telemetry/v1alpha1"
	"github.com/bucket-sailor/bucketeer/internal/gen/telemetry/v1alpha1/v1alpha1connect"
	"github.com/bucket-sailor/bucketeer/internal/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// If set to any non-empty value, telemetry reporting will be disabled.
	telemetryOptOutEnvVar = "BUCKETEER_NO_TELEMETRY"
)

type Reporter interface {
	io.Closer
	ReportStart(ctx context.Context, endpointURL string) error
	ReportEvent(event *v1alpha1.TelemetryEvent)
}

type RemoteReporter struct {
	ctx       context.Context
	logger    *slog.Logger
	client    v1alpha1connect.TelemetryClient
	enabled   bool
	processID string
}

func NewRemoteReporter(ctx context.Context, logger *slog.Logger, httpClient connect.HTTPClient, baseURL string) Reporter {
	enabled := os.Getenv(telemetryOptOutEnvVar) == ""

	logger = logger.WithGroup("telemetry")
	if !enabled {
		logger.Info("Telemetry reporting is disabled")
	}

	return &RemoteReporter{
		ctx:       ctx,
		logger:    logger,
		enabled:   enabled,
		client:    v1alpha1connect.NewTelemetryClient(httpClient, baseURL),
		processID: util.GenerateID(16),
	}
}

func (r *RemoteReporter) Close() error {
	r.ReportEvent(&v1alpha1.TelemetryEvent{
		Name: "ApplicationStop",
	})

	return nil
}

func (r *RemoteReporter) ReportStart(ctx context.Context, endpointURL string) error {
	provider, err := StripS3EndpointURL(endpointURL)
	if err != nil {
		return err
	}

	values := sysInfo(ctx)
	values["appVersion"] = constants.Version
	values["provider"] = provider

	r.ReportEvent(&v1alpha1.TelemetryEvent{
		Name:   "ApplicationStart",
		Values: values,
	})

	return nil
}

func (r *RemoteReporter) ReportEvent(event *v1alpha1.TelemetryEvent) {
	event.Timestamp = timestamppb.Now()

	if event.SessionId == "" {
		event.SessionId = r.processID
	}

	var webEvent bool
	for _, tag := range event.Tags {
		if tag == "web" {
			webEvent = true
			break
		}
	}

	if !webEvent {
		event.Tags = append(event.Tags, "backend")
	}

	if r.enabled {
		req := &connect.Request[v1alpha1.TelemetryEvent]{Msg: event}
		if constants.TelemetryToken != "" {
			req.Header().Set(
				"Authorization",
				"Bearer "+constants.TelemetryToken,
			)
		}

		_, err := r.client.Report(r.ctx, req)
		if err != nil {
			// Don't log errors when the user is offline.
			r.logger.Debug("Failed to report event", "err", err)
		}
	}
}
