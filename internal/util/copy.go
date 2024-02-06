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

package util

import (
	"io"

	"github.com/bucket-sailor/writablefs"
)

func CopyFile(srcFS writablefs.FS, srcPath string, dstFS writablefs.FS, dstPath string) error {
	src, err := srcFS.OpenFile(srcPath, writablefs.FlagReadOnly)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := dstFS.OpenFile(dstPath, writablefs.FlagWriteOnly|writablefs.FlagCreate)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}