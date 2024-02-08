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
	"fmt"
	"runtime"

	"github.com/docker/go-units"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func sysInfo(ctx context.Context) map[string]string {
	info := map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpu":  fmt.Sprintf("%d", runtime.NumCPU()),
	}

	if vmStat, err := mem.VirtualMemory(); err == nil {
		info["memory"] = units.HumanSize(float64(vmStat.Total))
	}

	if cpuInfo, err := cpu.Info(); err == nil {
		if len(cpuInfo) > 0 {
			info["cpuModel"] = cpuInfo[0].ModelName
		}
	}

	return info
}
