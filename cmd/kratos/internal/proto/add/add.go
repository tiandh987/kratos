package add

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CmdAdd represents the add command.
var CmdAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a proto API template",
	Long:  "Add a proto API template. Example: kratos add helloworld/v1/hello.proto",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	// kratos proto add helloworld/v1/helloworld.proto

	// input: helloworld/v1/helloworld.proto
	input := args[0]
	n := strings.LastIndex(input, "/")
	if n == -1 {
		// proto 路径需要通过 / 分层
		fmt.Println("The proto path needs to be hierarchical.")
		return
	}

	// path : helloworld/v1
	path := input[:n]
	// fileName : helloworld.proto
	fileName := input[n+1:]
	// pkgName : helloworld.v1
	pkgName := strings.ReplaceAll(path, "/", ".")

	p := &Proto{
		Name:        fileName,
		Path:        path,
		Package:     pkgName,
		GoPackage:   goPackage(path),
		JavaPackage: javaPackage(pkgName),
		Service:     serviceName(fileName),
	}
	// 基于 Proto 结构体生成模板
	if err := p.Generate(); err != nil {
		fmt.Println(err)
		return
	}
}

func modName() string {
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
		if modBytes, err = os.ReadFile("../go.mod"); err != nil {
			return ""
		}
	}
	return modfile.ModulePath(modBytes)
}

// eg: kratos proto add helloworld/v1/helloworld.proto
// 	返回: /helloworld/v1;v1
func goPackage(path string) string {
	s := strings.Split(path, "/")
	return modName() + "/" + path + ";" + s[len(s)-1]
}

// eg: kratos proto add helloworld/v1/helloworld.proto
//	返回：helloworld.v1
func javaPackage(name string) string {
	return name
}

// eg:
//	name : helloworld.proto
func serviceName(name string) string {
	// helloworld
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

func toUpperCamelCase(s string) string {
	// 将 s 中 _ 替换成 " "
	s = strings.ReplaceAll(s, "_", " ")
	// 将每个单词首字母大写
	s = cases.Title(language.Und, cases.NoLower).String(s)
	// 将 s 中 " " 替换成 ""
	return strings.ReplaceAll(s, " ", "")
}
