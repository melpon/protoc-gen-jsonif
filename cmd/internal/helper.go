package internal

import (
	"google.golang.org/protobuf/types/descriptorpb"
)

func GetJsonName(field *descriptorpb.FieldDescriptorProto, defaultName string) string {
	if field.JsonName == nil {
		return defaultName
	}
	return *field.JsonName
}
