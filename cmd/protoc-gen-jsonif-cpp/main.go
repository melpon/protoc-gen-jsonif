package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/melpon/protoc-gen-jsonif/cmd/generated"
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
func toEnumQualifiedName(name string, pkg *string, parents []*descriptorpb.DescriptorProto) (string, error) {
	qualifiedName := ""
	if pkg != nil {
		qualifiedName += "::" + strings.ReplaceAll(*pkg, ".", "::")
	}
	for _, parent := range parents {
		qualifiedName += "::" + *parent.Name
	}
	return qualifiedName, nil
}

func toTypeName(field *descriptorpb.FieldDescriptorProto) (string, string, error) {
	isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	typeName := ""
	defaultValue := ""
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		typeName = "double"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		typeName = "float"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		typeName = "int32_t"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		typeName = "int64_t"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		typeName = "uint32_t"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typeName = "uint64_t"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typeName = "bool"
		defaultValue = "false"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typeName = "std::string"
		defaultValue = ""
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		typeName = "std::string"
		defaultValue = ""
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		typeName = strings.ReplaceAll(*field.TypeName, ".", "::")
		if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
			defaultValue = fmt.Sprintf("(%s)0", typeName)
		}
	default:
		return "", "", errors.New("invalid type")
	}

	if isRepeated {
		return fmt.Sprintf("std::vector<%s>", typeName), "", nil
	} else {
		return typeName, defaultValue, nil
	}
}

func genEnum(enum *descriptorpb.EnumDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cppFile) error {
	cpp.Typedefs.PI("enum %s {", *enum.Name)
	for _, v := range enum.Value {
		cpp.Typedefs.P("%s = %d,", *v.Name, *v.Number)
	}
	cpp.Typedefs.PD("};")
	cpp.Typedefs.P("")

	qName, err := toQualifiedName(*enum.Name, pkg, parents)
	if err != nil {
		return err
	}
	qEnumName, err := toEnumQualifiedName(*enum.Name, pkg, parents)
	if err != nil {
		return err
	}
	cpp.TagInvokes.P("// %s", qName)
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("static void to_json(nlohmann::json& jv, const %s& v)", qName)
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("static void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const %s& v)", qName)
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.PI("{")
	cpp.TagInvokes.PI("switch (v) {")
	for _, v := range enum.Value {
		cpp.TagInvokes.P("case %s::%s:", qEnumName, *v.Name)
	}
	cpp.TagInvokes.Indent()
	cpp.TagInvokes.P("jv = (int)v;")
	cpp.TagInvokes.P("break;")
	cpp.TagInvokes.Deindent()
	cpp.TagInvokes.P("default:")
	cpp.TagInvokes.Indent()
	cpp.TagInvokes.P("jv = (int)(%s)0;", qName)
	cpp.TagInvokes.P("break;")
	cpp.TagInvokes.Deindent()
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("")
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.PI("static void from_json(const nlohmann::json& jv, %s& v) {", qName)
	cpp.TagInvokes.P("v = (%s)jv.template get<int>();", qName)
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.PI("static %s tag_invoke(const boost::json::value_to_tag<%s>&, const boost::json::value& jv) {", qName, qName)
	cpp.TagInvokes.P("return (%s)boost::json::value_to<int>(jv);", qName)
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.P("")

	return nil
}

func genOneof(oneof *descriptorpb.OneofDescriptorProto, fields []*descriptorpb.FieldDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cppFile) error {
	typeName := internal.ToUpperCamel(*oneof.Name) + "Case"
	fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
	cpp.Typedefs.PI("enum class %s {", typeName)
	cpp.Typedefs.P("NOT_SET = 0,")
	for _, field := range fields {
		cpp.Typedefs.P("k%s = %d,", internal.ToUpperCamel(*field.Name), *field.Number)
	}
	cpp.Typedefs.PD("};")
	cpp.Typedefs.P("%s %s = %s::NOT_SET;", typeName, fieldName, typeName)
	cpp.Typedefs.PI("void clear_%s() {", fieldName)
	cpp.Typedefs.P("%s = %s::NOT_SET;", fieldName, typeName)
	for _, field := range fields {
		fieldType, _, err := toTypeName(field)
		if err != nil {
			return err
		}
		cpp.Typedefs.P("%s = %s();", internal.ToSnakeCase(*field.Name), fieldType)
	}
	cpp.Typedefs.PD("}")
	cpp.Typedefs.P("")

	qName, err := toQualifiedName(typeName, pkg, parents)
	if err != nil {
		return err
	}
	cpp.TagInvokes.P("// %s", qName)
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("static void to_json(nlohmann::json& jv, const %s& v)", qName)
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("static void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const %s& v)", qName)
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.PI("{")
	cpp.TagInvokes.PI("switch (v) {")
	for _, field := range fields {
		cpp.TagInvokes.P("case %s::k%s:", qName, internal.ToUpperCamel(*field.Name))
	}
	cpp.TagInvokes.Indent()
	cpp.TagInvokes.P("jv = (int)v;")
	cpp.TagInvokes.P("break;")
	cpp.TagInvokes.Deindent()
	cpp.TagInvokes.P("default:")
	cpp.TagInvokes.Indent()
	cpp.TagInvokes.P("jv = (int)%s::NOT_SET;", qName)
	cpp.TagInvokes.P("break;")
	cpp.TagInvokes.Deindent()
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("")
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.PI("static void from_json(const nlohmann::json& jv, %s& v) {", qName)
	cpp.TagInvokes.P("v = (%s)jv.template get<int>();", qName)
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.PI("static %s tag_invoke(const boost::json::value_to_tag<%s>&, const boost::json::value& jv) {", qName, qName)
	cpp.TagInvokes.P("return (%s)boost::json::value_to<int>(jv);", qName)
	cpp.TagInvokes.PD("}")
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.P("")

	return nil
}

func genEquals(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cppFile) error {
	cpp.Typedefs.PI("friend bool operator==(const %s& a, const %s& b) {", *desc.Name, *desc.Name)

	// oneof 以外の比較
	for _, field := range desc.Field {
		if field.OneofIndex == nil {
			fieldName := internal.ToSnakeCase(*field.Name)
			cpp.Typedefs.P("if (a.%s != b.%s) return false;", fieldName, fieldName)
		}
	}
	// oneof の比較
	for i, oneof := range desc.OneofDecl {
		oneofFieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		oneofTypeName := internal.ToUpperCamel(*oneof.Name) + "Case"
		cpp.Typedefs.P("if (a.%s != b.%s) return false;", oneofFieldName, oneofFieldName)

		for _, field := range desc.Field {
			if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
				fieldName := internal.ToSnakeCase(*field.Name)
				enumFieldName := internal.ToUpperCamel(*field.Name)
				cpp.Typedefs.P("if (a.%s == %s::k%s && a.%s != b.%s) return false;",
					oneofFieldName, oneofTypeName, enumFieldName, fieldName, fieldName)
			}
		}
	}
	cpp.Typedefs.P("return true;")
	cpp.Typedefs.PD("}")
	cpp.Typedefs.P("friend bool operator!=(const %s& a, const %s& b) { return !(a == b); }", *desc.Name, *desc.Name)

	return nil
}

func genDescriptor(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cppFile) error {
	descOptimistic := proto.HasExtension(desc.Options, generated.E_JsonifMessageOptimistic) && proto.GetExtension(desc.Options, generated.E_JsonifMessageOptimistic).(bool)
	descDiscard := proto.HasExtension(desc.Options, generated.E_JsonifMessageDiscardIfDefault) && proto.GetExtension(desc.Options, generated.E_JsonifMessageDiscardIfDefault).(bool)
	noSerializer := proto.HasExtension(desc.Options, generated.E_JsonifNoSerializer) && proto.GetExtension(desc.Options, generated.E_JsonifNoSerializer).(bool)
	noDeserializer := proto.HasExtension(desc.Options, generated.E_JsonifNoDeserializer) && proto.GetExtension(desc.Options, generated.E_JsonifNoDeserializer).(bool)

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

	for i, oneof := range desc.OneofDecl {
		var fields []*descriptorpb.FieldDescriptorProto
		for _, field := range desc.Field {
			if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
				fields = append(fields, field)
			}
		}
		if err := genOneof(oneof, fields, pkg, append(parents, desc), cpp); err != nil {
			return err
		}
	}

	for _, field := range desc.Field {
		typeName, defaultValue, err := toTypeName(field)
		if err != nil {
			return err
		}
		fieldName := internal.ToSnakeCase(*field.Name)
		if len(defaultValue) != 0 {
			defaultValue = " = " + defaultValue
		}
		cpp.Typedefs.P("%s %s%s;", typeName, fieldName, defaultValue)

		if oneof := field.OneofIndex; oneof != nil {
			oneofTypeName := internal.ToUpperCamel(*desc.OneofDecl[*oneof].Name) + "Case"
			oneofFieldName := internal.ToSnakeCase(*desc.OneofDecl[*oneof].Name) + "_case"
			cpp.Typedefs.PI("void set_%s(%s %s) {", fieldName, typeName, fieldName)
			cpp.Typedefs.P("clear_%s();", oneofFieldName)
			cpp.Typedefs.P("%s = %s::k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
			cpp.Typedefs.P("this->%s = %s;", fieldName, fieldName)
			cpp.Typedefs.PD("}")
			if field.Proto3Optional != nil && *field.Proto3Optional {
				cpp.Typedefs.PI("bool has_%s() const {", fieldName)
				cpp.Typedefs.P("return %s == %s::k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
				cpp.Typedefs.PD("}")
			}
			cpp.Typedefs.PI("void clear_%s() {", fieldName)
			cpp.Typedefs.PI("if (%s == %s::k%s) {", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
			cpp.Typedefs.P("clear_%s();", oneofFieldName)
			cpp.Typedefs.PD("}")
			cpp.Typedefs.PD("}")
		}
	}

	err := genEquals(desc, pkg, append(parents, desc), cpp)
	if err != nil {
		return err
	}

	cpp.Typedefs.PD("};")
	cpp.Typedefs.P("")

	qName, err := toQualifiedName(*desc.Name, pkg, parents)
	if err != nil {
		return err
	}
	cpp.TagInvokes.P("// %s", qName)
	if noSerializer {
		cpp.TagInvokes.P("#if 0")
	}
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("static void to_json(nlohmann::json& jv, const %s& v)", qName)
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("static void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const %s& v)", qName)
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.PI("{")
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("nlohmann::json obj;")
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("boost::json::object obj;")
	cpp.TagInvokes.P("#endif")
	for _, field := range desc.Field {
		fieldName := internal.ToSnakeCase(*field.Name)
		fieldKey := internal.GetJsonName(field, fieldName)
		discard := descDiscard
		if proto.HasExtension(field.Options, generated.E_JsonifDiscardIfDefault) {
			discard = proto.GetExtension(field.Options, generated.E_JsonifDiscardIfDefault).(bool)
		}

		if discard {
			cpp.TagInvokes.PI("if (v.%s != decltype(v.%s)()) {", fieldName, fieldName)
		}
		cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
		cpp.TagInvokes.PI("{")
		cpp.TagInvokes.P("using nlohmann::to_json;")
		cpp.TagInvokes.P("to_json(obj[\"%s\"], v.%s);", fieldKey, fieldName)
		cpp.TagInvokes.PD("}")
		cpp.TagInvokes.P("#else")
		cpp.TagInvokes.P("obj[\"%s\"] = boost::json::value_from(v.%s);", fieldKey, fieldName)
		cpp.TagInvokes.P("#endif")
		if discard {
			cpp.TagInvokes.PD("}")
		}
	}
	for _, oneof := range desc.OneofDecl {
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
		cpp.TagInvokes.PI("{")
		cpp.TagInvokes.P("using nlohmann::to_json;")
		cpp.TagInvokes.P("to_json(obj[\"%s\"], v.%s);", fieldName, fieldName)
		cpp.TagInvokes.PD("}")
		cpp.TagInvokes.P("#else")
		cpp.TagInvokes.P("obj[\"%s\"] = boost::json::value_from(v.%s);", fieldName, fieldName)
		cpp.TagInvokes.P("#endif")
	}
	cpp.TagInvokes.P("jv = std::move(obj);")
	cpp.TagInvokes.PD("}")
	if noSerializer {
		cpp.TagInvokes.P("#endif")
	}
	cpp.TagInvokes.P("")
	if noDeserializer {
		cpp.TagInvokes.P("#if 0")
	}
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("static void from_json(const nlohmann::json& jv, %s& v)", qName)
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("static %s tag_invoke(const boost::json::value_to_tag<%s>&, const boost::json::value& jv)", qName, qName)
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.PI("{")
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("%s v;", qName)
	cpp.TagInvokes.P("#endif")
	for _, field := range desc.Field {
		typeName, _, err := toTypeName(field)
		if err != nil {
			return err
		}
		fieldName := internal.ToSnakeCase(*field.Name)
		fieldKey := internal.GetJsonName(field, fieldName)
		optimistic := descOptimistic
		if proto.HasExtension(field.Options, generated.E_JsonifOptimistic) {
			optimistic = proto.GetExtension(field.Options, generated.E_JsonifOptimistic).(bool)
		}
		if field.OneofIndex != nil || optimistic {
			cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
			cpp.TagInvokes.P("if (jv.contains(\"%s\"))", fieldKey)
			cpp.TagInvokes.P("#else")
			cpp.TagInvokes.P("if (jv.as_object().find(\"%s\") != jv.as_object().end())", fieldKey)
			cpp.TagInvokes.P("#endif")
			cpp.TagInvokes.PI("{")
		}
		cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
		cpp.TagInvokes.PI("{")
		cpp.TagInvokes.P("using nlohmann::from_json;")
		cpp.TagInvokes.P("from_json(jv.at(\"%s\"), v.%s);", fieldKey, fieldName)
		cpp.TagInvokes.PD("}")
		cpp.TagInvokes.P("#else")
		cpp.TagInvokes.P("v.%s = boost::json::value_to<%s>(jv.at(\"%s\"));", fieldName, typeName, fieldKey)
		cpp.TagInvokes.P("#endif")
		if field.OneofIndex != nil || optimistic {
			cpp.TagInvokes.PD("}")
		}
	}
	for _, oneof := range desc.OneofDecl {
		typeName, err := toQualifiedName(internal.ToUpperCamel(*oneof.Name)+"Case", pkg, append(parents, desc))
		if err != nil {
			return err
		}
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
		cpp.TagInvokes.PI("{")
		cpp.TagInvokes.P("using nlohmann::from_json;")
		cpp.TagInvokes.P("from_json(jv.at(\"%s\"), v.%s);", fieldName, fieldName)
		cpp.TagInvokes.PD("}")
		cpp.TagInvokes.P("#else")
		cpp.TagInvokes.P("v.%s = boost::json::value_to<%s>(jv.at(\"%s\"));", fieldName, typeName, fieldName)
		cpp.TagInvokes.P("#endif")
	}
	cpp.TagInvokes.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.TagInvokes.P("#else")
	cpp.TagInvokes.P("return v;")
	cpp.TagInvokes.P("#endif")
	cpp.TagInvokes.PD("}")
	if noDeserializer {
		cpp.TagInvokes.P("#endif")
	}
	cpp.TagInvokes.P("")

	return nil
}

// 大文字と数字はそのまま、小文字は大文字に、それ以外は _ にする
// test/foo.proto → TEST_FOO_PROTO
func toPreprocessorName(name string) string {
	r := ""
	for _, c := range name {
		switch {
		case 'A' <= c && c <= 'Z':
			r += string(rune(c))
		case '0' <= c && c <= '9':
			r += string(rune(c))
		case 'a' <= c && c <= 'z':
			r += string(rune(c - 'a' + 'A'))
		default:
			r += "_"
		}
	}
	return r
}

func genFile(file *descriptorpb.FileDescriptorProto, files []*descriptorpb.FileDescriptorProto) (*pluginpb.CodeGeneratorResponse_File, error) {
	var pkgs []string
	if file.Package != nil {
		pkgs = strings.Split(*file.Package, ".")
	}

	cpp := cppFile{}
	cpp.Top.P("#ifndef AUTO_GENERATED_PROTOC_GEN_JSONIF_CPP_%s", toPreprocessorName(*file.Name))
	cpp.Top.P("#define AUTO_GENERATED_PROTOC_GEN_JSONIF_CPP_%s", toPreprocessorName(*file.Name))
	cpp.Top.P("")
	cpp.Top.P("#include <string>")
	cpp.Top.P("#include <vector>")
	cpp.Top.P("#include <stddef.h>")
	cpp.Top.P("")
	cpp.Top.P("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.Top.P("#include <nlohmann/json.hpp>")
	cpp.Top.P("#else")
	cpp.Top.P("#include <boost/json.hpp>")
	cpp.Top.P("#endif")
	cpp.Top.P("")
	for _, dep := range file.Dependency {
		// ファイルが存在してない可能性もあるのでチェックする
		exists := false
		for _, f := range files {
			if dep == *f.Name {
				exists = true
				break
			}
		}
		if !exists {
			continue
		}

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
	cpp.Bottom.PI("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.Bottom.P("T t;")
	cpp.Bottom.P("from_json(nlohmann::json::parse(s), t);")
	cpp.Bottom.P("return t;")
	cpp.Bottom.PDI("#else")
	cpp.Bottom.P("return boost::json::value_to<T>(boost::json::parse(s));")
	cpp.Bottom.PD("#endif")
	cpp.Bottom.PD("}")
	cpp.Bottom.P("")
	cpp.Bottom.P("template<class T>")
	cpp.Bottom.PI("inline std::string to_json(const T& v) {")
	cpp.Bottom.PI("#if defined(JSONIF_USE_NLOHMANN_JSON)")
	cpp.Bottom.P("using nlohmann::to_json;")
	cpp.Bottom.P("nlohmann::json j;")
	cpp.Bottom.P("to_json(j, v);")
	cpp.Bottom.P("return j.dump();")
	cpp.Bottom.PDI("#else")
	cpp.Bottom.P("return boost::json::serialize(boost::json::value_from(v));")
	cpp.Bottom.PD("#endif")
	cpp.Bottom.PD("}")
	cpp.Bottom.P("")
	cpp.Bottom.P("}")
	cpp.Bottom.P("")
	cpp.Bottom.P("#endif")
	cpp.Bottom.P("")
	cpp.Bottom.P("#endif")

	for _, enum := range file.EnumType {
		if err := genEnum(enum, file.Package, nil, &cpp); err != nil {
			return nil, err
		}
	}

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
	resp.SupportedFeatures = proto.Uint64(uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))
	for _, file := range req.ProtoFile {
		respFile, err := genFile(file, req.ProtoFile)
		if err != nil {
			return nil, err
		}
		resp.File = append(resp.File, respFile)
	}
	return resp, nil
}

func main() {
	err := internal.RunPlugin(gen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}
