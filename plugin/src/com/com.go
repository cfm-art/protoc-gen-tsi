package com

import (
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// ReadFrom is 入力からバイナリを読み込んでプロトコルバッファへ
func ReadFrom(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req plugin.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// WriteTo is プロトコルバッファを出力へ書き込み
func WriteTo(res *plugin.CodeGeneratorResponse, w io.Writer) error {
	buf, err := proto.Marshal(res)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

// Option is コマンドライン引数
type Option struct {
	GenClient  bool
	ClientType string
	Nonull     bool
	DupArray   bool
}

// ParseArgument is 渡されたパラメータをいい加減に解析
func ParseArgument(req *plugin.CodeGeneratorRequest) Option {
	result := Option{true, "fetch", false, true}
	if req.Parameter != nil {
		for _, p := range strings.Split(req.GetParameter(), ",") {
			tokens := strings.SplitN(p, "=", 2)
			if len(tokens) == 2 {
				key, value := tokens[0], tokens[1]
				if key == "client" {
					if value == "true" {
						result.GenClient = true
					}
				} else if key == "clientType" {
					result.ClientType = value
					if value != "fetch" && value != "ajax" {
						log.Fatalf("Invalid ClientType : " + value)
					}
				} else if key == "nonull" {
					if value == "true" {
						result.Nonull = true
					}
				} else if key == "duparray" {
					if value == "true" {
						result.DupArray = true
					}
				}
			}
		}
	}
	return result
}
