// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Copyright 2024 Damian Peckett <damian@pecke.tt>.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// @generated by protoc-gen-connect-es v1.3.0 with parameter "target=ts,import_extension=none"
// @generated from file filesystem/v1alpha1/filesystem.proto (package bucketeer.filesystem.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { FileInfo, ReadDirRequest, ReadDirResponse } from "./filesystem_pb";
import { Empty, MethodKind, StringValue } from "@bufbuild/protobuf";

/**
 * @generated from service bucketeer.filesystem.v1alpha1.Filesystem
 */
export const Filesystem = {
  typeName: "bucketeer.filesystem.v1alpha1.Filesystem",
  methods: {
    /**
     * ReadDir returns a list of files in a directory.
     *
     * @generated from rpc bucketeer.filesystem.v1alpha1.Filesystem.ReadDir
     */
    readDir: {
      name: "ReadDir",
      I: ReadDirRequest,
      O: ReadDirResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stat returns information about a file or directory.
     *
     * @generated from rpc bucketeer.filesystem.v1alpha1.Filesystem.Stat
     */
    stat: {
      name: "Stat",
      I: StringValue,
      O: FileInfo,
      kind: MethodKind.Unary,
    },
    /**
     * MkdirAll creates a directory and any necessary parents.
     *
     * @generated from rpc bucketeer.filesystem.v1alpha1.Filesystem.MkdirAll
     */
    mkdirAll: {
      name: "MkdirAll",
      I: StringValue,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * RemoveAll removes a directory and any children it contains.
     *
     * @generated from rpc bucketeer.filesystem.v1alpha1.Filesystem.RemoveAll
     */
    removeAll: {
      name: "RemoveAll",
      I: StringValue,
      O: Empty,
      kind: MethodKind.Unary,
    },
  }
} as const;
