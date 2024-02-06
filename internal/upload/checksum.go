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

package upload

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/cespare/xxhash/v2"
)

const (
	algorithmXXH64 = "xxh64"
)

func verifyChecksum(r io.Reader, expected string) error {
	var algorithm string
	if strings.Contains(expected, ":") {
		parts := strings.SplitN(expected, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid checksum format: %s", expected)
		}

		algorithm = parts[0]
	}

	actual, err := checksum(r, algorithm)
	if err != nil {
		return err
	}

	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

func checksum(r io.Reader, algorithm string) (string, error) {
	if algorithm != algorithmXXH64 {
		return "", fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	h := xxhash.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("xxh64:%s", hex.EncodeToString(h.Sum(nil))), nil
}
