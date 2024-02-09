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
	"io"
	"log/slog"
	"strings"

	"github.com/bucket-sailor/bucketeer/internal/gen/telemetry/v1alpha1"
	"github.com/labstack/echo/v4"
	"github.com/maruel/panicparse/v2/stack"
)

// PanicReporter reports crashes to our telemetry server.
type PanicReporter struct {
	logger   *slog.Logger
	reporter Reporter
}

func NewPanicReporter(logger *slog.Logger, reporter Reporter) *PanicReporter {
	return &PanicReporter{
		logger:   logger.WithGroup("panic"),
		reporter: reporter,
	}
}

func (r *PanicReporter) OnPanic(c echo.Context, reportedError error, reportedStack []byte) error {
	r.logger.Error("Panic detected", "error", reportedError)

	snapshot, _, err := stack.ScanSnapshot(strings.NewReader(string(reportedStack)+"\n\n"), io.Discard, stack.DefaultOpts())
	if err != nil {
		r.logger.Error("Stack is probably corrupted", "error", err)

		return echo.NewHTTPError(500, "Internal server error")
	}

	var panickedRoutine *stack.Goroutine
	for _, g := range snapshot.Goroutines {
		if panickedRoutine != nil {
			break
		}

		for _, call := range g.Stack.Calls {
			if call.Func.Name == "panic" {
				panickedRoutine = g
				break
			}
		}
	}

	if panickedRoutine == nil {
		r.logger.Warn("Could not find panicked goroutine")

		return echo.NewHTTPError(500, "Internal server error")
	}

	var stackTrace []*v1alpha1.StackFrame
	for _, call := range panickedRoutine.Stack.Calls {
		stackTrace = append(stackTrace, &v1alpha1.StackFrame{
			File:     call.RemoteSrcPath,
			Function: call.Func.Name,
			Line:     int32(call.Line),
		})
	}

	r.reporter.ReportEvent(&v1alpha1.TelemetryEvent{
		Kind:       v1alpha1.TelemetryEventKind_ERROR,
		Name:       "Error",
		Message:    reportedError.Error(),
		StackTrace: stackTrace,
	})

	return echo.NewHTTPError(500, "Internal server error")
}
