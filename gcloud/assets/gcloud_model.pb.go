// Gcloud-specific extensions for the asset model

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: gcloud_model.proto

package assets

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type GCloudVM struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MachineType string `protobuf:"bytes,1,opt,name=machine_type,json=machineType,proto3" json:"machine_type,omitempty"`
	Region      string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty"`
	Zone        string `protobuf:"bytes,3,opt,name=zone,proto3" json:"zone,omitempty"`
	// OnDemand, Preemptible, Commit1Yr, Commit3Yr
	Scheduling  string `protobuf:"bytes,4,opt,name=scheduling,proto3" json:"scheduling,omitempty"`
	Sharing     string `protobuf:"bytes,5,opt,name=sharing,proto3" json:"sharing,omitempty"` // SoleTenancy, SharedCpu. Default is nothing.
	NetworkTier string `protobuf:"bytes,6,opt,name=network_tier,json=networkTier,proto3" json:"network_tier,omitempty"`
}

func (x *GCloudVM) Reset() {
	*x = GCloudVM{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gcloud_model_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GCloudVM) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GCloudVM) ProtoMessage() {}

func (x *GCloudVM) ProtoReflect() protoreflect.Message {
	mi := &file_gcloud_model_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GCloudVM.ProtoReflect.Descriptor instead.
func (*GCloudVM) Descriptor() ([]byte, []int) {
	return file_gcloud_model_proto_rawDescGZIP(), []int{0}
}

func (x *GCloudVM) GetMachineType() string {
	if x != nil {
		return x.MachineType
	}
	return ""
}

func (x *GCloudVM) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *GCloudVM) GetZone() string {
	if x != nil {
		return x.Zone
	}
	return ""
}

func (x *GCloudVM) GetScheduling() string {
	if x != nil {
		return x.Scheduling
	}
	return ""
}

func (x *GCloudVM) GetSharing() string {
	if x != nil {
		return x.Sharing
	}
	return ""
}

func (x *GCloudVM) GetNetworkTier() string {
	if x != nil {
		return x.NetworkTier
	}
	return ""
}

type GCloudDisk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DiskType   string `protobuf:"bytes,1,opt,name=disk_type,json=diskType,proto3" json:"disk_type,omitempty"`
	IsRegional bool   `protobuf:"varint,2,opt,name=is_regional,json=isRegional,proto3" json:"is_regional,omitempty"`
	Region     string `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty"`
	Zone       string `protobuf:"bytes,4,opt,name=zone,proto3" json:"zone,omitempty"` // only for zonal disks
}

func (x *GCloudDisk) Reset() {
	*x = GCloudDisk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gcloud_model_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GCloudDisk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GCloudDisk) ProtoMessage() {}

func (x *GCloudDisk) ProtoReflect() protoreflect.Message {
	mi := &file_gcloud_model_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GCloudDisk.ProtoReflect.Descriptor instead.
func (*GCloudDisk) Descriptor() ([]byte, []int) {
	return file_gcloud_model_proto_rawDescGZIP(), []int{1}
}

func (x *GCloudDisk) GetDiskType() string {
	if x != nil {
		return x.DiskType
	}
	return ""
}

func (x *GCloudDisk) GetIsRegional() bool {
	if x != nil {
		return x.IsRegional
	}
	return false
}

func (x *GCloudDisk) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *GCloudDisk) GetZone() string {
	if x != nil {
		return x.Zone
	}
	return ""
}

type GCloudSubnetwork struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tier   string `protobuf:"bytes,1,opt,name=tier,proto3" json:"tier,omitempty"`
	Region string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty"`
}

func (x *GCloudSubnetwork) Reset() {
	*x = GCloudSubnetwork{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gcloud_model_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GCloudSubnetwork) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GCloudSubnetwork) ProtoMessage() {}

func (x *GCloudSubnetwork) ProtoReflect() protoreflect.Message {
	mi := &file_gcloud_model_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GCloudSubnetwork.ProtoReflect.Descriptor instead.
func (*GCloudSubnetwork) Descriptor() ([]byte, []int) {
	return file_gcloud_model_proto_rawDescGZIP(), []int{2}
}

func (x *GCloudSubnetwork) GetTier() string {
	if x != nil {
		return x.Tier
	}
	return ""
}

func (x *GCloudSubnetwork) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

var File_gcloud_model_proto protoreflect.FileDescriptor

var file_gcloud_model_proto_rawDesc = []byte{
	0x0a, 0x12, 0x67, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x22, 0xb6, 0x01, 0x0a, 0x08,
	0x47, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x56, 0x4d, 0x12, 0x21, 0x0a, 0x0c, 0x6d, 0x61, 0x63, 0x68,
	0x69, 0x6e, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72,
	0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67,
	0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x7a, 0x6f, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x7a, 0x6f, 0x6e, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x69, 0x6e, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x68, 0x61, 0x72, 0x69,
	0x6e, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x68, 0x61, 0x72, 0x69, 0x6e,
	0x67, 0x12, 0x21, 0x0a, 0x0c, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x74, 0x69, 0x65,
	0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x54, 0x69, 0x65, 0x72, 0x22, 0x76, 0x0a, 0x0a, 0x47, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x44, 0x69,
	0x73, 0x6b, 0x12, 0x1b, 0x0a, 0x09, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x69, 0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x1f, 0x0a, 0x0b, 0x69, 0x73, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x69, 0x73, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x61, 0x6c,
	0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x7a, 0x6f, 0x6e, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x7a, 0x6f, 0x6e, 0x65, 0x22, 0x3e, 0x0a, 0x10,
	0x47, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x53, 0x75, 0x62, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x69, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x42, 0x0a, 0x5a, 0x08,
	0x2e, 0x3b, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gcloud_model_proto_rawDescOnce sync.Once
	file_gcloud_model_proto_rawDescData = file_gcloud_model_proto_rawDesc
)

func file_gcloud_model_proto_rawDescGZIP() []byte {
	file_gcloud_model_proto_rawDescOnce.Do(func() {
		file_gcloud_model_proto_rawDescData = protoimpl.X.CompressGZIP(file_gcloud_model_proto_rawDescData)
	})
	return file_gcloud_model_proto_rawDescData
}

var file_gcloud_model_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_gcloud_model_proto_goTypes = []interface{}{
	(*GCloudVM)(nil),         // 0: model.GCloudVM
	(*GCloudDisk)(nil),       // 1: model.GCloudDisk
	(*GCloudSubnetwork)(nil), // 2: model.GCloudSubnetwork
}
var file_gcloud_model_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_gcloud_model_proto_init() }
func file_gcloud_model_proto_init() {
	if File_gcloud_model_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gcloud_model_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GCloudVM); i {
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
		file_gcloud_model_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GCloudDisk); i {
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
		file_gcloud_model_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GCloudSubnetwork); i {
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
			RawDescriptor: file_gcloud_model_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_gcloud_model_proto_goTypes,
		DependencyIndexes: file_gcloud_model_proto_depIdxs,
		MessageInfos:      file_gcloud_model_proto_msgTypes,
	}.Build()
	File_gcloud_model_proto = out.File
	file_gcloud_model_proto_rawDesc = nil
	file_gcloud_model_proto_goTypes = nil
	file_gcloud_model_proto_depIdxs = nil
}
