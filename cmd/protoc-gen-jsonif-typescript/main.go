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

type typescriptFile struct {
	Top    internal.Formatter
	Bottom internal.Formatter
	Body   internal.Formatter
}

func (u *typescriptFile) String() string {
	return u.Top.String() + u.Body.String() + u.Bottom.String()
}

type pkgInfo struct {
	fullnameToType    map[string]string
	fullnameToPackage map[string]string
	packageToAlias    map[string]string
	filenameToAlias   map[string]string
}

func newPkgInfo() *pkgInfo {
	return &pkgInfo{
		fullnameToType:    make(map[string]string),
		fullnameToPackage: make(map[string]string),
		packageToAlias:    make(map[string]string),
		filenameToAlias:   make(map[string]string),
	}
}

func (p *pkgInfo) enumEnum(enum *descriptorpb.EnumDescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto) {
	fullname := "." + *pkg + "." + toClassName(parents, *enum.Name)
	p.fullnameToType[fullname] = toLocalClassName(parents, *enum.Name)
	p.fullnameToPackage[fullname] = *pkg
}
func (p *pkgInfo) enumDescriptor(desc *descriptorpb.DescriptorProto, pkg *string, parents []*descriptorpb.DescriptorProto) {
	fullname := "." + *pkg + "." + toClassName(parents, *desc.Name)
	p.fullnameToType[fullname] = toLocalClassName(parents, *desc.Name)
	p.fullnameToPackage[fullname] = *pkg
	for _, enum := range desc.EnumType {
		p.enumEnum(enum, pkg, append(parents, desc))
	}
	for _, nested := range desc.NestedType {
		p.enumDescriptor(nested, pkg, append(parents, desc))
	}
}

func (p *pkgInfo) addFile(file *descriptorpb.FileDescriptorProto) {
	alias := strings.ReplaceAll(*file.Package, ".", "_")
	p.packageToAlias[*file.Package] = alias
	p.filenameToAlias[*file.Name] = alias
	for _, enum := range file.EnumType {
		p.enumEnum(enum, file.Package, nil)
	}
	for _, desc := range file.MessageType {
		p.enumDescriptor(desc, file.Package, nil)
	}
}

func (p *pkgInfo) findTypeName(pkg string, typeName string) (string, error) {
	var name string
	var targetPkg string
	var ok bool
	if name, ok = p.fullnameToType[typeName]; !ok {
		return "", errors.New("not found")
	}
	if targetPkg, ok = p.fullnameToPackage[typeName]; !ok {
		return "", errors.New("not found")
	}
	// 処理中のパッケージと、このクラスのパッケージ名が一致してたら name をそのまま返す
	if pkg == targetPkg {
		return name, nil
	}
	// 処理中のパッケージと、このクラスのパッケージ名が一致してない場合、インポート時のエイリアスを付け加える
	var alias string
	if alias, ok = p.packageToAlias[targetPkg]; !ok {
		return "", errors.New("not found")
	}
	return alias + "." + name, nil
}

// toLocalClassName([Foo, Bar], Baz) を Foo_Bar_Baz に変換する
func toLocalClassName(parents []*descriptorpb.DescriptorProto, name string) string {
	var xs []string
	for _, parent := range parents {
		xs = append(xs, *parent.Name)
	}
	xs = append(xs, name)
	return strings.Join(xs, "_")
}

// toClassName([Foo, Bar], Baz) を Foo.Bar.Baz に変換する
func toClassName(parents []*descriptorpb.DescriptorProto, name string) string {
	var xs []string
	for _, parent := range parents {
		xs = append(xs, *parent.Name)
	}
	xs = append(xs, name)
	return strings.Join(xs, ".")
}

func getOneofFields(fields []*descriptorpb.FieldDescriptorProto, i int) []*descriptorpb.FieldDescriptorProto {
	var r []*descriptorpb.FieldDescriptorProto
	for _, field := range fields {
		if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
			if field.Proto3Optional != nil && *field.Proto3Optional {
				// optional だった場合はフィールドを生成しない
			} else {
				r = append(r, field)
			}
		}
	}
	return r
}

func toTypeName(pkg *string, pkgInfo *pkgInfo, field *descriptorpb.FieldDescriptorProto, forObject bool) (string, string, bool, error) {
	isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	isOptional := field.Proto3Optional != nil && *field.Proto3Optional
	typeName := ""
	defaultValue := ""
	var err error
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typeName = "number"
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typeName = "boolean"
		defaultValue = "false"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typeName = "string"
		defaultValue = "\"\""
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		typeName = "Uint8Array"
		defaultValue = "new Uint8Array(0)"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typeName, err = pkgInfo.findTypeName(*pkg, *field.TypeName)
		if err != nil {
			return "", "", false, fmt.Errorf("type not found: %s", *field.TypeName)
		}
		defaultValue = "0"
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		typeName, err = pkgInfo.findTypeName(*pkg, *field.TypeName)
		if err != nil {
			return "", "", false, fmt.Errorf("type not found: %s", *field.TypeName)
		}
		defaultValue = fmt.Sprintf("new %s()", typeName)
		if forObject {
			typeName += "Object"
		}
	default:
		return "", "", false, errors.New("invalid type")
	}

	if isRepeated {
		typeName = typeName + "[]"
		defaultValue = "[]"
	}
	if isOptional {
		defaultValue = "null"
	}

	return typeName, defaultValue, isOptional, nil
}

func genEnum(enum *descriptorpb.EnumDescriptorProto, parents []*descriptorpb.DescriptorProto, u *typescriptFile) error {
	u.Body.PI("export enum %s {", toLocalClassName(parents, *enum.Name))
	for _, v := range enum.Value {
		u.Body.P("%s = %d,", *v.Name, *v.Number)
	}
	u.Body.PD("}")
	u.Body.P("")
	return nil
}

func genOneofEnum(oneof *descriptorpb.OneofDescriptorProto, fields []*descriptorpb.FieldDescriptorProto, parents []*descriptorpb.DescriptorProto, u *typescriptFile) error {
	typeName := toLocalClassName(parents, internal.ToUpperCamel(*oneof.Name)) + "Case"
	u.Body.PI("export enum %s {", typeName)
	u.Body.P("NOT_SET = 0,")
	for _, field := range fields {
		u.Body.P("k%s = %d,", internal.ToUpperCamel(*field.Name), *field.Number)
	}
	u.Body.PD("}")
	u.Body.P("")
	return nil
}

func genOneof(oneof *descriptorpb.OneofDescriptorProto, fields []*descriptorpb.FieldDescriptorProto, pkg *string, pkgInfo *pkgInfo, parents []*descriptorpb.DescriptorProto, u *typescriptFile) error {
	typeName := toLocalClassName(parents, internal.ToUpperCamel(*oneof.Name)) + "Case"
	fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
	u.Body.P("%s: %s = %s.NOT_SET;", fieldName, typeName, typeName)
	u.Body.PI("clear%s() {", internal.ToUpperCamel(*oneof.Name))
	u.Body.P("this.%s = %s.NOT_SET;", fieldName, typeName)
	for _, field := range fields {
		_, defaultValue, _, err := toTypeName(pkg, pkgInfo, field, false)
		if err != nil {
			return err
		}
		u.Body.P("this.%s = %s;", *field.Name, defaultValue)
	}
	u.Body.PD("}")
	for _, field := range fields {
		fieldTypeName, _, _, err := toTypeName(pkg, pkgInfo, field, false)
		if err != nil {
			return err
		}
		u.Body.PI("set%s(value: %s) {", internal.ToUpperCamel(*field.Name), fieldTypeName)
		u.Body.P("this.%s = %s.k%s;", fieldName, typeName, internal.ToUpperCamel(*field.Name))
		u.Body.P("this.%s = value;", *field.Name)
		u.Body.PD("}")
		u.Body.PI("clear%s() {", internal.ToUpperCamel(*field.Name))
		u.Body.PI("if (this.%s === %s.k%s) {", fieldName, typeName, internal.ToUpperCamel(*field.Name))
		u.Body.P("this.clear%s();", internal.ToUpperCamel(*oneof.Name))
		u.Body.PD("}")
		u.Body.PD("}")
	}
	return nil
}

func genDescriptor(desc *descriptorpb.DescriptorProto, pkg *string, pkgInfo *pkgInfo, parents []*descriptorpb.DescriptorProto, u *typescriptFile) error {
	for _, nested := range desc.NestedType {
		if err := genDescriptor(nested, pkg, pkgInfo, append(parents, desc), u); err != nil {
			return err
		}
	}
	for i, oneof := range desc.OneofDecl {
		fields := getOneofFields(desc.Field, i)
		if len(fields) != 0 {
			if err := genOneofEnum(oneof, fields, append(parents, desc), u); err != nil {
				return err
			}
		}
	}

	localClassName := toLocalClassName(parents, *desc.Name)
	u.Body.PI("export type %sObject = {", localClassName)
	for _, field := range desc.Field {
		typeName, _, isOptional, err := toTypeName(pkg, pkgInfo, field, true)
		if err != nil {
			return err
		}
		fieldName := *field.Name
		if isOptional {
			u.Body.P("%s?: %s | null;", fieldName, typeName)
		} else {
			u.Body.P("%s?: %s;", fieldName, typeName)
		}

		//if oneof := field.OneofIndex; oneof != nil {
		//	oneofTypeName := *desc.OneofDecl[*oneof].Name + "Case"
		//	oneofFieldName := *desc.OneofDecl[*oneof].Name + "_case"
		//	u.Body.P("%s?: %s;", oneofFieldName, oneofTypeName)
		//}
	}
	for i, oneof := range desc.OneofDecl {
		// optional の oneof はフィールドを生成しない
		isOptional := false
		for _, field := range desc.Field {
			if field.OneofIndex != nil && *field.OneofIndex == int32(i) {
				if field.Proto3Optional != nil && *field.Proto3Optional {
					isOptional = true
				}
			}
		}
		if isOptional {
			continue
		}

		typeName := toLocalClassName(append(parents, desc), internal.ToUpperCamel(*oneof.Name)) + "Case"
		fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
		u.Body.P("%s?: %s;", fieldName, typeName)
	}
	u.Body.PD("}")
	u.Body.P("")

	for _, enum := range desc.EnumType {
		if err := genEnum(enum, append(parents, desc), u); err != nil {
			return err
		}
	}

	u.Body.PI("export class %s {", localClassName)
	for _, field := range desc.Field {
		typeName, defaultValue, isOptional, err := toTypeName(pkg, pkgInfo, field, false)
		if err != nil {
			return err
		}
		if isOptional {
			u.Body.P("%s: %s | null = %s;", *field.Name, typeName, defaultValue)
		} else {
			u.Body.P("%s: %s = %s;", *field.Name, typeName, defaultValue)
		}
	}
	for i, oneof := range desc.OneofDecl {
		fields := getOneofFields(desc.Field, i)
		if len(fields) != 0 {
			if err := genOneof(oneof, fields, pkg, pkgInfo, append(parents, desc), u); err != nil {
				return err
			}
		}
	}

	// constructor
	u.Body.PI("constructor(obj: %sObject = {}) {", localClassName)
	for _, field := range desc.Field {
		u.Body.PI("if (obj.%s !== undefined) {", *field.Name)

		typeName, _, isOptional, err := toTypeName(pkg, pkgInfo, field, false)
		if err != nil {
			return err
		}
		if isOptional {
			u.Body.PI("if (obj.%s !== null) {", *field.Name)
		}
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		isMessage := *field.Type == descriptorpb.FieldDescriptorProto_TYPE_GROUP || *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
		if isRepeated && isMessage {
			// isRepeated なので typeName の後ろ２文字は確実に [] となるはず
			elementType := typeName[:len(typeName)-2]
			u.Body.P("this.%s = obj.%s.map((x) => %s.fromObject(x));", *field.Name, *field.Name, elementType)
		} else if !isRepeated && isMessage {
			u.Body.P("this.%s = %s.fromObject(obj.%s);", *field.Name, typeName, *field.Name)
		} else {
			u.Body.P("this.%s = obj.%s;", *field.Name, *field.Name)
		}
		if isOptional {
			u.Body.PD("}")
		}
		u.Body.PD("}")
	}
	for i, oneof := range desc.OneofDecl {
		fields := getOneofFields(desc.Field, i)
		if len(fields) != 0 {
			fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
			u.Body.PI("if (obj.%s !== undefined) {", fieldName)
			u.Body.P("this.%s = obj.%s;", fieldName, fieldName)
			u.Body.PD("}")
		}
	}
	u.Body.PD("}")

	// getType
	u.Body.PI("getType(): typeof %s {", localClassName)
	u.Body.P("return %s;", localClassName)
	u.Body.PD("}")

	// fromJson
	u.Body.PI("static fromJson(json: string): %s {", localClassName)
	u.Body.P("return %s.fromObject(JSON.parse(json));", localClassName)
	u.Body.PD("}")

	// toJson
	u.Body.PI("toJson(): string {")
	u.Body.P("return JSON.stringify(this.toObject());")
	u.Body.PD("}")

	// fromObject
	u.Body.PI("static fromObject(obj: %sObject): %s {", localClassName, localClassName)
	u.Body.P("return new %s(obj);", localClassName)
	u.Body.PD("}")

	// toObject
	u.Body.PI("toObject(): %sObject {", localClassName)
	u.Body.PI("return {")
	for _, field := range desc.Field {
		isRepeated := *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		isMessage := *field.Type == descriptorpb.FieldDescriptorProto_TYPE_GROUP || *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
		isOptional := field.Proto3Optional != nil && *field.Proto3Optional
		if isOptional {
			if isRepeated && isMessage {
				u.Body.P("%s: this.%s === null ? null : this.%s.map((x) => x.toObject()),", *field.Name, *field.Name, *field.Name)
			} else if !isRepeated && isMessage {
				u.Body.P("%s: this.%s === null ? null : this.%s.toObject(),", *field.Name, *field.Name, *field.Name)
			} else {
				u.Body.P("%s: this.%s,", *field.Name, *field.Name)
			}
		} else {
			if isRepeated && isMessage {
				u.Body.P("%s: this.%s.map((x) => x.toObject()),", *field.Name, *field.Name)
			} else if !isRepeated && isMessage {
				u.Body.P("%s: this.%s.toObject(),", *field.Name, *field.Name)
			} else {
				u.Body.P("%s: this.%s,", *field.Name, *field.Name)
			}
		}
	}
	for i, oneof := range desc.OneofDecl {
		fields := getOneofFields(desc.Field, i)
		if len(fields) != 0 {
			fieldName := internal.ToSnakeCase(*oneof.Name) + "_case"
			u.Body.P("%s: this.%s,", fieldName, fieldName)
		}
	}
	u.Body.PD("};")
	u.Body.PD("}")

	// 	if oneof := field.OneofIndex; oneof != nil {
	// 		oneofTypeName := internal.ToUpperCamel(*desc.OneofDecl[*oneof].Name) + "Case"
	// 		oneofFieldName := internal.ToSnakeCase(*desc.OneofDecl[*oneof].Name) + "_case"
	// 		u.Body.P("public void Set%s(%s %s)", internal.ToUpperCamel(fieldName), typeName, fieldName)
	// 		u.Body.PI("{")
	// 		u.Body.P("Clear%s();", oneofTypeName)
	// 		u.Body.P("%s = %s.k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
	// 		u.Body.P("this.%s = %s;", fieldName, fieldName)
	// 		u.Body.PD("}")
	// 		u.Body.P("public bool Has%s()", internal.ToUpperCamel(fieldName))
	// 		u.Body.PI("{")
	// 		u.Body.P("return %s == %s.k%s;", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
	// 		u.Body.PD("}")
	// 		u.Body.P("public void Clear%s()", internal.ToUpperCamel(fieldName))
	// 		u.Body.PI("{")
	// 		u.Body.P("if (%s == %s.k%s)", oneofFieldName, oneofTypeName, internal.ToUpperCamel(fieldName))
	// 		u.Body.PI("{")
	// 		u.Body.P("Clear%s();", oneofTypeName)
	// 		u.Body.PD("}")
	// 		u.Body.PD("}")
	// 	}
	// }

	u.Body.PD("}")
	u.Body.P("")

	return nil
}

func genFile(file *descriptorpb.FileDescriptorProto, files []*descriptorpb.FileDescriptorProto, pkgInfo *pkgInfo) (*pluginpb.CodeGeneratorResponse_File, error) {
	u := typescriptFile{}
	u.Top.SetIndentUnit(4)
	u.Bottom.SetIndentUnit(4)
	u.Body.SetIndentUnit(4)

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
		alias, ok := pkgInfo.filenameToAlias[dep]
		if !ok {
			return nil, errors.New("filename not found")
		}
		u.Top.P("import * as %s from \"./%s\";", alias, fileName)
	}
	u.Top.P("")

	for _, enum := range file.EnumType {
		if err := genEnum(enum, nil, &u); err != nil {
			return nil, err
		}
	}
	for _, desc := range file.MessageType {
		if err := genDescriptor(desc, file.Package, pkgInfo, nil, &u); err != nil {
			return nil, err
		}
	}

	// 拡張子を取り除いて .ts を付ける
	fileName := *file.Name
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
	fileName = fileName + ".ts"

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

	f.PI("export interface Jsonif<T> {")
	f.P("getType: () => { fromJson(json: string): T };")
	f.P("toJson: () => string;")
	f.PD("}")
	f.P("")
	f.PI("export function getType<T extends number | string | boolean | Jsonif<T>>(v: T): any {")
	f.PI("if ((v as any).getType !== undefined) {")
	f.P("return (v as any).getType();")
	f.PDI("} else {")
	f.P("return v.constructor as any;")
	f.PD("}")
	f.PD("}")
	f.P("")
	f.PI("export function fromJson<T>(v: string, type: any): T {")
	f.PI("if (type.fromJson !== undefined) {")
	f.P("return type.fromJson(v) as T;")
	f.PDI("} else {")
	f.P("return JSON.parse(v) as T;")
	f.PD("}")
	f.PD("}")
	f.P("")
	f.PI("export function toJson<T extends number | string | boolean | Jsonif<T>>(v: T): string {")
	f.PI("if (typeof v === 'number' || typeof v === 'string' || typeof v === 'boolean') {")
	f.P("return JSON.stringify(v);")
	f.PDI("} else {")
	f.P("return v.toJson();")
	f.PD("}")
	f.PD("}")

	fileName := "jsonif.ts"

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
	pkgInfo := newPkgInfo()
	for _, file := range req.ProtoFile {
		pkgInfo.addFile(file)
	}

	for _, file := range req.ProtoFile {
		respFile, err := genFile(file, req.ProtoFile, pkgInfo)
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
