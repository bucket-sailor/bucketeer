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
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/bucket-sailor/bucketeer/internal/constants"
	"github.com/bucket-sailor/bucketeer/internal/gen/telemetry/v1alpha1"
	"github.com/bucket-sailor/bucketeer/internal/gen/telemetry/v1alpha1/v1alpha1connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProxyServer proxies telemetry events from the browser.
type ProxyServer struct {
	http.Handler
	logger   *slog.Logger
	reporter Reporter
}

func NewProxyServer(logger *slog.Logger, reporter Reporter) (string, http.Handler) {
	s := &ProxyServer{
		logger:   logger.WithGroup("telemetry"),
		reporter: reporter,
	}

	var path string
	path, s.Handler = v1alpha1connect.NewTelemetryHandler(s)

	s.Handler = http.StripPrefix("/api", s.Handler)

	return "/api" + path, s
}

func (s *ProxyServer) Report(_ context.Context, req *connect.Request[v1alpha1.TelemetryEvent]) (*connect.Response[emptypb.Empty], error) {
	event := req.Msg

	if event.Name == "SessionStart" {
		if event.Values == nil {
			req.Msg.Values = make(map[string]string)
		}

		event.Values["appVersion"] = constants.Version
	}

	s.logger.Debug("Proxying telemetry event", "event", event.Name)

	s.reporter.ReportEvent(event)

	return &connect.Response[emptypb.Empty]{}, nil
}
