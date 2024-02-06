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
// @generated from file upload/v1alpha1/upload.proto (package bucketeer.upload.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CompleteResponse, NewRequest } from "./upload_pb";
import { Empty, MethodKind, StringValue } from "@bufbuild/protobuf";

/**
 * @generated from service bucketeer.upload.v1alpha1.Upload
 */
export const Upload = {
  typeName: "bucketeer.upload.v1alpha1.Upload",
  methods: {
    /**
     * New initiates a new upload and returns a unique identifier for the upload.
     *
     * @generated from rpc bucketeer.upload.v1alpha1.Upload.New
     */
    new: {
      name: "New",
      I: NewRequest,
      O: StringValue,
      kind: MethodKind.Unary,
    },
    /**
     * Abort aborts an upload and cleans up any resources associated with it.
     *
     * @generated from rpc bucketeer.upload.v1alpha1.Upload.Abort
     */
    abort: {
      name: "Abort",
      I: StringValue,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * Complete begins the process of completing an upload, data isn't guaranteed
     * to be flushed to disk until PollForCompletion() returns a status of
     * COMPLETED. We split this into two calls to allow for the possibility of a
     * long-running completion process (eg. transferring to remote storage).
     *
     * @generated from rpc bucketeer.upload.v1alpha1.Upload.Complete
     */
    complete: {
      name: "Complete",
      I: StringValue,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * PollForCompletion polls for the completion of an upload (eg. has it been
     * fully flushed to disk?)
     *
     * @generated from rpc bucketeer.upload.v1alpha1.Upload.PollForCompletion
     */
    pollForCompletion: {
      name: "PollForCompletion",
      I: StringValue,
      O: CompleteResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

