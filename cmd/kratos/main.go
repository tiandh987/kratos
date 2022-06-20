package main

import (
	"log"

	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/change"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/project"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/proto"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/run"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/upgrade"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "kratos",
	Short:   "Kratos: An elegant toolkit for Go microservices.",
	Long:    `Kratos: An elegant toolkit for Go microservices.`,
	Version: release,
}

func init() {
	// kratos new [flags]
	//	创建项目
	// 		-r  --repo-url (默认: https://gitee.com/go-kratos/kratos-layout.git)
	//		-b  --branch   (默认: @main)
	//		-t  --timeout  (默认: 60s)
	//			--nomod
	// eg : kratos new helloworld -r https://gitee.com/go-kratos/kratos-layout.git
	rootCmd.AddCommand(project.CmdNew)

	// kratos proto [flags]
	//		未实现什么功能
	// kratos proto [command]
	// 		kratos proto add [flags]
	// 		kratos proto client [flags]
	// 		kratos proto server [flags]
	rootCmd.AddCommand(proto.CmdProto)
	rootCmd.AddCommand(upgrade.CmdUpgrade)
	rootCmd.AddCommand(change.CmdChange)

	// kratos run [flags]
	//	运行项目
	rootCmd.AddCommand(run.CmdRun)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
