package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"copilot-analytics/languages"
)

type Config struct {
	Repo      Repo
	Languages []languages.Language
	Commits   Commits
}

// bla
type Repo struct {
	Url      string
	Token    string
	LocalDir *string
}

type Commits struct {
	From *string
	To   *string
}

type result struct {
	filesMap map[languages.LanguageName][]string
}

func (r *result) eval() error {
	for language, fileList := range r.filesMap {
		languageInterface := languages.NewParser(language, fileList)
		if err := languageInterface.Parse(fileList); err != nil {
			return err
		}
	}
	if go_ext, ok := r.filesMap[languages.Go]; ok {
		fmt.Println("Go files:")
		for _, file := range go_ext {
			fmt.Println(file)
		}
	}
	return nil
}

func newResult() *result {
	return &result{
		filesMap: make(map[languages.LanguageName][]string),
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "copilot-analytics",
		Short: "Copilot Analytics is a tool for analyzing your GitHub Copilot usage.",
		Long:  `Copilot Analytics is a tool for analyzing your GitHub Copilot usage.`,
	}
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get information about your Copilot usage",
		Long:  `Get information about your Copilot usage`,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := get()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err := result.eval(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	config string
)

func initConfig() {

}

func readConfig() (*Config, error) {
	data, err := os.ReadFile(config)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}
	conf := &Config{}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	getCmd.PersistentFlags().StringVar(&config, "config", "", "path to config file")
	getCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(getCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func get() (*result, error) {

	conf, err := readConfig()
	if err != nil {
		return nil, err
	}
	token, err := readTokenFromFile(conf.Repo.Token)
	if err != nil {
		return nil, err
	}

	fmt.Println(conf)
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "a",
			Password: token,
		},
		URL:      conf.Repo.Url,
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}

	result := newResult()

	var fromRef plumbing.Hash
	var toRef plumbing.Hash

	if conf.Commits.From != nil && conf.Commits.To != nil {
		fromRef = plumbing.NewHash(*conf.Commits.From)
		fromCIter, err := r.Log(&git.LogOptions{
			From: fromRef,
		})
		if err != nil {
			return nil, err
		}

		toRef = plumbing.NewHash(*conf.Commits.To)
		toCIter, err := r.Log(&git.LogOptions{
			From: toRef,
		})
		if err != nil {
			return nil, err
		}
		var fileMap = make(map[string]bool)
		if err := fromCIter.ForEach(func(fromCommit *object.Commit) error {
			if err := toCIter.ForEach(func(toCommit *object.Commit) error {
				patch, err := fromCommit.Patch(toCommit)
				if err != nil {
					return err
				}
				patchString := patch.Stats().String()
				scanner := bufio.NewScanner(strings.NewReader(patchString))
				for scanner.Scan() {
					fileString := strings.TrimSpace(strings.Split(scanner.Text(), " | ")[0])
					fileMap[fileString] = true
				}

				if err := scanner.Err(); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return nil
			}
			return nil
		}); err != nil {
			return nil, err
		}
		fileList(result, conf, fileMap)
	} else {

		ref, err := r.Head()
		if err != nil {
			return nil, err
		}

		cIter, err := r.Log(&git.LogOptions{
			From: ref.Hash(),
		})
		if err != nil {
			return nil, err
		}
		var fileMap = make(map[string]bool)
		if err := cIter.ForEach(func(commit *object.Commit) error {
			files, err := commit.Files()
			if err != nil {
				return err
			}
			files.ForEach(func(file *object.File) error {
				fileMap[file.Name] = true
				return nil
			})
			return nil
		}); err != nil {
			return nil, err
		}
		fileList(result, conf, fileMap)
	}
	return result, nil
}

func fileList(res *result, conf *Config, fileMap map[string]bool) {
	for file := range fileMap {
		extension := path.Ext(file)
		for _, language := range conf.Languages {
			if len(language.Extensions) > 0 {
				for _, ext := range language.Extensions {
					if extension == ext {
						if val, ok := res.filesMap[language.Name]; ok {
							res.filesMap[language.Name] = append(val, file)
						} else {
							res.filesMap[language.Name] = []string{file}
						}

					}
				}
			} else {
				if val, ok := res.filesMap[language.Name]; ok {
					res.filesMap[language.Name] = append(val, file)
				} else {
					res.filesMap[language.Name] = []string{file}
				}
			}
		}
	}
}

func readTokenFromFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var token string
	_, err = fmt.Fscanf(file, "%s", &token)
	if err != nil {
		return "", err
	}

	tokenList := strings.Split(token, "token=")
	if len(tokenList) != 2 {
		return "", err
	}
	return tokenList[1], nil
}
