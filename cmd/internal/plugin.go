package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func setRemoveCandidates(m map[string]bool, files []*descriptorpb.FileDescriptorProto, file *descriptorpb.FileDescriptorProto) {
	for _, name := range file.Dependency {
		m[name] = true
		var found *descriptorpb.FileDescriptorProto
		for _, f := range files {
			if *f.Name == name {
				found = f
				break
			}
		}
		// 依存が依存してる先も削除候補として設定する
		setRemoveCandidates(m, files, found)
	}
}

func isWeakDeps(i int, file *descriptorpb.FileDescriptorProto) bool {
	for _, v := range file.WeakDependency {
		if int(v) == i {
			return true
		}
	}
	return false
}

func setDeps(m map[string]*Dep, files []*descriptorpb.FileDescriptorProto, file *descriptorpb.FileDescriptorProto) {
	for i, name := range file.Dependency {
		m[name].Refed = true

		if isWeakDeps(i, file) {
			continue
		}

		m[name].Ref += 1

		var found *descriptorpb.FileDescriptorProto
		for _, f := range files {
			if *f.Name == name {
				found = f
				break
			}
		}
		// 依存が依存してる先も参照を設定する
		setDeps(m, files, found)
	}
}

type Dep struct {
	// weak でない参照されている数
	Ref int
	// weak の有無にかかわらず、とにかく参照されたことがあるかどうか
	Refed   bool
	Removed bool
}

type PluginFunc = func(*pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)

func RunPlugin(gen PluginFunc) error {
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(in, req); err != nil {
		return err
	}

	// weak なインポートだったら依存先を見ないようにする

	m := make(map[string]*Dep)
	for _, file := range req.ProtoFile {
		m[*file.Name] = &Dep{
			Ref:     0,
			Refed:   false,
			Removed: false,
		}
	}
	for _, file := range req.ProtoFile {
		setDeps(m, req.ProtoFile, file)
	}

	// 削除しても良いファイルを削除対象に入れて、その依存先の数も減らして、変化が無くなるまで繰り返す
	for {
		changed := false
		for _, file := range req.ProtoFile {
			dep := m[*file.Name]
			if !dep.Removed && dep.Ref == 0 && dep.Refed {
				changed = true

				// このファイルは削除して良い
				dep.Removed = true
				// 依存先の参照を減らす
				for i, name := range file.Dependency {
					if isWeakDeps(i, file) {
						continue
					}
					m[name].Ref -= 1
				}
			}
		}

		if !changed {
			break
		}
	}

	// 削除する
	var files []*descriptorpb.FileDescriptorProto
	for _, file := range req.ProtoFile {
		dep := m[*file.Name]
		// 参照されてないファイルは削除する
		if dep.Removed {
			//fmt.Fprintf(os.Stderr, "Removed %s for weak dependendency or not referenced\n", *file.Name)
			continue
		}
		// enum, message の定義が１個も無い場合も削除する
		if len(file.MessageType) == 0 && len(file.EnumType) == 0 {
			//fmt.Fprintf(os.Stderr, "Removed %s for empty definition\n", *file.Name)
			continue
		}
		files = append(files, file)
	}
	req.ProtoFile = files

	// proto2 ファイルがあったらエラーにする
	for _, file := range req.ProtoFile {
		if file.Syntax == nil {
			return errors.New(fmt.Sprintf("%s: syntax not specified. Supported syntax=proto3 only.", *file.Name))
		}
		if *file.Syntax != "proto3" {
			return errors.New(fmt.Sprintf("%s: syntax=%s not supported. Supported syntax=proto3 only.", *file.Name, *file.Syntax))
		}
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
