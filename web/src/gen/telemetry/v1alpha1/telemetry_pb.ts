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

// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file telemetry/v1alpha1/telemetry.proto (package bucketeer.telemetry.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from enum bucketeer.telemetry.v1alpha1.TelemetryEventKind
 */
export enum TelemetryEventKind {
  /**
   * The event is an informational message.
   *
   * @generated from enum value: INFO = 0;
   */
  INFO = 0,

  /**
   * The event is a warning message.
   *
   * @generated from enum value: WARNING = 1;
   */
  WARNING = 1,

  /**
   * The event is an error message.
   *
   * @generated from enum value: ERROR = 2;
   */
  ERROR = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(TelemetryEventKind)
proto3.util.setEnumType(TelemetryEventKind, "bucketeer.telemetry.v1alpha1.TelemetryEventKind", [
  { no: 0, name: "INFO" },
  { no: 1, name: "WARNING" },
  { no: 2, name: "ERROR" },
]);

/**
 * @generated from message bucketeer.telemetry.v1alpha1.StackFrame
 */
export class StackFrame extends Message<StackFrame> {
  /**
   * The file name where the error occurred.
   *
   * @generated from field: string file = 1;
   */
  file = "";

  /**
   * The name of the method where the error occurred.
   *
   * @generated from field: string function = 2;
   */
  function = "";

  /**
   * The line number in the file where the error occurred.
   *
   * @generated from field: int32 line = 3;
   */
  line = 0;

  /**
   * The column number in the line where the error occurred.
   *
   * @generated from field: int32 column = 4;
   */
  column = 0;

  constructor(data?: PartialMessage<StackFrame>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "bucketeer.telemetry.v1alpha1.StackFrame";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "file", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "function", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "line", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 4, name: "column", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StackFrame {
    return new StackFrame().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StackFrame {
    return new StackFrame().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StackFrame {
    return new StackFrame().fromJsonString(jsonString, options);
  }

  static equals(a: StackFrame | PlainMessage<StackFrame> | undefined, b: StackFrame | PlainMessage<StackFrame> | undefined): boolean {
    return proto3.util.equals(StackFrame, a, b);
  }
}

/**
 * @generated from message bucketeer.telemetry.v1alpha1.TelemetryEvent
 */
export class TelemetryEvent extends Message<TelemetryEvent> {
  /**
   * The session ID associated with the event.
   *
   * @generated from field: string session_id = 1;
   */
  sessionId = "";

  /**
   * Timestamp when the event occurred.
   *
   * @generated from field: google.protobuf.Timestamp timestamp = 2;
   */
  timestamp?: Timestamp;

  /**
   * The kind of event.
   *
   * @generated from field: bucketeer.telemetry.v1alpha1.TelemetryEventKind kind = 3;
   */
  kind = TelemetryEventKind.INFO;

  /**
   * The name of the event.
   *
   * @generated from field: string name = 4;
   */
  name = "";

  /**
   * A message associated with the event.
   *
   * @generated from field: string message = 5;
   */
  message = "";

  /**
   * Any values associated with the event.
   *
   * @generated from field: map<string, string> values = 6;
   */
  values: { [key: string]: string } = {};

  /**
   * If an error, the stack trace associated with the event.
   *
   * @generated from field: repeated bucketeer.telemetry.v1alpha1.StackFrame stack_trace = 7;
   */
  stackTrace: StackFrame[] = [];

  /**
   * A set of tags associated with the event.
   *
   * @generated from field: repeated string tags = 8;
   */
  tags: string[] = [];

  constructor(data?: PartialMessage<TelemetryEvent>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "bucketeer.telemetry.v1alpha1.TelemetryEvent";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "session_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "timestamp", kind: "message", T: Timestamp },
    { no: 3, name: "kind", kind: "enum", T: proto3.getEnumType(TelemetryEventKind) },
    { no: 4, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "values", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
    { no: 7, name: "stack_trace", kind: "message", T: StackFrame, repeated: true },
    { no: 8, name: "tags", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TelemetryEvent {
    return new TelemetryEvent().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TelemetryEvent {
    return new TelemetryEvent().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TelemetryEvent {
    return new TelemetryEvent().fromJsonString(jsonString, options);
  }

  static equals(a: TelemetryEvent | PlainMessage<TelemetryEvent> | undefined, b: TelemetryEvent | PlainMessage<TelemetryEvent> | undefined): boolean {
    return proto3.util.equals(TelemetryEvent, a, b);
  }
}
