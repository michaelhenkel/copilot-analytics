package languages

type YamlParser struct {
	fileList []string
}

func NewYamlParser(fileList []string) *YamlParser {
	return &YamlParser{
		fileList: fileList,
	}
}

func (y *YamlParser) Parse(fileList []string) error {
	return nil
}
