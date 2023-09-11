package languages

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

type YamlParser struct {
	fileList []*object.File
}

func NewYamlParser(fileList []*object.File) *YamlParser {
	return &YamlParser{
		fileList: fileList,
	}
}

func (y *YamlParser) Parse(fileList []*object.File) error {

	return nil
}
