package languages

import "fmt"

type GoParser struct {
	fileList []string
}

func NewGoparser(fileList []string) *GoParser {
	return &GoParser{
		fileList: fileList,
	}
}

func (g *GoParser) Parse(fileList []string) error {
	for _, file := range g.fileList {
		fmt.Println(file)
	}
	return nil
}
