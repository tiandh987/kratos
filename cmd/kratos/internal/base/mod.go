package base

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ModulePath returns go module path.
// 返回 go module 路径
// eg： module test/rpc/client （go.mod 文件，第一行 module 关键字）
//	返回 test/rpc/client
func ModulePath(filename string) (string, error) {
	// ReadFile 读取指定文件并返回内容。
	modBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// 返回 go.mod 文件，第一行 module 关键字后的 字符串
	return modfile.ModulePath(modBytes), nil
}

// ModuleVersion returns module version.
func ModuleVersion(path string) (string, error) {
	stdout := &bytes.Buffer{}

	// go mod graph
	//	Graph 以文本形式打印模块需求图。
	//	输出中的每一行都有两个以 "空格" 分隔的字段：
	//		一个模块及其要求之一。
	//	每个模块都被标识为 "路径@version" 形式的字符串，但主模块除外，它没有@version 后缀。
	fd := exec.Command("go", "mod", "graph")
	fd.Stdout = stdout
	fd.Stderr = stdout
	if err := fd.Run(); err != nil {
		return "", err
	}
	rd := bufio.NewReader(stdout)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			return "", err
		}
		str := string(line)
		i := strings.Index(str, "@")
		if strings.Contains(str, path+"@") && i != -1 {
			// line 中包含 @, 且包含 path + "@"
			// eg: path = github.com/go-kratos/kratos/v2
			//	返回 github.com/go-kratos/kratos/v2@v2.3.1
			return path + str[i:], nil
		}
	}
}

// KratosMod returns kratos mod.
func KratosMod() string {
	// go 1.15+ read from env GOMODCACHE
	// go env GOMODCACHE
	//	/home/tian/workspace/golang/pkg/mod
	cacheOut, _ := exec.Command("go", "env", "GOMODCACHE").Output()
	cachePath := strings.Trim(string(cacheOut), "\n")

	// go env GOPATH
	//	/home/tian/workspace/golang
	pathOut, _ := exec.Command("go", "env", "GOPATH").Output()
	gopath := strings.Trim(string(pathOut), "\n")

	// 如果 cachePath 为 "", 则为 ${GOPATH}/pkg/mod
	if cachePath == "" {
		// ${GOPATH}/pkg/mod
		cachePath = filepath.Join(gopath, "pkg", "mod")
	}

	// 返回 github.com/go-kratos/kratos/v2@v2.3.1
	if path, err := ModuleVersion("github.com/go-kratos/kratos/v2"); err == nil {
		// $GOPATH/pkg/mod/github.com/go-kratos/kratos@v2.3.1
		return filepath.Join(cachePath, path)
	}

	// $GOPATH/src/github.com/go-kratos/kratos
	return filepath.Join(gopath, "src", "github.com", "go-kratos", "kratos")
}
