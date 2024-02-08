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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bucket-sailor/writablefs"
)

type Server struct {
	http.Handler
	logger *slog.Logger
	fsys   writablefs.FS
}

func NewServer(logger *slog.Logger, fsys writablefs.FS) (string, http.Handler) {
	s := &Server{
		logger: logger.WithGroup("download"),
		fsys:   fsys,
	}

	mux := http.NewServeMux()
	s.Handler = mux

	mux.HandleFunc("/files/", s.handleDownload)

	return "/files/", s
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/files/download/")

	fi, err := s.fsys.Stat(path)
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	if fi.IsDir() {
		s.handleDownloadDirectory(w, r, path, fi)
		return
	}

	s.logger.Debug("Download", "path", path)

	f, err := s.fsys.OpenFile(path, writablefs.FlagReadOnly)
	if err != nil {
		if errors.Is(err, writablefs.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Force download when viewing in browser.
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fi.Name()))

	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

func (s *Server) handleDownloadDirectory(w http.ResponseWriter, r *http.Request, path string, fi writablefs.FileInfo) {
	s.logger.Debug("Download directory", "path", path)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", fi.Name()))
	w.Header().Set("Content-Type", "application/zip")

	archiveFS, ok := s.fsys.(writablefs.ArchiveFS)
	if !ok {
		http.Error(w, "Archive not supported", http.StatusInternalServerError)
		return
	}

	tr, err := archiveFS.Archive(path)
	if err != nil {
		http.Error(w, "Error archiving directory", http.StatusInternalServerError)
		return
	}
	defer tr.Close()

	dirName := filepath.Base(path)
	if err := tarToZip(w, tr, dirName); err != nil {
		http.Error(w, "Error creating zip", http.StatusInternalServerError)
	}
}
