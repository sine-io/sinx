// protoc -I api/proto/ --go_out=types --go_opt=paths=source_relative --go-grpc_out=types --go-grpc_opt=paths=source_relative executor.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.1
// source: executor.proto

package types

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ExecuteRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	JobName       string                 `protobuf:"bytes,1,opt,name=job_name,json=jobName,proto3" json:"job_name,omitempty"`
	Config        map[string]string      `protobuf:"bytes,2,rep,name=config,proto3" json:"config,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	StatusServer  uint32                 `protobuf:"varint,3,opt,name=status_server,json=statusServer,proto3" json:"status_server,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExecuteRequest) Reset() {
	*x = ExecuteRequest{}
	mi := &file_executor_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecuteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteRequest) ProtoMessage() {}

func (x *ExecuteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteRequest.ProtoReflect.Descriptor instead.
func (*ExecuteRequest) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{0}
}

func (x *ExecuteRequest) GetJobName() string {
	if x != nil {
		return x.JobName
	}
	return ""
}

func (x *ExecuteRequest) GetConfig() map[string]string {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ExecuteRequest) GetStatusServer() uint32 {
	if x != nil {
		return x.StatusServer
	}
	return 0
}

type ExecuteResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Output        []byte                 `protobuf:"bytes,1,opt,name=output,proto3" json:"output,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExecuteResponse) Reset() {
	*x = ExecuteResponse{}
	mi := &file_executor_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecuteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteResponse) ProtoMessage() {}

func (x *ExecuteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteResponse.ProtoReflect.Descriptor instead.
func (*ExecuteResponse) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{1}
}

func (x *ExecuteResponse) GetOutput() []byte {
	if x != nil {
		return x.Output
	}
	return nil
}

func (x *ExecuteResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type StatusUpdateRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Output        []byte                 `protobuf:"bytes,2,opt,name=output,proto3" json:"output,omitempty"`
	Error         bool                   `protobuf:"varint,3,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StatusUpdateRequest) Reset() {
	*x = StatusUpdateRequest{}
	mi := &file_executor_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusUpdateRequest) ProtoMessage() {}

func (x *StatusUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusUpdateRequest.ProtoReflect.Descriptor instead.
func (*StatusUpdateRequest) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{2}
}

func (x *StatusUpdateRequest) GetOutput() []byte {
	if x != nil {
		return x.Output
	}
	return nil
}

func (x *StatusUpdateRequest) GetError() bool {
	if x != nil {
		return x.Error
	}
	return false
}

type StatusUpdateResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	R             int64                  `protobuf:"varint,1,opt,name=r,proto3" json:"r,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StatusUpdateResponse) Reset() {
	*x = StatusUpdateResponse{}
	mi := &file_executor_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusUpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusUpdateResponse) ProtoMessage() {}

func (x *StatusUpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusUpdateResponse.ProtoReflect.Descriptor instead.
func (*StatusUpdateResponse) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{3}
}

func (x *StatusUpdateResponse) GetR() int64 {
	if x != nil {
		return x.R
	}
	return 0
}

var File_executor_proto protoreflect.FileDescriptor

const file_executor_proto_rawDesc = "" +
	"\n" +
	"\x0eexecutor.proto\x12\x05types\"\xc6\x01\n" +
	"\x0eExecuteRequest\x12\x19\n" +
	"\bjob_name\x18\x01 \x01(\tR\ajobName\x129\n" +
	"\x06config\x18\x02 \x03(\v2!.types.ExecuteRequest.ConfigEntryR\x06config\x12#\n" +
	"\rstatus_server\x18\x03 \x01(\rR\fstatusServer\x1a9\n" +
	"\vConfigEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"?\n" +
	"\x0fExecuteResponse\x12\x16\n" +
	"\x06output\x18\x01 \x01(\fR\x06output\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error\"C\n" +
	"\x13StatusUpdateRequest\x12\x16\n" +
	"\x06output\x18\x02 \x01(\fR\x06output\x12\x14\n" +
	"\x05error\x18\x03 \x01(\bR\x05error\"$\n" +
	"\x14StatusUpdateResponse\x12\f\n" +
	"\x01r\x18\x01 \x01(\x03R\x01r2D\n" +
	"\bExecutor\x128\n" +
	"\aExecute\x12\x15.types.ExecuteRequest\x1a\x16.types.ExecuteResponse2Q\n" +
	"\fStatusHelper\x12A\n" +
	"\x06Update\x12\x1a.types.StatusUpdateRequest\x1a\x1b.types.StatusUpdateResponseB&Z$github.com/sine-io/sinx/plugin/typesb\x06proto3"

var (
	file_executor_proto_rawDescOnce sync.Once
	file_executor_proto_rawDescData []byte
)

func file_executor_proto_rawDescGZIP() []byte {
	file_executor_proto_rawDescOnce.Do(func() {
		file_executor_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_executor_proto_rawDesc), len(file_executor_proto_rawDesc)))
	})
	return file_executor_proto_rawDescData
}

var file_executor_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_executor_proto_goTypes = []any{
	(*ExecuteRequest)(nil),       // 0: types.ExecuteRequest
	(*ExecuteResponse)(nil),      // 1: types.ExecuteResponse
	(*StatusUpdateRequest)(nil),  // 2: types.StatusUpdateRequest
	(*StatusUpdateResponse)(nil), // 3: types.StatusUpdateResponse
	nil,                          // 4: types.ExecuteRequest.ConfigEntry
}
var file_executor_proto_depIdxs = []int32{
	4, // 0: types.ExecuteRequest.config:type_name -> types.ExecuteRequest.ConfigEntry
	0, // 1: types.Executor.Execute:input_type -> types.ExecuteRequest
	2, // 2: types.StatusHelper.Update:input_type -> types.StatusUpdateRequest
	1, // 3: types.Executor.Execute:output_type -> types.ExecuteResponse
	3, // 4: types.StatusHelper.Update:output_type -> types.StatusUpdateResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_executor_proto_init() }
func file_executor_proto_init() {
	if File_executor_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_executor_proto_rawDesc), len(file_executor_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_executor_proto_goTypes,
		DependencyIndexes: file_executor_proto_depIdxs,
		MessageInfos:      file_executor_proto_msgTypes,
	}.Build()
	File_executor_proto = out.File
	file_executor_proto_goTypes = nil
	file_executor_proto_depIdxs = nil
}
