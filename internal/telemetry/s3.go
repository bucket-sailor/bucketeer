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
	"fmt"
	"net/url"
	"strings"
)

// StripS3EndpointURL removes user identifiable information from an S3 endpoint URL.
func StripS3EndpointURL(endpointURL string) (string, error) {
	u, err := url.Parse(endpointURL)
	if err != nil {
		return "", err
	}

	hostParts := strings.Split(u.Host, ".")
	if len(hostParts) < 2 {
		return "", fmt.Errorf("invalid domain structure")
	}

	return strings.Join(hostParts[len(hostParts)-2:], "."), nil
}
