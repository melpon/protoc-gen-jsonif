package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/melpon/protoc-gen-jsonif/cmd/internal"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type unityFile struct {
	Top      internal.Formatter
	Bottom   internal.Formatter
	Typedefs internal.Formatter
}

func (u *unityFile) String() string {
	return u.Top.String() + u.Typedefs.String() + u.Bottom.String()
}

// foo.bar_baz を Foo.BarBaz に変換する
func packageToNamespace(pkg string) string {
	xs := strings.Split(pkg, ".")
	for i, x := range xs {
		xs[i] = internal.ToUpperCamel(x)
	}
	return strings.Join(xs, ".")
}

// foo/bar_baz.txt を Foo/BarBaz.txt に変換する
func pathToUpperCamel(pkg string) string {
	xs := strings.Split(pkg, "/")
	for i, x := range xs {
		xs[i] = internal.ToUpperCamel(x)
	}
	return strings.Join(xs, "/")
}

func toTypeName(field *descriptorpb.FieldDescriptorProto) (string, string, error) {
	isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	typeName := ""
	defaultValue := ""
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		typeName = "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		typeName = "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		typeName = "int"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		typeName = "long"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		typeName = "uint"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typeName = "ulong"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typeName = "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typeName = "string"
		defaultValue = "\"\""
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		//typeName = "std::string"
		return "", "", errors.New("bytes type not supported")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		typeName = "global::" + packageToNamespace((*field.TypeName)[1:])
		defaultValue = fmt.Sprintf("new %s()", typeName)
	default:
		return "", "", errors.New("invalid type")
	}

	if isRepeated {
		return fmt.Sprintf("List<%s>", typeName), fmt.Sprintf("new List<%s>()", typeName), nil
	} else {
		return typeName, defaultValue, nil
	}
}

func genEnum(enum *descriptorpb.EnumDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, u *unityFile) error {
	u.Typedefs.P("[System.Serializable]")
	u.Typedefs.P("public enum %s", *enum.Name)
	u.Typedefs.PI("{")
	for _, v := range enum.Value {
		u.Typedefs.P("%s = %d,", *v.Name, *v.Number)
	}
	u.Typedefs.PD("}")
	u.Typedefs.P("")
	return nil
}

func genOneof(oneof *descriptorpb.OneofDescriptorProto, fields []*descriptorpb.FieldDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, u *unityFile) error {
	typeName := internal.ToUpperCamel(*oneof.Name) + "Case"
	fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
	u.Typedefs.P("[System.Serializable]")
	u.Typedefs.P("public enum %s", typeName)
	u.Typedefs.PI("{")
	u.Typedefs.P("NOT_SET = 0,")
	for _, field := range fields {
		u.Typedefs.P("k%s = %d,", internal.ToUpperCamel(*field.Name), *field.Number)
	}
	u.Typedefs.PD("}")
	u.Typedefs.P("public %s %s;", typeName, fieldName)
	u.Typedefs.P("public void Clear%s()", typeName)
	u.Typedefs.PI("{")
	u.Typedefs.P("%s = %s.NOT_SET;", fieldName, typeName)
	for _, field := range fields {
		fieldType, defaultValue, err := toTypeName(field)
		if err != nil {
			return err
		}
		if len(defaultValue) == 0 {
			u.Typedefs.P("%s = default(%s);", internal.ToSnakeCase(*field.Name), fieldType)
		} else {
			u.Typedefs.P("%s = %s;", internal.ToSnakeCase(*field.Name), defaultValue)
		}
	}
	u.Typedefs.PD("}")
	return nil
}

func genEquals(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, u *unityFile) error {
	u.Typedefs.P("public override bool Equals(object obj)")
	u.Typedefs.PI("{")
	u.Typedefs.P("var v = obj as %s;", *desc.Name)
	u.Typedefs.P("if (v == null) return false;")

	// oneof 以外の比較
	for _, field := range desc.Field {
		if field.OneofIndex == nil {
			fieldName := internal.ToSnakeCase(*field.Name)
			if *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
				// List の場合は SequenceEqual で比較する
				u.Typedefs.P("if (!this.%s.SequenceEqual(v.%s)) return false;", fieldName, fieldName)
			} else {
				u.Typedefs.P("if (!this.%s.Equals(v.%s)) return false;", fieldName, fieldName)
			}
		}
	}
	// oneof の比較
	for i, oneof := range desc.OneofDecl {
		oneofFieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		oneofTypeName := internal.ToUpperCamel(*oneof.Name) + "Case"
		u.Typedefs.P("if (!this.%s.Equals(v.%s)) return false;", oneofFieldName, oneofFieldName)

		for _, field := range desc.Field {
			if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
				fieldName := internal.ToSnakeCase(*field.Name)
				enumFieldName := internal.ToUpperCamel(*field.Name)
				u.Typedefs.P("if (this.%s == %s.k%s && !this.%s.Equals(v.%s)) return false;",
					oneofFieldName, oneofTypeName, enumFieldName, fieldName, fieldName)
			}
		}
	}
	u.Typedefs.P("return true;")
	u.Typedefs.PD("}")
	u.Typedefs.P("")
	u.Typedefs.P("public override int GetHashCode()")
	u.Typedefs.PI("{")
	u.Typedefs.P("int hashcode = 1430287;")

	// oneof 以外のハッシュ値
	for _, field := range desc.Field {
		if field.OneofIndex == nil {
			fieldName := internal.ToSnakeCase(*field.Name)
			if *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
				// List の場合は各要素のハッシュ値を取得する
				u.Typedefs.P("foreach (var v in this.%s) hashcode = hashcode * 7302013 ^ v.GetHashCode();", fieldName)
			} else {
				u.Typedefs.P("hashcode = hashcode * 7302013 ^ %s.GetHashCode();", fieldName)
			}
		}
	}
	// oneof のハッシュ値
	for i, oneof := range desc.OneofDecl {
		oneofFieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		oneofTypeName := internal.ToUpperCamel(*oneof.Name) + "Case"
		u.Typedefs.P("hashcode = hashcode * 7302013 ^ %s.GetHashCode();", oneofFieldName)

		for _, field := range desc.Field {
			if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
				fieldName := internal.ToSnakeCase(*field.Name)
				enumFieldName := internal.ToUpperCamel(*field.Name)
				u.Typedefs.P("if (%s == %s.k%s) hashcode = hashcode * 7302013 ^ %s.GetHashCode();",
					oneofFieldName, oneofTypeName, enumFieldName, fieldName)
			}
		}
	}
	u.Typedefs.P("return hashcode;")
	u.Typedefs.PD("}")
	u.Typedefs.P("")

	return nil
}

func genDescriptor(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, u *unityFile) error {
	u.Typedefs.P("[System.Serializable]")
	u.Typedefs.P("public class %s", *desc.Name)
	u.Typedefs.PI("{")

	for _, enum := range desc.EnumType {
		if err := genEnum(enum, pkg, append(parents, desc), u); err != nil {
			return err
		}
	}

	for _, nested := range desc.NestedType {
		if err := genDescriptor(nested, pkg, append(parents, desc), u); err != nil {
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
		if err := genOneof(oneof, fields, pkg, append(parents, desc), u); err != nil {
			return err
		}
	}

	for _, field := range desc.Field {
		typeName, defaultValue, err := toTypeName(field)
		if err != nil {
			return err
		}
		fieldName := internal.ToSnakeCase(*field.Name)
		if len(defaultValue) == 0 {
			u.Typedefs.P("public %s %s;", typeName, fieldName)
		} else {
			u.Typedefs.P("public %s %s = %s;", typeName, fieldName, defaultValue)
		}

		if oneof := field.OneofIndex; oneof != nil {
			oneofTypeName := internal.ToUpperCamel(*desc.OneofDecl[*oneof].Name) + "Case"
			oneofFieldName := internal.ToSnakeCase(*desc.OneofDecl[*oneof].Name) + "_case"
			u.Typedefs.P("public void Set%s(%s %s)", internal.ToUpperCamel(fieldName), typeName, fieldName)
			u.Typedefs.PI("{")
			u.Typedefs.P("Clear%s();", oneofTypeName)
			u.Typedefs.P("%s = %s.k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
			u.Typedefs.P("this.%s = %s;", fieldName, fieldName)
			u.Typedefs.PD("}")
			u.Typedefs.P("public bool Has%s()", internal.ToUpperCamel(fieldName))
			u.Typedefs.PI("{")
			u.Typedefs.P("return %s == %s.k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
			u.Typedefs.PD("}")
			u.Typedefs.P("public void Clear%s()", internal.ToUpperCamel(fieldName))
			u.Typedefs.PI("{")
			u.Typedefs.P("if (%s == %s.k%s)", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
			u.Typedefs.PI("{")
			u.Typedefs.P("Clear%s();", oneofTypeName)
			u.Typedefs.PD("}")
			u.Typedefs.PD("}")
		}
	}

	err := genEquals(desc, pkg, append(parents, desc), u)
	if err != nil {
		return err
	}

	u.Typedefs.PD("}")
	u.Typedefs.P("")

	return nil
}

func genFile(file *descriptorpb.FileDescriptorProto) (*pluginpb.CodeGeneratorResponse_File, error) {
	u := unityFile{}
	u.Top.SetIndentUnit(4)
	u.Bottom.SetIndentUnit(4)
	u.Typedefs.SetIndentUnit(4)

	u.Top.P("using System.Collections.Generic;")
	u.Top.P("using System.Linq;")

	if file.Package != nil {
		u.Top.P("namespace %s", packageToNamespace(*file.Package))
		u.Top.PI("{")
		u.Top.P("")
	}

	u.Bottom.P("}")

	u.Typedefs.Indent()
	for _, enum := range file.EnumType {
		if err := genEnum(enum, file.Package, nil, &u); err != nil {
			return nil, err
		}
	}
	for _, desc := range file.MessageType {
		if err := genDescriptor(desc, file.Package, nil, &u); err != nil {
			return nil, err
		}
	}
	u.Typedefs.Deindent()

	// UpperCamel にして拡張子を取り除いて .cs を付ける
	fileName := pathToUpperCamel(*file.Name)
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
	fileName = fileName + ".cs"

	content := u.String()
	resp := &pluginpb.CodeGeneratorResponse_File{
		Name:    &fileName,
		Content: &content,
	}
	return resp, nil
}

func genJsonif() (*pluginpb.CodeGeneratorResponse_File, error) {
	f := internal.Formatter{}
	f.SetIndentUnit(4)

	f.P("using UnityEngine;")
	f.P("")
	f.P("namespace Jsonif")
	f.PI("{")
	f.P("")
	f.P("public static class Json")
	f.PI("{")
	f.P("public static string ToJson<T>(T v)")
	f.PI("{")
	f.P("return JsonUtility.ToJson(v);")
	f.PD("}")
	f.P("public static T FromJson<T>(string s)")
	f.PI("{")
	f.P("return JsonUtility.FromJson<T>(s);")
	f.PD("}")
	f.PD("}")
	f.P("")
	f.PD("}")

	fileName := "Jsonif.cs"

	content := f.String()
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
		respFile, err := genFile(file)
		if err != nil {
			return nil, err
		}
		resp.File = append(resp.File, respFile)
	}

	// 共通実装
	respFile, err := genJsonif()
	if err != nil {
		return nil, err
	}
	resp.File = append(resp.File, respFile)

	return resp, nil
}

func main() {
	err := internal.RunPlugin(gen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}
