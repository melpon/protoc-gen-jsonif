package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

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

	// proto2 ファイルがあったらエラーにする
	for _, file := range req.ProtoFile {
		if file.Syntax == nil {
			return errors.New("syntax not specified. Supported syntax=proto3 only.")
		}
		if *file.Syntax != "proto3" {
			return errors.New(fmt.Sprintf("syntax=%s not supported. Supported syntax=proto3 only.", *file.Syntax))
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
