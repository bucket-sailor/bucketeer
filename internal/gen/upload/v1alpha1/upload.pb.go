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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.21.12
// source: upload/v1alpha1/upload.proto

package v1alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// CompletionStatus is the status of an upload.
type CompletionStatus int32

const (
	// The completion of the upload is still pending.
	CompletionStatus_PENDING CompletionStatus = 0
	// The upload has been completed is ready for downloads.
	CompletionStatus_COMPLETED CompletionStatus = 1
	// Completion of the upload failed.
	CompletionStatus_FAILED CompletionStatus = 2
)

// Enum value maps for CompletionStatus.
var (
	CompletionStatus_name = map[int32]string{
		0: "PENDING",
		1: "COMPLETED",
		2: "FAILED",
	}
	CompletionStatus_value = map[string]int32{
		"PENDING":   0,
		"COMPLETED": 1,
		"FAILED":    2,
	}
)

func (x CompletionStatus) Enum() *CompletionStatus {
	p := new(CompletionStatus)
	*p = x
	return p
}

func (x CompletionStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CompletionStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_upload_v1alpha1_upload_proto_enumTypes[0].Descriptor()
}

func (CompletionStatus) Type() protoreflect.EnumType {
	return &file_upload_v1alpha1_upload_proto_enumTypes[0]
}

func (x CompletionStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CompletionStatus.Descriptor instead.
func (CompletionStatus) EnumDescriptor() ([]byte, []int) {
	return file_upload_v1alpha1_upload_proto_rawDescGZIP(), []int{0}
}

type NewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The destination path of the uploaded file.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// The total size of the uploaded file.
	Size int64 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	// The expected checksum of the uploaded file in the format "algorithm:hex".
	Checksum string `protobuf:"bytes,3,opt,name=checksum,proto3" json:"checksum,omitempty"`
}

func (x *NewRequest) Reset() {
	*x = NewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_upload_v1alpha1_upload_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewRequest) ProtoMessage() {}

func (x *NewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_upload_v1alpha1_upload_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewRequest.ProtoReflect.Descriptor instead.
func (*NewRequest) Descriptor() ([]byte, []int) {
	return file_upload_v1alpha1_upload_proto_rawDescGZIP(), []int{0}
}

func (x *NewRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *NewRequest) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *NewRequest) GetChecksum() string {
	if x != nil {
		return x.Checksum
	}
	return ""
}

type CompleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The status of the upload.
	Status CompletionStatus `protobuf:"varint,1,opt,name=status,proto3,enum=bucketeer.upload.v1alpha1.CompletionStatus" json:"status,omitempty"`
	// The error message if the upload failed.
	Error string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *CompleteResponse) Reset() {
	*x = CompleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_upload_v1alpha1_upload_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CompleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CompleteResponse) ProtoMessage() {}

func (x *CompleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_upload_v1alpha1_upload_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CompleteResponse.ProtoReflect.Descriptor instead.
func (*CompleteResponse) Descriptor() ([]byte, []int) {
	return file_upload_v1alpha1_upload_proto_rawDescGZIP(), []int{1}
}

func (x *CompleteResponse) GetStatus() CompletionStatus {
	if x != nil {
		return x.Status
	}
	return CompletionStatus_PENDING
}

func (x *CompleteResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_upload_v1alpha1_upload_proto protoreflect.FileDescriptor

var file_upload_v1alpha1_upload_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19,
	0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x65, 0x65, 0x72, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x50, 0x0a, 0x0a, 0x4e, 0x65, 0x77, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x75, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x75, 0x6d, 0x22, 0x6d, 0x0a, 0x10, 0x43, 0x6f, 0x6d, 0x70,
	0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x43, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x62,
	0x75, 0x63, 0x6b, 0x65, 0x74, 0x65, 0x65, 0x72, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2a, 0x3a, 0x0a, 0x10, 0x43, 0x6f, 0x6d, 0x70, 0x6c,
	0x65, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x50,
	0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4d, 0x50,
	0x4c, 0x45, 0x54, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45,
	0x44, 0x10, 0x02, 0x32, 0xb5, 0x02, 0x0a, 0x06, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x4a,
	0x0a, 0x03, 0x4e, 0x65, 0x77, 0x12, 0x25, 0x2e, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x65, 0x65,
	0x72, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x4e, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x3d, 0x0a, 0x05, 0x41, 0x62,
	0x6f, 0x72, 0x74, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x40, 0x0a, 0x08, 0x43, 0x6f, 0x6d,
	0x70, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x5e, 0x0a, 0x11, 0x50,
	0x6f, 0x6c, 0x6c, 0x46, 0x6f, 0x72, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x2b,
	0x2e, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x65, 0x65, 0x72, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c,
	0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x41, 0x5a, 0x3f, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74,
	0x2d, 0x73, 0x61, 0x69, 0x6c, 0x6f, 0x72, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x65, 0x65,
	0x72, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x75,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_upload_v1alpha1_upload_proto_rawDescOnce sync.Once
	file_upload_v1alpha1_upload_proto_rawDescData = file_upload_v1alpha1_upload_proto_rawDesc
)

func file_upload_v1alpha1_upload_proto_rawDescGZIP() []byte {
	file_upload_v1alpha1_upload_proto_rawDescOnce.Do(func() {
		file_upload_v1alpha1_upload_proto_rawDescData = protoimpl.X.CompressGZIP(file_upload_v1alpha1_upload_proto_rawDescData)
	})
	return file_upload_v1alpha1_upload_proto_rawDescData
}

var file_upload_v1alpha1_upload_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_upload_v1alpha1_upload_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_upload_v1alpha1_upload_proto_goTypes = []interface{}{
	(CompletionStatus)(0),          // 0: bucketeer.upload.v1alpha1.CompletionStatus
	(*NewRequest)(nil),             // 1: bucketeer.upload.v1alpha1.NewRequest
	(*CompleteResponse)(nil),       // 2: bucketeer.upload.v1alpha1.CompleteResponse
	(*wrapperspb.StringValue)(nil), // 3: google.protobuf.StringValue
	(*emptypb.Empty)(nil),          // 4: google.protobuf.Empty
}
var file_upload_v1alpha1_upload_proto_depIdxs = []int32{
	0, // 0: bucketeer.upload.v1alpha1.CompleteResponse.status:type_name -> bucketeer.upload.v1alpha1.CompletionStatus
	1, // 1: bucketeer.upload.v1alpha1.Upload.New:input_type -> bucketeer.upload.v1alpha1.NewRequest
	3, // 2: bucketeer.upload.v1alpha1.Upload.Abort:input_type -> google.protobuf.StringValue
	3, // 3: bucketeer.upload.v1alpha1.Upload.Complete:input_type -> google.protobuf.StringValue
	3, // 4: bucketeer.upload.v1alpha1.Upload.PollForCompletion:input_type -> google.protobuf.StringValue
	3, // 5: bucketeer.upload.v1alpha1.Upload.New:output_type -> google.protobuf.StringValue
	4, // 6: bucketeer.upload.v1alpha1.Upload.Abort:output_type -> google.protobuf.Empty
	4, // 7: bucketeer.upload.v1alpha1.Upload.Complete:output_type -> google.protobuf.Empty
	2, // 8: bucketeer.upload.v1alpha1.Upload.PollForCompletion:output_type -> bucketeer.upload.v1alpha1.CompleteResponse
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_upload_v1alpha1_upload_proto_init() }
func file_upload_v1alpha1_upload_proto_init() {
	if File_upload_v1alpha1_upload_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_upload_v1alpha1_upload_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_upload_v1alpha1_upload_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CompleteResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_upload_v1alpha1_upload_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_upload_v1alpha1_upload_proto_goTypes,
		DependencyIndexes: file_upload_v1alpha1_upload_proto_depIdxs,
		EnumInfos:         file_upload_v1alpha1_upload_proto_enumTypes,
		MessageInfos:      file_upload_v1alpha1_upload_proto_msgTypes,
	}.Build()
	File_upload_v1alpha1_upload_proto = out.File
	file_upload_v1alpha1_upload_proto_rawDesc = nil
	file_upload_v1alpha1_upload_proto_goTypes = nil
	file_upload_v1alpha1_upload_proto_depIdxs = nil
}