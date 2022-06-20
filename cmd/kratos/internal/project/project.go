package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	// survey 交互式命令行工具库
	"github.com/AlecAivazis/survey/v2"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/base"
	"github.com/spf13/cobra"
)

// CmdNew represents the new command.
// kratos new 命令: 使用仓库模板, 创建一个服务项目
var CmdNew = &cobra.Command{
	Use:   "new",
	Short: "Create a service template",
	Long:  "Create a service project using the repository template. Example: kratos new helloworld",
	Run:   run,
}

var (
	repoURL string  // -r 指定模板仓库源  1、kratos new 项目名 -r 仓库地址  2、环境变量  3、默认值（https://github.com/go-kratos/kratos-layout.git）
	branch  string  // -b 指定分支
	timeout string  // -t 创建项目的超时时间
	nomod   bool    // --nomod 添加服务, 共用 go.mod ,大仓模式
)

func init() {
	// 默认模板仓库地址:
	//     优先从 KRATOS_LAYOUT_REPO 环境变量读取;
	//     若未设置环境变量, 默认为: https://github.com/go-kratos/kratos-layout.git
	if repoURL = os.Getenv("KRATOS_LAYOUT_REPO"); repoURL == "" {
		repoURL = "https://github.com/go-kratos/kratos-layout.git"
	}
	timeout = "60s"
	CmdNew.Flags().StringVarP(&repoURL, "repo-url", "r", repoURL, "layout repo")
	CmdNew.Flags().StringVarP(&branch, "branch", "b", branch, "repo branch")
	CmdNew.Flags().StringVarP(&timeout, "timeout", "t", timeout, "time out")
	CmdNew.Flags().BoolVarP(&nomod, "nomod", "", nomod, "retain go mod")
}

func run(cmd *cobra.Command, args []string) {
	// Getwd 返回与当前目录对应的根路径名
	// 即: 运行 kratos new 所在当前目录
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// 解析超时时间
	t, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}

	// 创建带有 timeout 的上下文
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	// 用于存储项目名
	// eg: kratos new helloworld
	// 	name = "helloworld"
	name := ""

	// kratos new
	if len(args) == 0 {
		prompt := &survey.Input{
			Message: "What is project name ?",
			Help:    "Created project name.",
		}
		err = survey.AskOne(prompt, &name)
		if err != nil || name == "" {
			return
		}
	} else {
		// kratos new helloworld
		name = args[0]
	}

	// 初始化 Project 结构体
	// 	Base 获取路径的最后一个元素
	//	项目名, 项目路径
	p := &Project{Name: path.Base(name), Path: name}

	// 用于接收 error 的 channel
	done := make(chan error, 1)

	go func() {
		// nomod = false
		if !nomod {
			done <- p.New(ctx, wd, repoURL, branch)
			return
		}

		// nomod = true
		// 检查当前目录下是否有 go.mod, 如果不存在则报错
		if _, e := os.Stat(path.Join(wd, "go.mod")); os.IsNotExist(e) {
			done <- fmt.Errorf("🚫 go.mod don't exists in %s", wd)
			return
		}

		mod, e := base.ModulePath(path.Join(wd, "go.mod"))
		if e != nil {
			panic(e)
		}
		done <- p.Add(ctx, wd, repoURL, branch, mod)
	}()
	select {
	case <-ctx.Done():
		// 创建项目超时
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprint(os.Stderr, "\033[31mERROR: project creation timed out\033[m\n")
			return
		}
		// 创建项目失败
		fmt.Fprintf(os.Stderr, "\033[31mERROR: failed to create project(%s)\033[m\n", ctx.Err().Error())
	case err = <-done:
		// 创建项目过程中出错
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: Failed to create project(%s)\033[m\n", err.Error())
		}
	}
}
