package proto

import (
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/proto/add"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/proto/client"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/proto/server"

	"github.com/spf13/cobra"
)

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
	Run:   run,
}

func init() {
	// 根据模板生成 .proto 文件
	CmdProto.AddCommand(add.CmdAdd)

	// 对给定的 .proto 文件生成客户端代码
	// 对给定的目录, 将目录下 .proto 文件生成客户端代码
	// 	kratos proto client api/helloworld/demo.proto
	CmdProto.AddCommand(client.CmdClient)

	// 通过 proto文件，可以直接生成对应的 Service 实现代码
	//	kratos proto server api/helloworld/demo.proto -t internal/service
	CmdProto.AddCommand(server.CmdServer)
}

func run(cmd *cobra.Command, args []string) {
}
