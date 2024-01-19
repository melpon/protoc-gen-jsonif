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

type cFile struct {
	HTop        internal.Formatter
	HBottom     internal.Formatter
	Enums       internal.Formatter
	Typedefs    internal.Formatter
	CTop        internal.Formatter
	CBottom     internal.Formatter
	CImplTop    internal.Formatter
	CImpl       internal.Formatter
	CImplBottom internal.Formatter
	CppImpl     internal.Formatter
	HppTop      internal.Formatter
	HppBottom   internal.Formatter
	HppDefs     internal.Formatter
}

func (cpp *cFile) HeaderString() string {
	return cpp.HTop.String() + cpp.Enums.String() + cpp.Typedefs.String() + cpp.HBottom.String()
}
func (cpp *cFile) HppString() string {
	return cpp.HppTop.String() + cpp.HppDefs.String() + cpp.HppBottom.String()
}
func (cpp *cFile) CppString() string {
	return cpp.CTop.String() + cpp.CppImpl.String() + cpp.CImplTop.String() + cpp.CImpl.String() + cpp.CImplBottom.String() + cpp.CBottom.String()
}

func toQualifiedName(name string, pkg *string, parents []*descriptorpb.DescriptorProto) (string, error) {
	qualifiedName := ""
	if pkg != nil {
		qualifiedName += strings.ReplaceAll(*pkg, ".", "_") + "_"
	}
	for _, parent := range parents {
		qualifiedName += *parent.Name + "_"
	}
	qualifiedName += name
	return qualifiedName, nil
}
func toEnumQualifiedName(name string, pkg *string, parents []*descriptorpb.DescriptorProto) (string, error) {
	qualifiedName := ""
	if pkg != nil {
		qualifiedName += strings.ReplaceAll(*pkg, ".", "_")
	}
	for _, parent := range parents {
		if len(qualifiedName) == 0 {
			qualifiedName += *parent.Name
		} else {
			qualifiedName += "_" + *parent.Name
		}
	}
	return qualifiedName, nil
}
func toCppQualifiedName(name string, pkg *string, parents []*descriptorpb.DescriptorProto) (string, error) {
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

func getMessageTypeName(field *descriptorpb.FieldDescriptorProto) (string, error) {
	if *field.Type != descriptorpb.FieldDescriptorProto_TYPE_MESSAGE &&
		*field.Type != descriptorpb.FieldDescriptorProto_TYPE_ENUM &&
		*field.Type != descriptorpb.FieldDescriptorProto_TYPE_GROUP {
		return "", errors.New("not message type")
	}
	typeName := *field.TypeName
	if strings.HasPrefix(typeName, ".") {
		typeName = typeName[1:]
	}
	typeName = strings.ReplaceAll(typeName, ".", "_")
	return typeName, nil
}

func toTypeName(field *descriptorpb.FieldDescriptorProto) (string, bool, bool, error) {
	isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	typeName := ""
	needLen := false
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		typeName = "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		typeName = "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		typeName = "int32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		typeName = "int64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		typeName = "uint32_t"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typeName = "uint64_t"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typeName = "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typeName = "char*"
		needLen = true
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		typeName = "uint8_t*"
		needLen = true
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		var err error
		typeName, err = getMessageTypeName(field)
		if err != nil {
			return "", false, false, err
		}
	default:
		return "", false, false, errors.New("invalid type")
	}

	if isRepeated {
		return fmt.Sprintf("%s*", typeName), true, needLen, nil
	} else {
		return typeName, false, needLen, nil
	}
}

func genEnum(enum *descriptorpb.EnumDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cFile) error {
	cpp.Enums.P("// %s", *enum.Name)

	qName, err := toQualifiedName(*enum.Name, pkg, parents)
	if err != nil {
		return err
	}
	qEnumName, err := toEnumQualifiedName(*enum.Name, pkg, parents)
	if err != nil {
		return err
	}

	cpp.Enums.P("typedef int %s;", qName)
	for _, v := range enum.Value {
		cpp.Enums.P("extern const %s %s_%s;", qName, qEnumName, *v.Name)
	}
	cpp.Enums.P("")

	cpp.CppImpl.P("// %s", *enum.Name)
	for _, v := range enum.Value {
		cpp.CppImpl.P("const %s %s_%s = %d;", qName, qEnumName, *v.Name, *v.Number)
	}
	cpp.CppImpl.P("")

	return nil
}

func genOneofEnum(oneof *descriptorpb.OneofDescriptorProto, fields []*descriptorpb.FieldDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cFile) error {
	typeName := internal.ToUpperCamel(*oneof.Name) + "Case"
	qName, err := toQualifiedName(typeName, pkg, parents)
	if err != nil {
		return err
	}
	cpp.Enums.P("// %s", *oneof.Name)
	cpp.Enums.P("typedef int %s;", qName)
	cpp.Enums.P("const %s %s_NOT_SET = 0;", qName, qName)
	for _, field := range fields {
		cpp.Enums.P("const %s %s_k%s = %d;", qName, qName, internal.ToUpperCamel(*field.Name), *field.Number)
	}
	cpp.Enums.P("")

	return nil
}

func genDescriptor(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto, cpp *cFile) error {
	// descOptimistic := proto.HasExtension(desc.Options, generated.E_JsonifMessageOptimistic) && proto.GetExtension(desc.Options, generated.E_JsonifMessageOptimistic).(bool)
	// descDiscard := proto.HasExtension(desc.Options, generated.E_JsonifMessageDiscardIfDefault) && proto.GetExtension(desc.Options, generated.E_JsonifMessageDiscardIfDefault).(bool)
	// noSerializer := proto.HasExtension(desc.Options, generated.E_JsonifNoSerializer) && proto.GetExtension(desc.Options, generated.E_JsonifNoSerializer).(bool)
	// noDeserializer := proto.HasExtension(desc.Options, generated.E_JsonifNoDeserializer) && proto.GetExtension(desc.Options, generated.E_JsonifNoDeserializer).(bool)

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
		if err := genOneofEnum(oneof, fields, pkg, append(parents, desc), cpp); err != nil {
			return err
		}
	}

	qName, err := toQualifiedName(*desc.Name, pkg, parents)
	if err != nil {
		return err
	}

	cpp.Typedefs.P("// %s", *desc.Name)
	cpp.Typedefs.PI("typedef struct {")

	for _, field := range desc.Field {
		typeName, isRepeated, needLen, err := toTypeName(field)
		if err != nil {
			return err
		}
		fieldName := internal.ToSnakeCase(*field.Name)
		cpp.Typedefs.P("%s %s;", typeName, fieldName)
		if isRepeated && needLen {
			cpp.Typedefs.P("int* %s_lens;", fieldName)
		}
		if isRepeated || needLen {
			cpp.Typedefs.P("int %s_len;", fieldName)
		}
	}

	for _, oneof := range desc.OneofDecl {
		typeName := internal.ToUpperCamel(*oneof.Name) + "Case"
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		qName, err := toQualifiedName(typeName, pkg, append(parents, desc))
		if err != nil {
			return err
		}
		cpp.Typedefs.P("%s %s;", qName, fieldName)
	}

	// for _, field := range desc.Field {
	// 	typeName, defaultValue, err := toTypeName(field)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fieldName := internal.ToSnakeCase(*field.Name)
	// 	if len(defaultValue) != 0 {
	// 		defaultValue = " = " + defaultValue
	// 	}
	// 	cpp.Typedefs.P("%s %s%s;", typeName, fieldName, defaultValue)

	// 	if oneof := field.OneofIndex; oneof != nil {
	// 		oneofTypeName := internal.ToUpperCamel(*desc.OneofDecl[*oneof].Name) + "Case"
	// 		oneofFieldName := internal.ToSnakeCase(*desc.OneofDecl[*oneof].Name) + "_case"
	// 		cpp.Typedefs.PI("void set_%s(%s %s) {", fieldName, typeName, fieldName)
	// 		cpp.Typedefs.P("clear_%s();", oneofFieldName)
	// 		cpp.Typedefs.P("%s = %s::k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
	// 		cpp.Typedefs.P("this->%s = %s;", fieldName, fieldName)
	// 		cpp.Typedefs.PD("}")
	// 	}
	// }

	// err := genEquals(desc, pkg, append(parents, desc), cpp)
	// if err != nil {
	// 	return err
	// }

	cpp.Typedefs.PD("} %s;", qName)
	cpp.Typedefs.P("")

	// qName, err := toQualifiedName(*desc.Name, pkg, parents)
	// if err != nil {
	// 	return err
	// }
	// cpp.TagInvokes.P("// %s", qName)
	// if noSerializer {
	// 	cpp.TagInvokes.P("#if 0")
	// }
	// cpp.TagInvokes.PI("static void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const %s& v) {", qName)
	// cpp.TagInvokes.P("boost::json::object obj;")
	// for _, field := range desc.Field {
	// 	fieldName := internal.ToSnakeCase(*field.Name)
	// 	fieldKey := internal.GetJsonName(field, fieldName)
	// 	discard := descDiscard
	// 	if proto.HasExtension(field.Options, generated.E_JsonifDiscardIfDefault) {
	// 		discard = proto.GetExtension(field.Options, generated.E_JsonifDiscardIfDefault).(bool)
	// 	}

	// 	if discard {
	// 		cpp.TagInvokes.PI("if (v.%s != decltype(v.%s)()) {", fieldName, fieldName)
	// 	}
	// 	cpp.TagInvokes.P("obj[\"%s\"] = boost::json::value_from(v.%s);", fieldKey, fieldName)
	// 	if discard {
	// 		cpp.TagInvokes.PD("}")
	// 	}
	// }
	// for _, oneof := range desc.OneofDecl {
	// 	fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
	// 	cpp.TagInvokes.P("obj[\"%s\"] = boost::json::value_from(v.%s);", fieldName, fieldName)
	// }
	// cpp.TagInvokes.P("jv = std::move(obj);")
	// cpp.TagInvokes.PD("}")
	// if noSerializer {
	// 	cpp.TagInvokes.P("#endif")
	// }
	// cpp.TagInvokes.P("")
	// if noDeserializer {
	// 	cpp.TagInvokes.P("#if 0")
	// }
	// cpp.TagInvokes.PI("static %s tag_invoke(const boost::json::value_to_tag<%s>&, const boost::json::value& jv) {", qName, qName)
	// cpp.TagInvokes.P("%s v;", qName)
	// for _, field := range desc.Field {
	// 	typeName, _, err := toTypeName(field)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fieldName := internal.ToSnakeCase(*field.Name)
	// 	fieldKey := internal.GetJsonName(field, fieldName)
	// 	optimistic := descOptimistic
	// 	if proto.HasExtension(field.Options, generated.E_JsonifOptimistic) {
	// 		optimistic = proto.GetExtension(field.Options, generated.E_JsonifOptimistic).(bool)
	// 	}
	// 	if field.OneofIndex != nil || optimistic {
	// 		cpp.TagInvokes.PI("if (jv.as_object().find(\"%s\") != jv.as_object().end()) {", fieldKey)
	// 	}
	// 	cpp.TagInvokes.P("v.%s = boost::json::value_to<%s>(jv.at(\"%s\"));", fieldName, typeName, fieldKey)
	// 	if field.OneofIndex != nil || optimistic {
	// 		cpp.TagInvokes.PD("}")
	// 	}
	// }
	// for _, oneof := range desc.OneofDecl {
	// 	typeName, err := toQualifiedName(internal.ToUpperCamel(*oneof.Name)+"Case", pkg, append(parents, desc))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
	// 	cpp.TagInvokes.P("v.%s = boost::json::value_to<%s>(jv.at(\"%s\"));", fieldName, typeName, fieldName)
	// }
	// cpp.TagInvokes.P("return v;")
	// cpp.TagInvokes.PD("}")
	// if noDeserializer {
	// 	cpp.TagInvokes.P("#endif")
	// }
	// cpp.TagInvokes.P("")
	cpp.Typedefs.P("int %s_size();", qName)
	cpp.Typedefs.P("void %s_init(%s* v);", qName, qName)
	cpp.Typedefs.P("void %s_destroy(%s*);", qName, qName)
	cpp.Typedefs.P("void %s_copy(const %s* a, %s* b);", qName, qName, qName)
	cpp.Typedefs.P("bool %s_is_equal(const %s* a, const %s* b);", qName, qName, qName)
	cpp.Typedefs.P("int %s_to_json_size(const %s*);", qName, qName)
	cpp.Typedefs.P("void %s_to_json(const %s*, char* json);", qName, qName)
	cpp.Typedefs.P("void %s_from_json(const char* json, %s*);", qName, qName)
	for _, field := range desc.Field {
		fieldName := internal.ToSnakeCase(*field.Name)
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		if !isRepeated {
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.Typedefs.P("void %s_set_%s(%s* v, const char* s);", qName, fieldName, qName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.Typedefs.P("void %s_set_%s(%s* v, const uint8_t* buf, int size);", qName, fieldName, qName)
			} else {
				typeName, _, _, err := toTypeName(field)
				if err != nil {
					return err
				}
				if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
					cpp.Typedefs.P("void %s_set_%s(%s* v, const %s* m);", qName, fieldName, qName, typeName)
				} else {
					cpp.Typedefs.P("void %s_set_%s(%s* v, %s m);", qName, fieldName, qName, typeName)
				}
			}
		}
		if isRepeated {
			cpp.Typedefs.P("void %s_alloc_%s(%s* v, int num);", qName, fieldName, qName)
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.Typedefs.P("void %s_set_%s(%s* v, int n, const char* s);", qName, fieldName, qName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.Typedefs.P("void %s_set_%s(%s* v, int n, const uint8_t* buf, int size);", qName, fieldName, qName)
			} else {
				typeName, _, _, err := toTypeName(field)
				if err != nil {
					return err
				}
				typeName = strings.ReplaceAll(typeName, "*", "")
				if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
					cpp.Typedefs.P("void %s_set_%s(%s* v, int n, const %s* m);", qName, fieldName, qName, typeName)
				} else {
					cpp.Typedefs.P("void %s_set_%s(%s* v, int n, %s m);", qName, fieldName, qName, typeName)
				}
			}
		}
	}
	cpp.Typedefs.P("")

	// oneof clear_<case> declarations
	for _, oneof := range desc.OneofDecl {
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		cpp.Typedefs.P("void %s_clear_%s(%s* v);", qName, fieldName, qName)
	}

	qCppName, err := toCppQualifiedName(*desc.Name, pkg, parents)
	// C++ 専用の宣言
	cpp.HppDefs.P("%s %s_to_cpp(const %s* v);", qCppName, qName, qName)
	cpp.HppDefs.P("void %s_from_cpp(const %s& u, %s* v);", qName, qCppName, qName)

	// to_cpp
	cpp.CppImpl.PI("%s %s_to_cpp(const %s* v) {", qCppName, qName, qName)
	cpp.CppImpl.P("%s u;", qCppName)
	for _, field := range desc.Field {
		fieldName := internal.ToSnakeCase(*field.Name)
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		if !isRepeated {
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CppImpl.P("if (v->%s_len != 0) u.%s = std::string(v->%s, v->%s_len);", fieldName, fieldName, fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CppImpl.P("if (v->%s_len != 0) u.%s = std::string((const char*)v->%s, v->%s_len);", fieldName, fieldName, fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName, err := getMessageTypeName(field)
				if err != nil {
					return err
				}
				cpp.CppImpl.P("u.%s = %s_to_cpp(&v->%s);", fieldName, typeName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				cpp.CppImpl.P("u.%s = (decltype(u.%s))v->%s;", fieldName, fieldName, fieldName)
			} else {
				cpp.CppImpl.P("u.%s = v->%s;", fieldName, fieldName)
			}
		}
		if isRepeated {
			cpp.CppImpl.PI("for (int i = 0; i < v->%s_len; i++) {", fieldName)
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CppImpl.PI("if (v->%s_lens[i] != 0) {", fieldName)
				cpp.CppImpl.P("u.%s.push_back(std::string(v->%s[i], v->%s_lens[i]));", fieldName, fieldName, fieldName)
				cpp.CppImpl.PDI("} else {")
				cpp.CppImpl.P("u.%s.push_back(\"\");", fieldName)
				cpp.CppImpl.PD("}")
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CppImpl.PI("if (v->%s_lens[i] != 0) {", fieldName)
				cpp.CppImpl.P("u.%s.push_back(std::string((const char*)v->%s[i], v->%s_lens[i]));", fieldName, fieldName, fieldName)
				cpp.CppImpl.PDI("} else {")
				cpp.CppImpl.P("u.%s.push_back(\"\");", fieldName)
				cpp.CppImpl.PD("}")
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName, err := getMessageTypeName(field)
				if err != nil {
					return err
				}
				cpp.CppImpl.P("u.%s.push_back(%s_to_cpp(&v->%s[i]));", fieldName, typeName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				cpp.CppImpl.P("u.%s.push_back((decltype(u.%s[0]))v->%s[i]);", fieldName, fieldName, fieldName)
			} else {
				cpp.CppImpl.P("u.%s.push_back(v->%s[i]);", fieldName, fieldName)
			}
			cpp.CppImpl.PD("}")
		}
	}
	for _, oneof := range desc.OneofDecl {
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		typeName := internal.ToUpperCamel(*oneof.Name) + "Case"
		oneofQName, err := toCppQualifiedName(typeName, pkg, append(parents, desc))
		if err != nil {
			return err
		}
		cpp.CppImpl.P("u.%s = (%s)v->%s;", fieldName, oneofQName, fieldName)
	}
	cpp.CppImpl.P("return u;")
	cpp.CppImpl.PD("}")

	// from_cpp
	cpp.CppImpl.PI("void %s_from_cpp(const %s& u, %s* v) {", qName, qCppName, qName)
	cpp.CppImpl.P("%s_destroy(v);", qName)
	cpp.CppImpl.P("%s_init(v);", qName)
	for _, field := range desc.Field {
		fieldName := internal.ToSnakeCase(*field.Name)
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		if !isRepeated {
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CppImpl.P("if (!u.%s.empty()) v->%s = strdup(u.%s.c_str());", fieldName, fieldName, fieldName)
				cpp.CppImpl.P("v->%s_len = (int)u.%s.size();", fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CppImpl.PI("if (!u.%s.empty()) {", fieldName)
				cpp.CppImpl.P("v->%s = (uint8_t*)malloc(sizeof(uint8_t) * u.%s.size());", fieldName, fieldName)
				cpp.CppImpl.P("memcpy(v->%s, u.%s.data(), u.%s.size());", fieldName, fieldName, fieldName)
				cpp.CppImpl.PD("}")
				cpp.CppImpl.P("v->%s_len = (int)u.%s.size();", fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName, err := getMessageTypeName(field)
				if err != nil {
					return err
				}
				cpp.CppImpl.P("%s_from_cpp(u.%s, &v->%s);", typeName, fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				cpp.CppImpl.P("v->%s = (int)u.%s;", fieldName, fieldName)
			} else {
				cpp.CppImpl.P("v->%s = u.%s;", fieldName, fieldName)
			}
		}
		if isRepeated {
			cpp.CppImpl.P("v->%s_len = (int)u.%s.size();", fieldName, fieldName)
			cpp.CppImpl.P("v->%s = v->%s_len == 0 ? nullptr : (decltype(v->%s))malloc(sizeof(v->%s[0]) * u.%s.size());",
				fieldName, fieldName, fieldName, fieldName, fieldName)
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING ||
				*field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CppImpl.P("v->%s_lens = v->%s_len == 0 ? nullptr : (int*)malloc(sizeof(int) * u.%s.size());",
					fieldName, fieldName, fieldName)
			}

			cpp.CppImpl.PI("for (int i = 0; i < (int)u.%s.size(); i++) {", fieldName)
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CppImpl.P("if (!u.%s[i].empty()) v->%s[i] = strdup(u.%s[i].c_str());", fieldName, fieldName, fieldName)
				cpp.CppImpl.P("v->%s_lens[i] = (int)u.%s[i].size();", fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CppImpl.PI("if (!u.%s[i].empty()) {", fieldName)
				cpp.CppImpl.P("v->%s[i] = (uint8_t*)malloc(sizeof(uint8_t) * u.%s[i].size());", fieldName, fieldName)
				cpp.CppImpl.P("memcpy(v->%s[i], u.%s[i].data(), u.%s[i].size());", fieldName, fieldName, fieldName)
				cpp.CppImpl.PD("}")
				cpp.CppImpl.P("v->%s_lens[i] = (int)u.%s[i].size();", fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName, err := getMessageTypeName(field)
				if err != nil {
					return err
				}
				cpp.CppImpl.P("%s_init(&v->%s[i]);", typeName, fieldName)
				cpp.CppImpl.P("%s_from_cpp(u.%s[i], &v->%s[i]);", typeName, fieldName, fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				cpp.CppImpl.P("v->%s[i] = (int)u.%s[i];", fieldName, fieldName)
			} else {
				cpp.CppImpl.P("v->%s[i] = u.%s[i];", fieldName, fieldName)
			}
			cpp.CppImpl.PD("}")
		}
	}
	for _, oneof := range desc.OneofDecl {
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		cpp.CppImpl.P("v->%s = (int)u.%s;", fieldName, fieldName)
	}
	cpp.CppImpl.PD("}")

	// size
	cpp.CImpl.PI("int %s_size() {", qName)
	cpp.CImpl.P("return sizeof(%s);", qName)
	cpp.CImpl.PD("}")

	// init
	cpp.CImpl.PI("void %s_init(%s* v) {", qName, qName)
	cpp.CImpl.P("memset(v, 0, sizeof(%s));", qName)
	cpp.CImpl.PD("}")

	// destroy
	cpp.CImpl.PI("void %s_destroy(%s* v) {", qName, qName)
	for _, field := range desc.Field {
		fieldName := internal.ToSnakeCase(*field.Name)
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		if !isRepeated {
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
				cpp.CImpl.P("v->%s = nullptr;", fieldName)
				cpp.CImpl.P("v->%s_len = 0;", fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
				cpp.CImpl.P("v->%s = nullptr;", fieldName)
				cpp.CImpl.P("v->%s_len = 0;", fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName, err := getMessageTypeName(field)
				if err != nil {
					return err
				}
				cpp.CImpl.P("%s_destroy(&v->%s);", typeName, fieldName)
			} else {
				cpp.CImpl.P("memset(&v->%s, 0, sizeof(v->%s));", fieldName, fieldName)
			}
		}
		if isRepeated {
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING ||
				*field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CImpl.PI("for (int i = 0; i < v->%s_len; i++) {", fieldName)
				cpp.CImpl.P("if (v->%s[i]) free(v->%s[i]);", fieldName, fieldName)
				cpp.CImpl.P("v->%s[i] = nullptr;", fieldName)
				cpp.CImpl.P("v->%s_lens[i] = 0;", fieldName)
				cpp.CImpl.PD("}")
				cpp.CImpl.P("if (v->%s_lens) free(v->%s_lens);", fieldName, fieldName)
				cpp.CImpl.P("v->%s_lens = nullptr;", fieldName)
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName, err := getMessageTypeName(field)
				if err != nil {
					return err
				}
				cpp.CImpl.PI("for (int i = 0; i < v->%s_len; i++) {", fieldName)
				cpp.CImpl.P("%s_destroy(&v->%s[i]);", typeName, fieldName)
				cpp.CImpl.PD("}")
			}
			cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
			cpp.CImpl.P("v->%s = nullptr;", fieldName)
			cpp.CImpl.P("v->%s_len = 0;", fieldName)
		}
	}
	cpp.CImpl.PD("}")

	// copy
	cpp.CImpl.PI("void %s_copy(const %s* a, %s* b) {", qName, qName, qName)
	cpp.CImpl.P("if (a == b) return;")
	cpp.CImpl.P("int size = %s_to_json_size(a);", qName)
	cpp.CImpl.P("std::string json(size - 1, 0);")
	cpp.CImpl.P("%s_to_json(a, &json[0]);", qName)
	cpp.CImpl.P("%s_from_json(json.c_str(), b);", qName)
	cpp.CImpl.PD("}")

	// is_equal
	cpp.CImpl.PI("bool %s_is_equal(const %s* a, const %s* b) {", qName, qName, qName)
	cpp.CImpl.P("if (a == b) return true;")
	cpp.CImpl.P("%s ua = %s_to_cpp(a);", qCppName, qName)
	cpp.CImpl.P("%s ub = %s_to_cpp(b);", qCppName, qName)
	cpp.CImpl.P("return ua == ub;")
	cpp.CImpl.PD("}")

	// to_json_size
	cpp.CImpl.PI("int %s_to_json_size(const %s* v) {", qName, qName)
	cpp.CImpl.P("%s u = %s_to_cpp(v);", qCppName, qName)
	cpp.CImpl.P("return jsonif::to_json(u).size() + 1;")
	cpp.CImpl.PD("}")

	// to_json
	cpp.CImpl.PI("void %s_to_json(const %s* v, char* json) {", qName, qName)
	cpp.CImpl.P("%s u = %s_to_cpp(v);", qCppName, qName)
	cpp.CImpl.P("std::string str = jsonif::to_json(u);")
	cpp.CImpl.P("memcpy(json, str.c_str(), str.size() + 1);")
	cpp.CImpl.PD("}")

	// from_json
	cpp.CImpl.PI("void %s_from_json(const char* json, %s* v) {", qName, qName)
	cpp.CImpl.P("%s u = jsonif::from_json<%s>(json);", qCppName, qCppName)
	cpp.CImpl.P("%s_from_cpp(u, v);", qName)
	cpp.CImpl.PD("}")

	// set_<field>
	for _, field := range desc.Field {
		fieldName := internal.ToSnakeCase(*field.Name)
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		if !isRepeated {
			genCase := func() error {
				if oneof := field.OneofIndex; oneof != nil {
					oneofTypeName := internal.ToUpperCamel(*desc.OneofDecl[*oneof].Name) + "Case"
					oneofFieldName := internal.ToSnakeCase(*desc.OneofDecl[*oneof].Name) + "_case"
					oneofQName, err := toQualifiedName(oneofTypeName, pkg, append(parents, desc))
					if err != nil {
						return err
					}
					cpp.CImpl.P("%s_clear_%s(v);", qName, oneofFieldName)
					cpp.CImpl.P("v->%s = %s_k%s;", oneofFieldName, oneofQName, internal.ToUpperCamel(*field.Name))
				}
				return nil
			}

			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CImpl.PI("void %s_set_%s(%s* v, const char* s) {", qName, fieldName, qName)
				err := genCase()
				if err != nil {
					return err
				}
				cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
				cpp.CImpl.P("v->%s_len = s == nullptr ? 0 : strlen(s);", fieldName)
				cpp.CImpl.P("v->%s = v->%s_len == 0 ? nullptr : strdup(s);", fieldName, fieldName)
				cpp.CImpl.PD("}")
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CImpl.PI("void %s_set_%s(%s* v, const uint8_t* buf, int size) {", qName, fieldName, qName)
				err := genCase()
				if err != nil {
					return err
				}
				cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
				cpp.CImpl.P("v->%s = nullptr;", fieldName)
				cpp.CImpl.P("v->%s_len = buf == nullptr ? 0 : size;", fieldName)
				cpp.CImpl.PI("if (v->%s_len != 0) {", fieldName)
				cpp.CImpl.P("v->%s = (uint8_t*)malloc(size);", fieldName)
				cpp.CImpl.P("memcpy(v->%s, buf, size);", fieldName)
				cpp.CImpl.PD("}")
				cpp.CImpl.PD("}")
			} else {
				typeName, _, _, err := toTypeName(field)
				if err != nil {
					return err
				}
				if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
					cpp.CImpl.PI("void %s_set_%s(%s* v, const %s* m) {", qName, fieldName, qName, typeName)
					err := genCase()
					if err != nil {
						return err
					}
					cpp.CImpl.P("%s_copy(m, &v->%s);", typeName, fieldName)
					cpp.CImpl.PD("}")
				} else {
					cpp.CImpl.PI("void %s_set_%s(%s* v, %s m) {", qName, fieldName, qName, typeName)
					err := genCase()
					if err != nil {
						return err
					}
					cpp.CImpl.P("v->%s = m;", fieldName)
					cpp.CImpl.PD("}")
				}
			}
		}
		if isRepeated {
			cpp.CImpl.PI("void %s_alloc_%s(%s* v, int num) {", qName, fieldName, qName)
			cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
			cpp.CImpl.P("v->%s = nullptr;", fieldName)
			cpp.CImpl.P("v->%s_len = 0;", fieldName)
			cpp.CImpl.PI("if (num != 0) {")
			cpp.CImpl.P("v->%s = (decltype(v->%s))malloc(sizeof(v->%s[0]) * num);", fieldName, fieldName, fieldName)
			cpp.CImpl.P("memset(v->%s, 0, sizeof(v->%s[0]) * num);", fieldName, fieldName)
			cpp.CImpl.P("v->%s_len = num;", fieldName)
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING ||
				*field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CImpl.P("v->%s_lens = (decltype(v->%s_lens))malloc(sizeof(v->%s_lens[0]) * num);", fieldName, fieldName, fieldName)
				cpp.CImpl.P("memset(v->%s_lens, 0, sizeof(v->%s_lens[0]) * num);", fieldName, fieldName)
			}
			cpp.CImpl.PD("}")
			cpp.CImpl.PD("}")
			cpp.CImpl.P("")
			if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
				cpp.CImpl.PI("void %s_set_%s(%s* v, int n, const char* s) {", qName, fieldName, qName)
				cpp.CImpl.P("if (v->%s[n]) free(v->%s[n]);", fieldName, fieldName)
				cpp.CImpl.P("v->%s_lens[n] = s == nullptr ? 0 : strlen(s);", fieldName)
				cpp.CImpl.P("v->%s[n] = v->%s_lens[n] == 0 ? nullptr : strdup(s);", fieldName, fieldName)
				cpp.CImpl.PD("}")
			} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
				cpp.CImpl.PI("void %s_set_%s(%s* v, int n, const uint8_t* buf, int size) {", qName, fieldName, qName)
				cpp.CImpl.P("if (v->%s[n]) free(v->%s[n]);", fieldName, fieldName)
				cpp.CImpl.P("v->%s[n] = nullptr;", fieldName)
				cpp.CImpl.P("v->%s_lens[n] = buf == nullptr ? 0 : size;", fieldName)
				cpp.CImpl.PI("if (v->%s_lens[n] != 0) {", fieldName)
				cpp.CImpl.P("v->%s[n] = (uint8_t*)malloc(size);", fieldName)
				cpp.CImpl.P("memcpy(v->%s[n], buf, size);", fieldName)
				cpp.CImpl.PD("}")
				cpp.CImpl.PD("}")
			} else {
				typeName, _, _, err := toTypeName(field)
				if err != nil {
					return err
				}
				typeName = strings.ReplaceAll(typeName, "*", "")
				if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
					cpp.CImpl.PI("void %s_set_%s(%s* v, int n, const %s* m) {", qName, fieldName, qName, typeName)
					cpp.CImpl.P("%s_copy(m, &v->%s[n]);", typeName, fieldName)
					cpp.CImpl.PD("}")
				} else {
					cpp.CImpl.PI("void %s_set_%s(%s* v, int n, %s m) {", qName, fieldName, qName, typeName)
					cpp.CImpl.P("v->%s[n] = m;", fieldName)
					cpp.CImpl.PD("}")
				}
			}
		}
	}

	// oneof clear_<case>
	for i, oneof := range desc.OneofDecl {
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		typeName := internal.ToUpperCamel(*oneof.Name) + "Case"
		oneofQName, err := toQualifiedName(typeName, pkg, append(parents, desc))
		if err != nil {
			return err
		}
		cpp.CImpl.PI("void %s_clear_%s(%s* v) {", qName, fieldName, qName)

		var fields []*descriptorpb.FieldDescriptorProto
		for _, field := range desc.Field {
			if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
				fields = append(fields, field)
			}
		}
		for _, field := range fields {
			// destroy 実装からのコピペ
			fieldName := internal.ToSnakeCase(*field.Name)
			isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
			if !isRepeated {
				if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING {
					cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
					cpp.CImpl.P("v->%s = nullptr;", fieldName)
					cpp.CImpl.P("v->%s_len = 0;", fieldName)
				} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
					cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
					cpp.CImpl.P("v->%s = nullptr;", fieldName)
					cpp.CImpl.P("v->%s_len = 0;", fieldName)
				} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
					typeName, err := getMessageTypeName(field)
					if err != nil {
						return err
					}
					cpp.CImpl.P("%s_destroy(&v->%s);", typeName, fieldName)
				} else {
					cpp.CImpl.P("memset(&v->%s, 0, sizeof(v->%s));", fieldName, fieldName)
				}
			}
			if isRepeated {
				if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_STRING ||
					*field.Type == descriptorpb.FieldDescriptorProto_TYPE_BYTES {
					cpp.CImpl.PI("for (int i = 0; i < v->%s_len; i++) {", fieldName)
					cpp.CImpl.P("if (v->%s[i]) free(v->%s[i]);", fieldName, fieldName)
					cpp.CImpl.P("v->%s[i] = nullptr;", fieldName)
					cpp.CImpl.P("v->%s_lens[i] = 0;", fieldName)
					cpp.CImpl.PD("}")
					cpp.CImpl.P("if (v->%s_lens) free(v->%s_lens);", fieldName, fieldName)
				} else if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
					typeName, err := getMessageTypeName(field)
					if err != nil {
						return err
					}
					cpp.CImpl.PI("for (int i = 0; i < v->%s_len; i++) {", fieldName)
					cpp.CImpl.P("%s_destroy(&v->%s[i]);", typeName, fieldName)
					cpp.CImpl.PD("}")
				}
				cpp.CImpl.P("if (v->%s) free(v->%s);", fieldName, fieldName)
				cpp.CImpl.P("v->%s_len = 0;", fieldName)
			}
		}
		cpp.CImpl.P("v->%s = %s_NOT_SET;", fieldName, oneofQName)
		cpp.CImpl.PD("}")
	}

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

func genFile(file *descriptorpb.FileDescriptorProto, files []*descriptorpb.FileDescriptorProto) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	// 拡張子を取り除いて .json.h を付ける
	fileName := *file.Name
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
	cpphFileName := fileName + ".json.h"
	hFileName := fileName + ".json.c.h"
	hppFileName := fileName + ".json.c.hpp"
	cppFileName := fileName + ".json.c.cpp"

	depFileNames := []string{}
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

		// 拡張子を取り除く
		fileName := dep
		fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
		depFileNames = append(depFileNames, fileName)
	}

	cpp := cFile{}
	cpp.HTop.P("#ifndef AUTO_GENERATED_PROTOC_GEN_JSONIF_C_%s", toPreprocessorName(*file.Name))
	cpp.HTop.P("#define AUTO_GENERATED_PROTOC_GEN_JSONIF_C_%s", toPreprocessorName(*file.Name))
	cpp.HTop.P("")
	cpp.HTop.P("#include <stdbool.h>")
	cpp.HTop.P("#include <stddef.h>")
	cpp.HTop.P("#include <stdint.h>")
	cpp.HTop.P("")
	for _, fileName := range depFileNames {
		cpp.HTop.P("#include \"%s\"", fileName+".json.c.h")
	}
	cpp.HTop.P("")
	cpp.HTop.P("#ifdef __cplusplus")
	cpp.HTop.P("extern \"C\" {")
	cpp.HTop.P("#endif")
	cpp.HTop.P("")
	cpp.HBottom.P("")
	cpp.HBottom.P("#ifdef __cplusplus")
	cpp.HBottom.P("}")
	cpp.HBottom.P("#endif")
	cpp.HBottom.P("")
	cpp.HBottom.P("#endif")

	cpp.CTop.P("#include \"%s\"", hFileName)
	cpp.CTop.P("")
	cpp.CTop.P("#include <stdlib.h>")
	cpp.CTop.P("#include <string.h>")
	cpp.CTop.P("")
	cpp.CTop.P("#include \"%s\"", cpphFileName)
	cpp.CTop.P("")
	for _, fileName := range depFileNames {
		cpp.CTop.P("#include \"%s\"", fileName+".json.c.hpp")
	}
	cpp.CTop.P("")
	cpp.CImplTop.P("extern \"C\" {")
	cpp.CImplTop.P("")
	cpp.CImplBottom.P("")
	cpp.CImplBottom.P("}")

	cpp.HppTop.P("#ifndef AUTO_GENERATED_PROTOC_GEN_JSONIF_HPP_%s", toPreprocessorName(*file.Name))
	cpp.HppTop.P("#define AUTO_GENERATED_PROTOC_GEN_JSONIF_HPP_%s", toPreprocessorName(*file.Name))
	cpp.HppTop.P("")
	cpp.HppTop.P("#include \"%s\"", cpphFileName)
	cpp.HppTop.P("#include \"%s\"", hFileName)
	cpp.HppTop.P("")
	for _, fileName := range depFileNames {
		cpp.HppTop.P("#include \"%s\"", fileName+".json.c.hpp")
	}
	cpp.HppTop.P("")
	cpp.HppBottom.P("")
	cpp.HppBottom.P("#endif")

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

	hContent := cpp.HeaderString()
	hppContent := cpp.HppString()
	cppContent := cpp.CppString()
	resp := []*pluginpb.CodeGeneratorResponse_File{}
	resp = append(resp, &pluginpb.CodeGeneratorResponse_File{
		Name:    &hFileName,
		Content: &hContent,
	}, &pluginpb.CodeGeneratorResponse_File{
		Name:    &hppFileName,
		Content: &hppContent,
	}, &pluginpb.CodeGeneratorResponse_File{
		Name:    &cppFileName,
		Content: &cppContent,
	})
	return resp, nil
}

func gen(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	resp := &pluginpb.CodeGeneratorResponse{}
	resp.SupportedFeatures = proto.Uint64(uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))
	for _, file := range req.ProtoFile {
		respFiles, err := genFile(file, req.ProtoFile)
		if err != nil {
			return nil, err
		}
		resp.File = append(resp.File, respFiles...)
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
