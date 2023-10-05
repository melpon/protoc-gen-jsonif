// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.24.4
// source: extensions.proto

package generated

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_extensions_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         5012,
		Name:          "jsonif_message_optimistic",
		Tag:           "varint,5012,opt,name=jsonif_message_optimistic",
		Filename:      "extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         5013,
		Name:          "jsonif_message_discard_if_default",
		Tag:           "varint,5013,opt,name=jsonif_message_discard_if_default",
		Filename:      "extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         5014,
		Name:          "jsonif_no_serializer",
		Tag:           "varint,5014,opt,name=jsonif_no_serializer",
		Filename:      "extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         5015,
		Name:          "jsonif_no_deserializer",
		Tag:           "varint,5015,opt,name=jsonif_no_deserializer",
		Filename:      "extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         5012,
		Name:          "jsonif_optimistic",
		Tag:           "varint,5012,opt,name=jsonif_optimistic",
		Filename:      "extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         5013,
		Name:          "jsonif_discard_if_default",
		Tag:           "varint,5013,opt,name=jsonif_discard_if_default",
		Filename:      "extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         5014,
		Name:          "jsonif_name",
		Tag:           "bytes,5014,opt,name=jsonif_name",
		Filename:      "extensions.proto",
	},
}

// Extension fields to descriptorpb.MessageOptions.
var (
	// JSON から読み込む際に、存在しないフィールドがあってもエラーにしない（各型のデフォルト値になる）
	//
	// optional bool jsonif_message_optimistic = 5012;
	E_JsonifMessageOptimistic = &file_extensions_proto_extTypes[0]
	// JSON に変換する際に、デフォルト値のままだったらフィールドを出力しない
	//
	// optional bool jsonif_message_discard_if_default = 5013;
	E_JsonifMessageDiscardIfDefault = &file_extensions_proto_extTypes[1]
	// JSON へのシリアライズ処理を出力しない
	//
	// optional bool jsonif_no_serializer = 5014;
	E_JsonifNoSerializer = &file_extensions_proto_extTypes[2]
	// JSON からのデシリアライズ処理を出力しない
	//
	// optional bool jsonif_no_deserializer = 5015;
	E_JsonifNoDeserializer = &file_extensions_proto_extTypes[3]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional bool jsonif_optimistic = 5012;
	E_JsonifOptimistic = &file_extensions_proto_extTypes[4]
	// optional bool jsonif_discard_if_default = 5013;
	E_JsonifDiscardIfDefault = &file_extensions_proto_extTypes[5]
	// optional string jsonif_name = 5014;
	E_JsonifName = &file_extensions_proto_extTypes[6]
)

var File_extensions_proto protoreflect.FileDescriptor

var file_extensions_proto_rawDesc = []byte{
	0x0a, 0x10, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x5f, 0x0a, 0x19, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x5f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x73, 0x74, 0x69,
	0x63, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x94, 0x27, 0x20, 0x01, 0x28, 0x08, 0x52, 0x17, 0x6a, 0x73, 0x6f, 0x6e, 0x69,
	0x66, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x73, 0x74,
	0x69, 0x63, 0x88, 0x01, 0x01, 0x3a, 0x6d, 0x0a, 0x21, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x5f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x63, 0x61, 0x72, 0x64, 0x5f,
	0x69, 0x66, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x95, 0x27, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x1d, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x44, 0x69, 0x73, 0x63, 0x61, 0x72, 0x64, 0x49, 0x66, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x88, 0x01, 0x01, 0x3a, 0x55, 0x0a, 0x14, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x5f, 0x6e,
	0x6f, 0x5f, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x72, 0x12, 0x1f, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x96, 0x27,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x4e, 0x6f, 0x53, 0x65,
	0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x72, 0x88, 0x01, 0x01, 0x3a, 0x59, 0x0a, 0x16, 0x6a,
	0x73, 0x6f, 0x6e, 0x69, 0x66, 0x5f, 0x6e, 0x6f, 0x5f, 0x64, 0x65, 0x73, 0x65, 0x72, 0x69, 0x61,
	0x6c, 0x69, 0x7a, 0x65, 0x72, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x97, 0x27, 0x20, 0x01, 0x28, 0x08, 0x52, 0x14, 0x6a,
	0x73, 0x6f, 0x6e, 0x69, 0x66, 0x4e, 0x6f, 0x44, 0x65, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69,
	0x7a, 0x65, 0x72, 0x88, 0x01, 0x01, 0x3a, 0x4e, 0x0a, 0x11, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66,
	0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x73, 0x74, 0x69, 0x63, 0x12, 0x1d, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x94, 0x27, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x10, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x73,
	0x74, 0x69, 0x63, 0x88, 0x01, 0x01, 0x3a, 0x5c, 0x0a, 0x19, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66,
	0x5f, 0x64, 0x69, 0x73, 0x63, 0x61, 0x72, 0x64, 0x5f, 0x69, 0x66, 0x5f, 0x64, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x95, 0x27, 0x20, 0x01, 0x28, 0x08, 0x52, 0x16, 0x6a, 0x73, 0x6f, 0x6e, 0x69,
	0x66, 0x44, 0x69, 0x73, 0x63, 0x61, 0x72, 0x64, 0x49, 0x66, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x88, 0x01, 0x01, 0x3a, 0x42, 0x0a, 0x0b, 0x6a, 0x73, 0x6f, 0x6e, 0x69, 0x66, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x96, 0x27, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6a, 0x73, 0x6f, 0x6e, 0x69,
	0x66, 0x4e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0f, 0x5a, 0x0d, 0x63, 0x6d, 0x64, 0x2f,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x58, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var file_extensions_proto_goTypes = []interface{}{
	(*descriptorpb.MessageOptions)(nil), // 0: google.protobuf.MessageOptions
	(*descriptorpb.FieldOptions)(nil),   // 1: google.protobuf.FieldOptions
}
var file_extensions_proto_depIdxs = []int32{
	0, // 0: jsonif_message_optimistic:extendee -> google.protobuf.MessageOptions
	0, // 1: jsonif_message_discard_if_default:extendee -> google.protobuf.MessageOptions
	0, // 2: jsonif_no_serializer:extendee -> google.protobuf.MessageOptions
	0, // 3: jsonif_no_deserializer:extendee -> google.protobuf.MessageOptions
	1, // 4: jsonif_optimistic:extendee -> google.protobuf.FieldOptions
	1, // 5: jsonif_discard_if_default:extendee -> google.protobuf.FieldOptions
	1, // 6: jsonif_name:extendee -> google.protobuf.FieldOptions
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	0, // [0:7] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_extensions_proto_init() }
func file_extensions_proto_init() {
	if File_extensions_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_extensions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 7,
			NumServices:   0,
		},
		GoTypes:           file_extensions_proto_goTypes,
		DependencyIndexes: file_extensions_proto_depIdxs,
		ExtensionInfos:    file_extensions_proto_extTypes,
	}.Build()
	File_extensions_proto = out.File
	file_extensions_proto_rawDesc = nil
	file_extensions_proto_goTypes = nil
	file_extensions_proto_depIdxs = nil
}
