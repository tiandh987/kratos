package base

import (
	"context"
	"fmt"
	stdurl "net/url"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Repo is git repository manager.
// git 仓库管理者
type Repo struct {
	// 仓库地址 eg: https://gitee.com/go-kratos/kratos-layout.git
	url    string

	// /home/tian/.kratos/repo/gitee.com/go-kratos (${当前用户Home目录}/.kratos/repo/${指定源目录})
	home   string

	// 分支, 默认为 main 分支
	branch string
}

// 返回项目的目录
// eg:
//	url  : https://gitee.com/go-kratos/kratos-layout.git
//  返回 : gitee.com/go-kratos
func repoDir(url string) string {
	// 如果 url 中没有 // , 则进行添加
	if !strings.Contains(url, "//") {
		url = "//" + url
	}

	// 添加使用的协议 ssh / https
	if strings.HasPrefix(url, "//git@") {
		url = "ssh:" + url
	} else if strings.HasPrefix(url, "//") {
		url = "https:" + url
	}

	// 解析 url
	u, err := stdurl.Parse(url)
	if err == nil {
		// eg: https://gitee.com/go-kratos/kratos-layout.git
		// 	u.Scheme       https
		//  u.Hostname()   gitee.com
		//  u.Path         /go-kratos/kratos-layout.git
		url = fmt.Sprintf("%s://%s%s", u.Scheme, u.Hostname(), u.Path)
	}

	var start int
	start = strings.Index(url, "//")
	if start == -1 {
		start = strings.Index(url, ":") + 1
	} else {
		start += 2
	}
	end := strings.LastIndex(url, "/")

	// eg : https://gitee.com/go-kratos/kratos-layout.git
	// 	start :     8
	//	end	  :                        27
	//	url[start:end] : gitee.com/go-kratos
	return url[start:end]
}

// NewRepo new a repository manager.
// 创建一个仓库管理者
func NewRepo(url string, branch string) *Repo {
	return &Repo{
		// https://gitee.com/go-kratos/kratos-layout.git (kratos new helloworld -r https://gitee.com/go-kratos/kratos-layout.git)
		url:    url,
		// /home/tian/.kratos/repo/gitee.com/go-kratos
		home:   kratosHomeWithDir("repo/" + repoDir(url)),
		branch: branch, // 分支
	}
}

// Path returns the repository cache path.
// 返回 kratos new helloworld -r xxx 指定仓库缓存到本地的路径
// eg:
//	url   : https://gitee.com/go-kratos/kratos-layout.git
//  start :                            27
//  end   :                                          41
//
// 返回: /home/tian/.kratos/repo/gitee.com/go-kratos/kratos-layout@main
func (r *Repo) Path() string {
	// 获取仓库名称: kratos-layout
	start := strings.LastIndex(r.url, "/")
	end := strings.LastIndex(r.url, ".git")
	if end == -1 {
		end = len(r.url)
	}

	// 获取仓库分支, 默认为 @main
	var branch string
	if r.branch == "" {
		branch = "@main"
	} else {
		branch = "@" + r.branch
	}

	// /home/tian/.kratos/repo/gitee.com/go-kratos/kratos-layout@main
	// r.home             : /home/tian/.kratos/repo/gitee.com/go-kratos (Home目录 + .kratos + repo/ + -r指定仓库目录)
	// r.url[start+1:end] : kratos-layout
	// branch             : @main
	return path.Join(r.home, r.url[start+1:end]+branch)
}

// Pull fetch the repository from remote url.
func (r *Repo) Pull(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD")
	cmd.Dir = r.Path()
	_, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	cmd = exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = r.Path()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return err
}

// Clone clones the repository to cache path.
// 克隆仓库到缓存路径 (git pull 或者 git clone)
func (r *Repo) Clone(ctx context.Context) error {
	// 判断仓库缓存目录是否存在
	// Path: 返回仓库的缓存目录
	if _, err := os.Stat(r.Path()); !os.IsNotExist(err) {
		// 存在, 则通过 git pull 命令进行更新
		return r.Pull(ctx)
	}

	// 不存在, 则需要通过 git clone 命令, 将 -r 指定的源克隆到缓存目录
	var cmd *exec.Cmd
	if r.branch == "" {
		cmd = exec.CommandContext(ctx, "git", "clone", r.url, r.Path())
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "-b", r.branch, r.url, r.Path())
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// CopyTo copies the repository to project path.
// 拷贝缓存的源 到 新建的项目目录
func (r *Repo) CopyTo(ctx context.Context, to string, modPath string, ignores []string) error {
	if err := r.Clone(ctx); err != nil {
		return err
	}

	// 输入: /home/tian/.kratos/repo/gitee.com/go-kratos/kratos-layout@main/go.mod
	// 该 mod 文件内容为: module github.com/go-kratos/kratos-layout
	// 所以, mod 为 github.com/go-kratos/kratos-layout
	mod, err := ModulePath(path.Join(r.Path(), "go.mod"))
	if err != nil {
		return err
	}

	return copyDir(r.Path(), to, []string{mod, modPath}, ignores)
}

// CopyToV2 copies the repository to project path
// 拷贝仓库到项目路径
// to       : 新项目创建的目录 (/home/tian/workspace/golang/src/kratos/helloworld)
// modPath  :
// ignores  :
// replaces :
func (r *Repo) CopyToV2(ctx context.Context, to string, modPath string, ignores, replaces []string) error {
	// git clone 或者 git pull  -r 制定的源
	if err := r.Clone(ctx); err != nil {
		return err
	}

	// /home/tian/.kratos/repo/gitee.com/go-kratos/kratos-layout@main/go.mod
	mod, err := ModulePath(path.Join(r.Path(), "go.mod"))
	if err != nil {
		return err
	}
	replaces = append([]string{mod, modPath}, replaces...)
	return copyDir(r.Path(), to, replaces, ignores)
}
