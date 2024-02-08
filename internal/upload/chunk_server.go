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
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/bucket-sailor/bucketeer/internal/util/contentrange"
	"github.com/bucket-sailor/rangelock"
	"github.com/bucket-sailor/writablefs"
	"github.com/google/uuid"
)

type ChunkServer struct {
	http.Handler
	logger     *slog.Logger
	fsys       writablefs.FS
	cacheFS    writablefs.FS
	rangeLocks sync.Map
}

func NewChunkServer(logger *slog.Logger, fsys, cacheFS writablefs.FS) (string, http.Handler) {
	s := &ChunkServer{
		logger:  logger.WithGroup("upload"),
		fsys:    fsys,
		cacheFS: cacheFS,
	}

	mux := http.NewServeMux()
	s.Handler = mux

	mux.HandleFunc("/files/upload", s.handleUpload)

	return "/files/upload", s
}

func (s *ChunkServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Error reading multipart request", http.StatusInternalServerError)
		return
	}

	for {
		part, err := multipartReader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			http.Error(w, "Error reading next part", http.StatusBadRequest)
			return
		}

		if part.FormName() == "file" {
			// Fallback to the global content range header if the part doesn't have one.
			// Setting part headers from JS is a bit of a pain, so this is a workaround
			// for cases with a single part.
			if part.Header.Get("Content-Range") != "" {
				if err := s.processChunk(r.Context(), part, part.Header.Get("Content-Range")); err != nil {
					http.Error(w, "Error processing chunk: "+err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				if err := s.processChunk(r.Context(), part, r.Header.Get("Content-Range")); err != nil {
					http.Error(w, "Error processing chunk: "+err.Error(), http.StatusInternalServerError)
					return
				}

				// Can only process one part without per-part content range headers.
				break
			}
		}

		if err := part.Close(); err != nil {
			http.Error(w, "Error closing part", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *ChunkServer) processChunk(ctx context.Context, part *multipart.Part, contentRangeHeader string) error {
	uploadID := part.FileName()

	if _, err := uuid.Parse(uploadID); err != nil {
		return fmt.Errorf("invalid upload id: %w", err)
	}

	cachePath := filepath.Join(cacheDir, uploadID)

	f, err := s.cacheFS.OpenFile(cachePath, writablefs.FlagReadWrite|writablefs.FlagCreate)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	rng, err := contentrange.Parse(contentRangeHeader)
	if err != nil {
		return fmt.Errorf("error parsing content range: %w", err)
	}

	s.logger.Debug("Upload", "id", uploadID, "start", rng.Start, "end", rng.End)

	lock, _ := s.rangeLocks.LoadOrStore(uploadID, rangelock.New())

	id, err := lock.(*rangelock.RangeLock).Lock(ctx, rng.Start, rng.End)
	if err != nil {
		return fmt.Errorf("error acquiring lock: %w", err)
	}
	defer lock.(*rangelock.RangeLock).Unlock(id)

	_, err = io.Copy(io.NewOffsetWriter(f, rng.Start), part)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
