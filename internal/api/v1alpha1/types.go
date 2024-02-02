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

package v1alpha1

import (
	"encoding/json"
	"time"
)

type Time time.Time

func (t *Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*t).UTC().Format(time.RFC3339))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = Time(tt)
	return nil
}

type FileInfo struct {
	Name         string `json:"name"`
	IsDir        bool   `json:"isDir"`
	Size         *int64 `json:"size,omitempty"`
	LastModified *Time  `json:"lastModified,omitempty"`
}

type FileInfoWithIndex struct {
	FileInfo
	// The position of the file in the complete directory listing.
	// This is effectively a stable identifier for the file across
	// multiple requests with the same list identifier.
	Index int `json:"index"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ListResponse struct {
	// ID is a unique identifier for the list. It can be used to
	// retrieve the same list of files in a subsequent request.
	// This avoids issues with unstable pagination.
	ID string `json:"id"`
	// Files is the list of files in the directory (limited to the
	// optionally provided start and stop indexes).
	Files []FileInfoWithIndex `json:"files,omitempty"`
}
