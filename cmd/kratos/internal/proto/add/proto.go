package add

import (
	"fmt"
	"os"
	"path"
)

// Proto is a proto generator.
// proto 生成器
// eg : kratos proto add helloworld/v1/helloworld.proto
type Proto struct {
	Name        string  // helloworld.proto
	Path        string  // helloworld/v1
	Service     string  // Helloworld
	Package     string  // helloworld.v1
	GoPackage   string  // /helloworld/v1;v1
	JavaPackage string  // helloworld.v1
}

// Generate generate a proto template.
// 生成 proto 模板
func (p *Proto) Generate() error {
	body, err := p.execute()
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	to := path.Join(wd, p.Path)
	if _, err := os.Stat(to); os.IsNotExist(err) {
		if err := os.MkdirAll(to, 0o700); err != nil {
			return err
		}
	}
	name := path.Join(to, p.Name)
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", p.Name)
	}
	return os.WriteFile(name, body, 0o644)
}
