package project

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/base"
)

var repoAddIgnores = []string{
	".git", ".github", "api", "README.md", "LICENSE", "go.mod", "go.sum", "third_party",
}

// dir    : /home/tian/workspace/golang/src/kratos
// layout : https://github.com/go-kratos/kratos-layout.git
// branch :
// mod    :
func (p *Project) Add(ctx context.Context, dir string, layout string, branch string, mod string) error {
	// to : /home/tian/workspace/golang/src/kratos/helloworld
	to := path.Join(dir, p.Path)

	// 如果 to 目录(执行 kratos new 是所在目录 + 项目路径)已经存在,可以选择覆盖(先删除)
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

	fmt.Printf("🚀 Add service %s, layout repo is %s, please wait a moment.\n\n", p.Name, layout)

	// 创建一个 git 仓库管理者
	repo := base.NewRepo(layout, branch)

	if err := repo.CopyToV2(ctx, to, path.Join(mod, p.Path), repoAddIgnores, []string{path.Join(p.Path, "api"), "api"}); err != nil {
		return err
	}

	e := os.Rename(
		path.Join(to, "cmd", "server"),
		path.Join(to, "cmd", p.Name),
	)
	if e != nil {
		return e
	}

	base.Tree(to, dir)

	fmt.Printf("\n🍺 Repository creation succeeded %s\n", color.GreenString(p.Name))
	fmt.Print("💻 Use the following command to add a project 👇:\n\n")

	fmt.Println(color.WhiteString("$ cd %s", p.Name))
	fmt.Println(color.WhiteString("$ go generate ./..."))
	fmt.Println(color.WhiteString("$ go build -o ./bin/ ./... "))
	fmt.Println(color.WhiteString("$ ./bin/%s -conf ./configs\n", p.Name))
	fmt.Println("			🤝 Thanks for using Kratos")
	fmt.Println("	📚 Tutorial: https://go-kratos.dev/docs/getting-started/start")
	return nil
}
