package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/melpon/protoc-gen-jsonif/cmd/internal"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type cppFile struct {
	Top        internal.Formatter
	Bottom     internal.Formatter
	Typedefs   internal.Formatter
	TagInvokes internal.Formatter
}

func (cpp *cppFile) String() string {
	return cpp.Top.String() + cpp.Typedefs.String() + cpp.TagInvokes.String() + cpp.Bottom.String()
}

func toQualifiedName(name string, pkg *string, parents []*descriptorpb.DescriptorProto) (string, error) {
	qualifiedName := ""
	if pkg != nil {
		qualifiedName += "::" + strings.ReplaceAll(*pkg, ".", "::")
	}
	for _, parent := range parents {
		qualifiedName += "::" + *parent.Name
	}
	qualifiedName += "::" + name
	return qualifiedName, nil
}

func toTypeName(field *descriptorpb.FieldDescriptorProto) (string, error) {
	isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	typeName := ""
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		typeName = "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		typeName = "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		typeName = "int32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		typeName = "int64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		typeName = "uint64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		typeName = "uint32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		typeName = "int32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		typeName = "int64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		typeName = "int32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		typeName = "int64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typeName = "uint64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		typeName = "uint32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typeName = "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typeName = "std::string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		//typeName = "std::string"
		return "", errors.New("bytes type not supported")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		typeName = strings.ReplaceAll(*field.TypeName, ".", "::")
	default:
		return "", errors.New("invalid type")
	}

	if isRepeated {
		return fmt.Sprintf("std::vector<%s>", typeName), nil
	} else {
		return typeName, nil
	}
}

func genEnum(enum *descriptorpb.EnumDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cppFile) error {
	cpp.Typedefs.PI("enum class %s {", *enum.Name)
	cpp.Typedefs.P("%s_Invalid = -1,", *enum.Name)
	for _, v := range enum.Value {
		cpp.Typedefs.P("%s = %d,", *v.Name, *v.Number)
	}
	cpp.Typedefs.PD("};")
	cpp.Typedefs.P("")

	qName, err := toQualifiedName(*enum.Name, pkg, parents)
	if err != nil {
		return err
	}
	cpp.TagInvokes.P("// %s", qName)
	cpp.TagInvokes.PI("void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const %s& v) {", qName)
	cpp.TagInvokes.PI("switch (v) {")
	for _, v := range enum.Value {
		cpp.TagInvokes.P("case %s::%s:", qName, *v.Name)
	}
	cpp.TagInvokes.Indent()
	cpp.TagInvokes.P("jv = (int)v;")
	cpp.TagInvokes.P("break;")
	cpp.TagInvokes.Deindent()
	cpp.TagInvokes.P("default:")
	cpp.TagInvokes.Indent()
	cpp.TagInvokes.P("jv = (int)%s::%s_Invalid;", qName, *enum.Name)
	cpp.TagInvokes.P("break;")
	cpp.TagInvokes.Deindent()
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("")
	cpp.TagInvokes.PI("%s tag_invoke(const boost::json::value_to_tag<%s>&, const boost::json::value& jv) {", qName, qName)
	cpp.TagInvokes.P("return (%s)boost::json::value_to<int>(jv);", qName)
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("")

	return nil
}

func genDescriptor(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cppFile) error {
	cpp.Typedefs.PI("struct %s {", *desc.Name)

	for _, enum := range desc.EnumType {
		if err := genEnum(enum, pkg, append(parents, desc), cpp); err != nil {
			return err
		}
	}

	for _, nested := range desc.NestedType {
		if err := genDescriptor(nested, pkg, append(parents, desc), cpp); err != nil {
			return err
		}
	}

	for _, field := range desc.Field {
		typeName, err := toTypeName(field)
		if err != nil {
			return err
		}
		cpp.Typedefs.P("%s %s;", typeName, *field.Name)
	}

	cpp.Typedefs.PD("};")
	cpp.Typedefs.P("")

	qName, err := toQualifiedName(*desc.Name, pkg, parents)
	if err != nil {
		return err
	}
	cpp.TagInvokes.P("// %s", qName)
	cpp.TagInvokes.PI("void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const %s& v) {", qName)
	cpp.TagInvokes.PI("jv = {")
	for _, field := range desc.Field {
		cpp.TagInvokes.P("{\"%s\", boost::json::value_from(v.%s)},", *field.Name, *field.Name)
	}
	cpp.TagInvokes.PD("};")
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("")
	cpp.TagInvokes.PI("%s tag_invoke(const boost::json::value_to_tag<%s>&, const boost::json::value& jv) {", qName, qName)
	cpp.TagInvokes.P("%s v;", qName)
	for _, field := range desc.Field {
		typeName, err := toTypeName(field)
		if err != nil {
			return err
		}
		cpp.TagInvokes.P("v.%s = boost::json::value_to<%s>(jv.at(\"%s\"));", *field.Name, typeName, *field.Name)
	}
	cpp.TagInvokes.P("return v;")
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("")

	return nil
}

func genFile(file *descriptorpb.FileDescriptorProto) (*pluginpb.CodeGeneratorResponse_File, error) {
	var pkgs []string
	if file.Package != nil {
		pkgs = strings.Split(*file.Package, ".")
	}

	cpp := cppFile{}
	cpp.Top.P("#include <string>")
	cpp.Top.P("#include <vector>")
	cpp.Top.P("#include <stddef.h>")
	cpp.Top.P("")
	cpp.Top.P("#include <boost/json.hpp>")
	cpp.Top.P("")
	for _, dep := range file.Dependency {
		// 拡張子を取り除いて .json.h を付ける
		fileName := dep
		fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
		fileName = fileName + ".json.h"
		cpp.Top.P("#include \"%s\"", fileName)
	}
	cpp.Top.P("")
	for _, pkg := range pkgs {
		cpp.Top.P("namespace %s {", pkg)
	}
	cpp.Top.P("")

	cpp.Bottom.P("")
	for range pkgs {
		cpp.Bottom.P("}")
	}
	cpp.Bottom.P("")
	cpp.Bottom.P("#ifndef JSONIF_HELPER_DEFINED")
	cpp.Bottom.P("#define JSONIF_HELPER_DEFINED")
	cpp.Bottom.P("")
	cpp.Bottom.P("namespace jsonif {")
	cpp.Bottom.P("")
	cpp.Bottom.P("template<class T>")
	cpp.Bottom.PI("inline T from_json(const std::string& s) {")
	cpp.Bottom.P("return boost::json::value_to<T>(boost::json::parse(s));")
	cpp.Bottom.PD("}")
	cpp.Bottom.P("")
	cpp.Bottom.P("template<class T>")
	cpp.Bottom.PI("inline std::string to_json(const T& v) {")
	cpp.Bottom.P("return boost::json::serialize(boost::json::value_from(v));")
	cpp.Bottom.PD("}")
	cpp.Bottom.P("")
	cpp.Bottom.P("}")
	cpp.Bottom.P("")
	cpp.Bottom.P("#endif")

	for _, desc := range file.MessageType {
		if err := genDescriptor(desc, file.Package, nil, &cpp); err != nil {
			return nil, err
		}
	}

	// 拡張子を取り除いて .json.h を付ける
	fileName := *file.Name
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
	fileName = fileName + ".json.h"

	content := cpp.String()
	resp := &pluginpb.CodeGeneratorResponse_File{
		Name:    &fileName,
		Content: &content,
	}
	return resp, nil
}

func gen(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	resp := &pluginpb.CodeGeneratorResponse{}
	for _, file := range req.ProtoFile {
		respFile, err := genFile(file)
		if err != nil {
			return nil, err
		}
		resp.File = append(resp.File, respFile)
	}
	return resp, nil
}

func run() error {
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(in, req); err != nil {
		return err
	}

	resp, err := gen(req)
	if err != nil {
		return err
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(out); err != nil {
		return err
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}
