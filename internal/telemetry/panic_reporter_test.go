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

package telemetry_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bucket-sailor/bucketeer/internal/gen/telemetry/v1alpha1"
	"github.com/bucket-sailor/bucketeer/internal/telemetry"
	"github.com/bucket-sailor/bucketeer/internal/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPanicReporter(t *testing.T) {
	logger := slogt.New(t)

	e := echo.New()
	e.HideBanner = true

	mockReporter := &mockTelemetryReporter{}

	mockReporter.On("ReportEvent", mock.Anything).Return()

	panicReporter := telemetry.NewPanicReporter(logger, mockReporter)

	recoverConfig := middleware.DefaultRecoverConfig
	// Capture as much as practical
	recoverConfig.StackSize = 8000000
	recoverConfig.LogErrorFunc = panicReporter.OnPanic

	e.Use(middleware.RecoverWithConfig(recoverConfig))

	var foo *int
	e.GET("/panic", func(c echo.Context) error {
		// null pointer dereference.
		*foo += 42

		return nil
	})

	go func() {
		if err := e.Start(":0"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", "error", err)
		}
	}()
	t.Cleanup(func() {
		require.NoError(t, e.Close())
	})

	err := util.WaitForServerReady(e, 10*time.Second)
	require.NoError(t, err)

	_, err = http.Get("http://" + e.Listener.Addr().String() + "/panic")
	require.NoError(t, err)

	mockReporter.AssertCalled(t, "ReportEvent", mock.Anything)

	event := mockReporter.Calls[0].Arguments.Get(0).(*v1alpha1.TelemetryEvent)

	assert.NotEmpty(t, event.StackTrace)
}

type mockTelemetryReporter struct {
	mock.Mock
}

func (r *mockTelemetryReporter) ReportEvent(event *v1alpha1.TelemetryEvent) {
	r.Called(event)
}

func (r *mockTelemetryReporter) Close() error {
	r.Called()
	return nil
}
