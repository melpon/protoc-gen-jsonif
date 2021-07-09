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

type unityFile struct {
	Top      internal.Formatter
	Bottom   internal.Formatter
	Typedefs internal.Formatter
}

/*
namespace Tutorial
{
    [System.Serializable]
    public class Person
    {
        [System.Serializable]
        public enum PhoneType
        {
            PhoneType_Invalid = -1,
            MOBILE = 0,
            HOME = 1,
            WORK = 2,
        }

        [System.Serializable]
        public class PhoneNumber
        {
            public string number;
            public PhoneType type;
        }

        public string name;
        public int id;
        public string email;
        public PhoneNumber[] phones;
        //Google.Protobuf.Timestamp LastUpdated;
    }

    [System.Serializable]
    public class AddressBook
    {
        public Person[] people;
    }
}
*/
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

func toTypeName(field *descriptorpb.FieldDescriptorProto) (string, error) {
	isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	typeName := ""
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		typeName = "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		typeName = "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		typeName = "int"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		typeName = "long"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		typeName = "ulong"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		typeName = "uint"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		typeName = "int"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		typeName = "long"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		typeName = "int"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		typeName = "long"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typeName = "ulong"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		typeName = "uint"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typeName = "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typeName = "string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		//typeName = "std::string"
		return "", errors.New("bytes type not supported")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		typeName = "global::" + packageToNamespace((*field.TypeName)[1:])
	default:
		return "", errors.New("invalid type")
	}

	if isRepeated {
		return fmt.Sprintf("%s[]", typeName), nil
	} else {
		return typeName, nil
	}
}

func genEnum(enum *descriptorpb.EnumDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, u *unityFile) error {
	u.Typedefs.P("[System.Serializable]")
	u.Typedefs.P("public enum %s", *enum.Name)
	u.Typedefs.PI("{")
	u.Typedefs.P("%s_Invalid = -1,", *enum.Name)
	for _, v := range enum.Value {
		u.Typedefs.P("%s = %d,", *v.Name, *v.Number)
	}
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

	for _, field := range desc.Field {
		typeName, err := toTypeName(field)
		if err != nil {
			return err
		}
		u.Typedefs.P("public %s %s;", typeName, *field.Name)
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

	if file.Package != nil {
		u.Top.P("namespace %s", packageToNamespace(*file.Package))
		u.Top.PI("{")
		u.Top.P("")
	}

	u.Bottom.P("}")

	u.Typedefs.Indent()
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
