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

package contentrange

import (
	"fmt"
	"strconv"
	"strings"
)

type ContentRange struct {
	Start, End, Total int64
}

// Parse parses a Content-Range header string as per RFC 7233.
// It returns the parsed ContentRange or an error if the header is invalid.
func Parse(s string) (*ContentRange, error) {
	if s == "" {
		return nil, fmt.Errorf("content-range header is empty")
	}

	const prefix = "bytes "
	if !strings.HasPrefix(s, prefix) {
		return nil, fmt.Errorf("invalid content-range header")
	}

	s = strings.TrimPrefix(s, prefix)
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid content-range format")
	}

	rangePart, totalStr := parts[0], parts[1]
	startEnd := strings.Split(rangePart, "-")
	if len(startEnd) != 2 {
		return nil, fmt.Errorf("invalid range format")
	}

	startStr, endStr := startEnd[0], startEnd[1]
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start value")
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid end value")
	}

	var total int64
	if totalStr == "*" {
		total = -1 // Indicate unknown total size
	} else {
		total, err = strconv.ParseInt(totalStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid total size")
		}
	}

	if start > end {
		return nil, fmt.Errorf("start cannot be greater than end")
	}

	return &ContentRange{
		Start: start,
		End:   end,
		Total: total,
	}, nil
}
