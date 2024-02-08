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

package download

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"io"
	"path/filepath"
)

func tarToZip(w io.Writer, r io.Reader, prefix string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		name := header.Name
		if prefix != "" {
			name = filepath.Join(prefix, name)
		}

		f, err := zw.CreateHeader(&zip.FileHeader{
			Name:               name,
			Method:             zip.Deflate,
			Modified:           header.ModTime,
			UncompressedSize64: uint64(header.Size),
		})
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, tr); err != nil {
			return err
		}
	}

	return zw.Close()
}
