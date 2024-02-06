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

package filesystem

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/bucket-sailor/bucketeer/internal/gen/filesystem/v1alpha1"
	"github.com/bucket-sailor/bucketeer/internal/gen/filesystem/v1alpha1/v1alpha1connect"
	"github.com/bucket-sailor/writablefs"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	readDirCacheMaxSize = 100
	readDirCacheTTL     = 5 * time.Minute
)

type Server struct {
	http.Handler
	logger *slog.Logger
	fsys   writablefs.FS
	// Cache for directory listings (in the future this should support being stored in Redis etc.).
	readDirCache *expirable.LRU[string, []*v1alpha1.ReadDirResponse_FileInfoWithIndex]
}

func NewServer(logger *slog.Logger, fsys writablefs.FS) (string, http.Handler) {
	s := &Server{
		logger:       logger,
		fsys:         fsys,
		readDirCache: expirable.NewLRU[string, []*v1alpha1.ReadDirResponse_FileInfoWithIndex](readDirCacheMaxSize, nil, readDirCacheTTL),
	}

	var path string
	path, s.Handler = v1alpha1connect.NewFilesystemHandler(s)

	s.Handler = http.StripPrefix("/api", s.Handler)

	return "/api" + path, s
}

func (s *Server) ReadDir(ctx context.Context, req *connect.Request[v1alpha1.ReadDirRequest]) (*connect.Response[v1alpha1.ReadDirResponse], error) {
	populateCache := func(id string) ([]*v1alpha1.ReadDirResponse_FileInfoWithIndex, error) {
		entries, err := s.fsys.ReadDir(req.Msg.Path)
		if err != nil {
			if errors.Is(err, writablefs.ErrNotExist) {
				return nil, fmt.Errorf("unable to list directory: %w", err)
			}

			return nil, err
		}

		files := make([]*v1alpha1.ReadDirResponse_FileInfoWithIndex, len(entries))
		for i, entry := range entries {
			fi, err := toFileInfo(entry)
			if err != nil {
				return nil, err
			}

			files[i] = &v1alpha1.ReadDirResponse_FileInfoWithIndex{
				Index:    int64(i),
				FileInfo: fi,
			}
		}

		s.readDirCache.Add(id, files)

		return files, nil
	}

	var err error
	var files []*v1alpha1.ReadDirResponse_FileInfoWithIndex

	id := req.Msg.Id
	if id == "" {
		id = generateID(8)

		files, err = populateCache(id)
		if err != nil {
			if errors.Is(err, writablefs.ErrNotExist) {
				return nil, connect.NewError(connect.CodeNotFound, err)
			}

			return nil, connect.NewError(connect.CodeInternal, err)
		}
	} else {
		var ok bool
		files, ok = s.readDirCache.Get(id)
		if !ok {
			files, err = populateCache(id)
			if err != nil {
				if errors.Is(err, writablefs.ErrNotExist) {
					return nil, connect.NewError(connect.CodeNotFound, err)
				}

				return nil, connect.NewError(connect.CodeInternal, err)
			}
		}
	}

	// No pagination.
	if req.Msg.StartIndex == 0 && req.Msg.StopIndex == 0 {
		return &connect.Response[v1alpha1.ReadDirResponse]{
			Msg: &v1alpha1.ReadDirResponse{
				Id:    id,
				Files: files,
			},
		}, nil
	}

	// Bounds checking.
	maxIndex := int64(len(files)) - 1
	if len(files) == 0 || req.Msg.StartIndex > maxIndex {
		return &connect.Response[v1alpha1.ReadDirResponse]{
			Msg: &v1alpha1.ReadDirResponse{
				Id: id,
			},
		}, nil
	}

	startIndex := min(max(req.Msg.StartIndex, 0), maxIndex)
	stopIndex := min(req.Msg.StopIndex, maxIndex)

	if startIndex > stopIndex {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("start index must be less than stop index"))
	}

	return &connect.Response[v1alpha1.ReadDirResponse]{
		Msg: &v1alpha1.ReadDirResponse{
			Id:    id,
			Files: files[startIndex : stopIndex+1],
		},
	}, nil
}

func (s *Server) Stat(ctx context.Context, req *connect.Request[wrapperspb.StringValue]) (*connect.Response[v1alpha1.FileInfo], error) {
	fi, err := s.fsys.Stat(req.Msg.Value)
	if err != nil {
		if errors.Is(err, writablefs.ErrNotExist) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &connect.Response[v1alpha1.FileInfo]{
		Msg: &v1alpha1.FileInfo{
			Name:    fi.Name(),
			IsDir:   fi.IsDir(),
			Size:    fi.Size(),
			ModTime: timestamppb.New(fi.ModTime()),
		},
	}, nil
}

func (s *Server) MkdirAll(ctx context.Context, req *connect.Request[wrapperspb.StringValue]) (*connect.Response[emptypb.Empty], error) {
	if err := s.fsys.MkdirAll(req.Msg.Value); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &connect.Response[emptypb.Empty]{
		Msg: &emptypb.Empty{},
	}, nil
}

func (s *Server) RemoveAll(ctx context.Context, req *connect.Request[wrapperspb.StringValue]) (*connect.Response[emptypb.Empty], error) {
	if err := s.fsys.RemoveAll(req.Msg.Value); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &connect.Response[emptypb.Empty]{
		Msg: &emptypb.Empty{},
	}, nil
}

func toFileInfo(entry writablefs.DirEntry) (*v1alpha1.FileInfo, error) {
	resp := &v1alpha1.FileInfo{
		Name:  entry.Name(),
		IsDir: entry.IsDir(),
	}

	if fi, err := entry.Info(); err == nil {
		resp.Size = fi.Size()
		resp.ModTime = timestamppb.New(fi.ModTime())
	}

	return resp, nil
}

func generateID(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	for i := range b {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}

		b[i] = letters[r.Int64()]
	}

	return string(b)
}
