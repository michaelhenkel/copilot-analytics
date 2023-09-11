package languages

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Config struct {
	Repo      Repo
	Languages []Language
	Commits   Commits
}

// bla
type Repo struct {
	Url   string
	Token string
	Local *string
}

type Commits struct {
	From *string
	To   *string
}

type Language struct {
	Name       LanguageName
	Extensions []string
	Annotation Annotation
}

type Annotation struct {
	Start string
	End   string
}

type LanguageName string

func NewParser(languageName LanguageName, conf *Config, fileList []*object.File) LanguageInterface {
	switch languageName {
	case Go:
		for _, lang := range conf.Languages {
			if lang.Name == languageName {
				return NewGoparser(fileList, &lang)
			}
		}
	case Yaml:
		return NewYamlParser(fileList)
	default:
		return nil
	}
	return nil
}

const (
	Go   LanguageName = "go"
	Yaml LanguageName = "yaml"
)

type LanguageInterface interface {
	// GetExtensions returns the extensions for the language
	Parse(fileList []*object.File) error
}
