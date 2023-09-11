package languages

type Language struct {
	Name       LanguageName
	Extensions []string
	Annotation string
}

type LanguageName string

func NewParser(language LanguageName, fileList []string) LanguageInterface {
	switch language {
	case Go:
		return NewGoparser(fileList)
	case Yaml:
		return NewYamlParser(fileList)
	default:
		return nil
	}
}

const (
	Go   LanguageName = "go"
	Yaml LanguageName = "yaml"
)

type LanguageInterface interface {
	// GetExtensions returns the extensions for the language
	Parse(fileList []string) error
}
