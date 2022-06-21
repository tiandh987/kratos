package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/base"

	"github.com/spf13/cobra"
)

// CmdClient represents the source command.
var CmdClient = &cobra.Command{
	Use:   "client",
	Short: "Generate the proto client code",
	Long:  "Generate the proto client code. Example: kratos proto client helloworld.proto",
	Run:   run,
}

var protoPath string

func init() {
	if protoPath = os.Getenv("KRATOS_PROTO_PATH"); protoPath == "" {
		protoPath = "./third_party"
	}
	CmdClient.Flags().StringVarP(&protoPath, "proto_path", "p", protoPath, "proto path")
}

func run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		// 执行 kratos proto client 命令, 必须指定 proto 文件
		fmt.Println("Please enter the proto file or directory")
		return
	}

	var (
		err   error
		proto = strings.TrimSpace(args[0])
	)

	// 查看  "protoc-gen-go",
	//		"protoc-gen-go-grpc",
	//		"protoc-gen-go-http",
	//		"protoc-gen-go-errors",
	//		"protoc-gen-openapi"
	//  是否已经安装。
	if err = look("protoc-gen-go", "protoc-gen-go-grpc", "protoc-gen-go-http", "protoc-gen-go-errors", "protoc-gen-openapi"); err != nil {
		// update the kratos plugins
		// 若没有安装，则执行 kratos upgrade 进行更新
		cmd := exec.Command("kratos", "upgrade")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			fmt.Println(err)
			return
		}
	}

	// 判断提供的是否为 .proto 文件
	if strings.HasSuffix(proto, ".proto") {
		// 基于 .proto 文件生成客户端代码
		err = generate(proto, args)
	} else {
		// 给定的是目录,对目录下所有 proto 文件生成客户端代码
		err = walk(proto, args)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func look(name ...string) error {
	for _, n := range name {
		// LookPath 在 PATH 环境变量命名的目录中搜索可执行的命名文件。
		// 如果文件包含斜杠，则直接尝试，不参考 PATH。
		if _, err := exec.LookPath(n); err != nil {
			return err
		}
	}
	return nil
}

func walk(dir string, args []string) error {
	if dir == "" {
		dir = "."
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ext := filepath.Ext(path); ext != ".proto" || strings.HasPrefix(path, "third_party") {
			return nil
		}
		return generate(path, args)
	})
}

// generate is used to execute the generate command for the specified proto file
// 对指定的 proto 文件执行 generate 命令
func generate(proto string, args []string) error {
	input := []string{
		"--proto_path=.",
	}
	if pathExists(protoPath) {
		input = append(input, "--proto_path="+protoPath)
	}
	inputExt := []string{
		// base.KratosMod() : $GOPATH/pkg/mod/github.com/go-kratos/kratos/v2@v2.3.1
		"--proto_path=" + base.KratosMod(),
		// $GOPATH/pkg/mod/github.com/go-kratos/kratos/v2@v2.3.1/third_party
		"--proto_path=" + filepath.Join(base.KratosMod(), "third_party"),
		"--go_out=paths=source_relative:.",
		"--go-grpc_out=paths=source_relative:.",
		"--go-http_out=paths=source_relative:.",
		"--go-errors_out=paths=source_relative:.",
		"--openapi_out=paths=source_relative:.",
	}
	input = append(input, inputExt...)
	protoBytes, err := os.ReadFile(proto)
	if err == nil && len(protoBytes) > 0 {
		if ok, _ := regexp.Match(`\n[^/]*(import)\s+"validate/validate.proto"`, protoBytes); ok {
			input = append(input, "--validate_out=lang=go,paths=source_relative:.")
		}
	}
	input = append(input, proto)
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			input = append(input, a)
		}
	}
	fd := exec.Command("protoc", input...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	fd.Dir = "."
	if err := fd.Run(); err != nil {
		return err
	}
	fmt.Printf("proto: %s\n", proto)
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
