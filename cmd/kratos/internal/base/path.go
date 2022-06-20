package base

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// 返回 kratos Home 目录
// eg: 当前 linux 用户 tian
//	返回 /home/tian/.kratos
func kratosHome() string {
	// 获取当前用户的 home 目录
	// eg: /home/tian
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// /home/tian/.kratos
	home := path.Join(dir, ".kratos")
	// 不存在则进行创建
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

// 返回 : ${当前用户名 Home 目录}/.kratos/${dir}
func kratosHomeWithDir(dir string) string {
	// kratosHome() : /home/tian/.kratos
	// dir : repo/gitee.com/go-kratos
	// home : /home/tian/.kratos/repo/gitee.com/go-kratos
	home := path.Join(kratosHome(), dir)
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func copyFile(src, dst string, replaces []string) error {
	var err error
	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	buf, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	var old string
	for i, next := range replaces {
		if i%2 == 0 {
			old = next
			continue
		}
		buf = bytes.ReplaceAll(buf, []byte(old), []byte(next))
	}
	return os.WriteFile(dst, buf, srcinfo.Mode())
}

// src      : /home/tian/.kratos/repo/gitee.com/go-kratos/kratos-layout@main
//		kratos new -r 指定源, 在本地进行缓存的路径
// dst      : /home/tian/workspace/golang/src/kratos/helloworld
//		项目创建的路径(取决于在那个目录下执行 kratos 命令)
// replaces : []string{"github.com/go-kratos/kratos-layout", ""}
//
// ignores  : []string{".git", ".github"}
//		仓库缓存目录下被忽略的文件
func copyDir(src, dst string, replaces, ignores []string) error {
	var err error
	var fds []os.DirEntry
	var srcinfo os.FileInfo

	// 获取仓库缓存目录的信息
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	// 创建 项目目录
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	// 读取仓库缓存目录
	if fds, err = os.ReadDir(src); err != nil {
		return err
	}

	// 遍历仓库缓存目录下的文件
	for _, fd := range fds {
		// 如果被忽略, 则跳过
		if hasSets(fd.Name(), ignores) {
			continue
		}

		// 递归拷贝该目录下的文件夹、文件
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())
		var e error
		if fd.IsDir() {
			e = copyDir(srcfp, dstfp, replaces, ignores)
		} else {
			e = copyFile(srcfp, dstfp, replaces)
		}
		if e != nil {
			return e
		}
	}
	return nil
}

func hasSets(name string, sets []string) bool {
	for _, ig := range sets {
		if ig == name {
			return true
		}
	}
	return false
}

func Tree(path string, dir string) {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			fmt.Printf("%s %s (%v bytes)\n", color.GreenString("CREATED"), strings.Replace(path, dir+"/", "", -1), info.Size())
		}
		return nil
	})
}
