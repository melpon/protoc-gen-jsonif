package internal

import (
	"github.com/melpon/protoc-gen-jsonif/cmd/generated"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func GetJsonName(field *descriptorpb.FieldDescriptorProto, defaultName string) string {
	if !proto.HasExtension(field.Options, generated.E_JsonifName) {
		return defaultName
	}
	return proto.GetExtension(field.Options, generated.E_JsonifName).(string)
}
