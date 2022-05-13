// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: influxdata/iox/ingester/v1/write_info.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// the state
type KafkaPartitionStatus int32

const (
	// Unspecified status, will result in an error.
	KafkaPartitionStatus_KAFKA_PARTITION_STATUS_UNSPECIFIED KafkaPartitionStatus = 0
	// The ingester has not yet processed data in this write
	KafkaPartitionStatus_KAFKA_PARTITION_STATUS_DURABLE KafkaPartitionStatus = 1
	// The ingester has processed the data in this write and it is
	// readable (will be included in a query response)?
	KafkaPartitionStatus_KAFKA_PARTITION_STATUS_READABLE KafkaPartitionStatus = 2
	// The ingester has processed the data in this write and it is both
	// readable and completly persisted to parquet files.
	KafkaPartitionStatus_KAFKA_PARTITION_STATUS_PERSISTED KafkaPartitionStatus = 3
	// The ingester does not have information about this kafka
	// partition
	KafkaPartitionStatus_KAFKA_PARTITION_STATUS_UNKNOWN KafkaPartitionStatus = 4
)

// Enum value maps for KafkaPartitionStatus.
var (
	KafkaPartitionStatus_name = map[int32]string{
		0: "KAFKA_PARTITION_STATUS_UNSPECIFIED",
		1: "KAFKA_PARTITION_STATUS_DURABLE",
		2: "KAFKA_PARTITION_STATUS_READABLE",
		3: "KAFKA_PARTITION_STATUS_PERSISTED",
		4: "KAFKA_PARTITION_STATUS_UNKNOWN",
	}
	KafkaPartitionStatus_value = map[string]int32{
		"KAFKA_PARTITION_STATUS_UNSPECIFIED": 0,
		"KAFKA_PARTITION_STATUS_DURABLE":     1,
		"KAFKA_PARTITION_STATUS_READABLE":    2,
		"KAFKA_PARTITION_STATUS_PERSISTED":   3,
		"KAFKA_PARTITION_STATUS_UNKNOWN":     4,
	}
)

func (x KafkaPartitionStatus) Enum() *KafkaPartitionStatus {
	p := new(KafkaPartitionStatus)
	*p = x
	return p
}

func (x KafkaPartitionStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (KafkaPartitionStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_influxdata_iox_ingester_v1_write_info_proto_enumTypes[0].Descriptor()
}

func (KafkaPartitionStatus) Type() protoreflect.EnumType {
	return &file_influxdata_iox_ingester_v1_write_info_proto_enumTypes[0]
}

func (x KafkaPartitionStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use KafkaPartitionStatus.Descriptor instead.
func (KafkaPartitionStatus) EnumDescriptor() ([]byte, []int) {
	return file_influxdata_iox_ingester_v1_write_info_proto_rawDescGZIP(), []int{0}
}

type GetWriteInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The write token returned from a write that was written to one or
	// more kafka partitions
	WriteToken string `protobuf:"bytes,1,opt,name=write_token,json=writeToken,proto3" json:"write_token,omitempty"`
}

func (x *GetWriteInfoRequest) Reset() {
	*x = GetWriteInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetWriteInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetWriteInfoRequest) ProtoMessage() {}

func (x *GetWriteInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetWriteInfoRequest.ProtoReflect.Descriptor instead.
func (*GetWriteInfoRequest) Descriptor() ([]byte, []int) {
	return file_influxdata_iox_ingester_v1_write_info_proto_rawDescGZIP(), []int{0}
}

func (x *GetWriteInfoRequest) GetWriteToken() string {
	if x != nil {
		return x.WriteToken
	}
	return ""
}

type GetWriteInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Information for all partitions in this write
	KafkaPartitionInfos []*KafkaPartitionInfo `protobuf:"bytes,3,rep,name=kafka_partition_infos,json=kafkaPartitionInfos,proto3" json:"kafka_partition_infos,omitempty"`
}

func (x *GetWriteInfoResponse) Reset() {
	*x = GetWriteInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetWriteInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetWriteInfoResponse) ProtoMessage() {}

func (x *GetWriteInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetWriteInfoResponse.ProtoReflect.Descriptor instead.
func (*GetWriteInfoResponse) Descriptor() ([]byte, []int) {
	return file_influxdata_iox_ingester_v1_write_info_proto_rawDescGZIP(), []int{1}
}

func (x *GetWriteInfoResponse) GetKafkaPartitionInfos() []*KafkaPartitionInfo {
	if x != nil {
		return x.KafkaPartitionInfos
	}
	return nil
}

// Status of a part of a write for in a particular kafka partition
type KafkaPartitionInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique kafka partition id
	KafkaPartitionId int32 `protobuf:"varint,1,opt,name=kafka_partition_id,json=kafkaPartitionId,proto3" json:"kafka_partition_id,omitempty"`
	// the status of the data for this partition
	Status KafkaPartitionStatus `protobuf:"varint,2,opt,name=status,proto3,enum=influxdata.iox.ingester.v1.KafkaPartitionStatus" json:"status,omitempty"`
}

func (x *KafkaPartitionInfo) Reset() {
	*x = KafkaPartitionInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KafkaPartitionInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KafkaPartitionInfo) ProtoMessage() {}

func (x *KafkaPartitionInfo) ProtoReflect() protoreflect.Message {
	mi := &file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KafkaPartitionInfo.ProtoReflect.Descriptor instead.
func (*KafkaPartitionInfo) Descriptor() ([]byte, []int) {
	return file_influxdata_iox_ingester_v1_write_info_proto_rawDescGZIP(), []int{2}
}

func (x *KafkaPartitionInfo) GetKafkaPartitionId() int32 {
	if x != nil {
		return x.KafkaPartitionId
	}
	return 0
}

func (x *KafkaPartitionInfo) GetStatus() KafkaPartitionStatus {
	if x != nil {
		return x.Status
	}
	return KafkaPartitionStatus_KAFKA_PARTITION_STATUS_UNSPECIFIED
}

var File_influxdata_iox_ingester_v1_write_info_proto protoreflect.FileDescriptor

var file_influxdata_iox_ingester_v1_write_info_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x69, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x69, 0x6f, 0x78,
	0x2f, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x77, 0x72, 0x69,
	0x74, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x69,
	0x6e, 0x66, 0x6c, 0x75, 0x78, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x69, 0x6f, 0x78, 0x2e, 0x69, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x36, 0x0a, 0x13, 0x47, 0x65, 0x74,
	0x57, 0x72, 0x69, 0x74, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1f, 0x0a, 0x0b, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x72, 0x69, 0x74, 0x65, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x22, 0x7a, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x65, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x62, 0x0a, 0x15, 0x6b, 0x61, 0x66,
	0x6b, 0x61, 0x5f, 0x70, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x66,
	0x6f, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x69, 0x6e, 0x66, 0x6c, 0x75,
	0x78, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x69, 0x6f, 0x78, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x61, 0x66, 0x6b, 0x61, 0x50, 0x61, 0x72, 0x74, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x13, 0x6b, 0x61, 0x66, 0x6b, 0x61, 0x50,
	0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x22, 0x8c, 0x01,
	0x0a, 0x12, 0x4b, 0x61, 0x66, 0x6b, 0x61, 0x50, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2c, 0x0a, 0x12, 0x6b, 0x61, 0x66, 0x6b, 0x61, 0x5f, 0x70, 0x61,
	0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x10, 0x6b, 0x61, 0x66, 0x6b, 0x61, 0x50, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x49, 0x64, 0x12, 0x48, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x30, 0x2e, 0x69, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x64, 0x61, 0x74, 0x61, 0x2e,
	0x69, 0x6f, 0x78, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x4b, 0x61, 0x66, 0x6b, 0x61, 0x50, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0xd1, 0x01, 0x0a,
	0x14, 0x4b, 0x61, 0x66, 0x6b, 0x61, 0x50, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x26, 0x0a, 0x22, 0x4b, 0x41, 0x46, 0x4b, 0x41, 0x5f, 0x50,
	0x41, 0x52, 0x54, 0x49, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x22, 0x0a,
	0x1e, 0x4b, 0x41, 0x46, 0x4b, 0x41, 0x5f, 0x50, 0x41, 0x52, 0x54, 0x49, 0x54, 0x49, 0x4f, 0x4e,
	0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x44, 0x55, 0x52, 0x41, 0x42, 0x4c, 0x45, 0x10,
	0x01, 0x12, 0x23, 0x0a, 0x1f, 0x4b, 0x41, 0x46, 0x4b, 0x41, 0x5f, 0x50, 0x41, 0x52, 0x54, 0x49,
	0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x41, 0x44,
	0x41, 0x42, 0x4c, 0x45, 0x10, 0x02, 0x12, 0x24, 0x0a, 0x20, 0x4b, 0x41, 0x46, 0x4b, 0x41, 0x5f,
	0x50, 0x41, 0x52, 0x54, 0x49, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53,
	0x5f, 0x50, 0x45, 0x52, 0x53, 0x49, 0x53, 0x54, 0x45, 0x44, 0x10, 0x03, 0x12, 0x22, 0x0a, 0x1e,
	0x4b, 0x41, 0x46, 0x4b, 0x41, 0x5f, 0x50, 0x41, 0x52, 0x54, 0x49, 0x54, 0x49, 0x4f, 0x4e, 0x5f,
	0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x04,
	0x32, 0x85, 0x01, 0x0a, 0x10, 0x57, 0x72, 0x69, 0x74, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x71, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2f, 0x2e, 0x69, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x69, 0x6f, 0x78, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x69, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x64,
	0x61, 0x74, 0x61, 0x2e, 0x69, 0x6f, 0x78, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x65, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x64, 0x61, 0x74,
	0x61, 0x2f, 0x69, 0x6f, 0x78, 0x2f, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2f, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_influxdata_iox_ingester_v1_write_info_proto_rawDescOnce sync.Once
	file_influxdata_iox_ingester_v1_write_info_proto_rawDescData = file_influxdata_iox_ingester_v1_write_info_proto_rawDesc
)

func file_influxdata_iox_ingester_v1_write_info_proto_rawDescGZIP() []byte {
	file_influxdata_iox_ingester_v1_write_info_proto_rawDescOnce.Do(func() {
		file_influxdata_iox_ingester_v1_write_info_proto_rawDescData = protoimpl.X.CompressGZIP(file_influxdata_iox_ingester_v1_write_info_proto_rawDescData)
	})
	return file_influxdata_iox_ingester_v1_write_info_proto_rawDescData
}

var file_influxdata_iox_ingester_v1_write_info_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_influxdata_iox_ingester_v1_write_info_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_influxdata_iox_ingester_v1_write_info_proto_goTypes = []interface{}{
	(KafkaPartitionStatus)(0),    // 0: influxdata.iox.ingester.v1.KafkaPartitionStatus
	(*GetWriteInfoRequest)(nil),  // 1: influxdata.iox.ingester.v1.GetWriteInfoRequest
	(*GetWriteInfoResponse)(nil), // 2: influxdata.iox.ingester.v1.GetWriteInfoResponse
	(*KafkaPartitionInfo)(nil),   // 3: influxdata.iox.ingester.v1.KafkaPartitionInfo
}
var file_influxdata_iox_ingester_v1_write_info_proto_depIdxs = []int32{
	3, // 0: influxdata.iox.ingester.v1.GetWriteInfoResponse.kafka_partition_infos:type_name -> influxdata.iox.ingester.v1.KafkaPartitionInfo
	0, // 1: influxdata.iox.ingester.v1.KafkaPartitionInfo.status:type_name -> influxdata.iox.ingester.v1.KafkaPartitionStatus
	1, // 2: influxdata.iox.ingester.v1.WriteInfoService.GetWriteInfo:input_type -> influxdata.iox.ingester.v1.GetWriteInfoRequest
	2, // 3: influxdata.iox.ingester.v1.WriteInfoService.GetWriteInfo:output_type -> influxdata.iox.ingester.v1.GetWriteInfoResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_influxdata_iox_ingester_v1_write_info_proto_init() }
func file_influxdata_iox_ingester_v1_write_info_proto_init() {
	if File_influxdata_iox_ingester_v1_write_info_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetWriteInfoRequest); i {
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
		file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetWriteInfoResponse); i {
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
		file_influxdata_iox_ingester_v1_write_info_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KafkaPartitionInfo); i {
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
			RawDescriptor: file_influxdata_iox_ingester_v1_write_info_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_influxdata_iox_ingester_v1_write_info_proto_goTypes,
		DependencyIndexes: file_influxdata_iox_ingester_v1_write_info_proto_depIdxs,
		EnumInfos:         file_influxdata_iox_ingester_v1_write_info_proto_enumTypes,
		MessageInfos:      file_influxdata_iox_ingester_v1_write_info_proto_msgTypes,
	}.Build()
	File_influxdata_iox_ingester_v1_write_info_proto = out.File
	file_influxdata_iox_ingester_v1_write_info_proto_rawDesc = nil
	file_influxdata_iox_ingester_v1_write_info_proto_goTypes = nil
	file_influxdata_iox_ingester_v1_write_info_proto_depIdxs = nil
}