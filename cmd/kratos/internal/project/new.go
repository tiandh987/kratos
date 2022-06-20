package project

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/base"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// Project is a project template.
// eg : kratos new helloworld -r https://gitee.com/go-kratos/kratos-layout.git
type Project struct {
	// 项目名    helloworld
	Name string
	// 项目路径  ""
	Path string
}

// New new a project from remote repo.
// eg : kratos new helloworld -r https://gitee.com/go-kratos/kratos-layout.git
// 	dir    : /home/tian/workspace/golang/src/kratos (执行命令时所在的当前路径)
//	layout : 远程仓库地址
//  branch : 分支名
func (p *Project) New(ctx context.Context, dir string, layout string, branch string) error {
	// to : /home/tian/workspace/golang/src/kratos/helloworld
	to := path.Join(dir, p.Name)

	// 若 to 目录存在, 可以进行覆盖
	if _, err := os.Stat(to); !os.IsNotExist(err) {
		fmt.Printf("🚫 %s already exists\n", p.Name)
		override := false
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		e := survey.AskOne(prompt, &override)
		if e != nil {
			return e
		}
		if !override {
			return err
		}
		os.RemoveAll(to)
	}

	fmt.Printf("🚀 Creating service %s, layout repo is %s, please wait a moment.\n\n", p.Name, layout)

	repo := base.NewRepo(layout, branch)
	if err := repo.CopyTo(ctx, to, p.Path, []string{".git", ".github"}); err != nil {
		return err
	}

	// mv ${to}/cmd/server ${to}/cmd/helloworld
	e := os.Rename(
		path.Join(to, "cmd", "server"),
		path.Join(to, "cmd", p.Name),
	)
	if e != nil {
		return e
	}

	// 命令行打印创建的文件
	base.Tree(to, dir)

	// 命令行提示：项目创建成功
	fmt.Printf("\n🍺 Project creation succeeded %s\n", color.GreenString(p.Name))
	// 命令行提示：使用下面的命令启动项目
	fmt.Print("💻 Use the following command to start the project 👇:\n\n")

	// 命令行提示：cd helloworld
	fmt.Println(color.WhiteString("$ cd %s", p.Name))
	// 命令行提示：go generate ./...
	fmt.Println(color.WhiteString("$ go generate ./..."))
	// 命令行提示：go build -o ./bin/ ./...
	fmt.Println(color.WhiteString("$ go build -o ./bin/ ./... "))
	// 命令行提示：./bin/helloworld -conf ./configs
	fmt.Println(color.WhiteString("$ ./bin/%s -conf ./configs\n", p.Name))

	// 命令行提示：感谢使用 Kratos
	fmt.Println("			🤝 Thanks for using Kratos")
	fmt.Println("	📚 Tutorial: https://go-kratos.dev/docs/getting-started/start")
	return nil
}
