package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"copilot-analytics/languages"
)

type result struct {
	filesMap map[languages.LanguageName][]*object.File
}

func (r *result) eval(conf *languages.Config) error {
	for languageName, fileList := range r.filesMap {
		languageInterface := languages.NewParser(languageName, conf, fileList)
		if err := languageInterface.Parse(fileList); err != nil {
			return err
		}
	}
	return nil
}

func newResult() *result {
	return &result{
		filesMap: make(map[languages.LanguageName][]*object.File),
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
			conf, err := readConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			result, err := get(conf)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err := result.eval(conf); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	configFile string
)

func initConfig() {

}

func readConfig() (*languages.Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}
	conf := &languages.Config{}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	getCmd.PersistentFlags().StringVar(&configFile, "config", "", "path to config file")
	getCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(getCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func get(conf *languages.Config) (*result, error) {

	var repo *git.Repository

	if conf.Repo.Local != nil {
		fs := osfs.New(*conf.Repo.Local)
		if _, err := fs.Stat(git.GitDirName); err == nil {
			fs, err = fs.Chroot(git.GitDirName)
			if err != nil {
				return nil, err
			}
		}
		s := filesystem.NewStorageWithOptions(fs, cache.NewObjectLRUDefault(), filesystem.Options{KeepDescriptors: true})
		r, err := git.Open(s, fs)
		if err != nil {
			return nil, err
		}
		repo = r
	} else {
		token, err := readTokenFromFile(conf.Repo.Token)
		if err != nil {
			return nil, err
		}
		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: "a",
				Password: token,
			},
			URL: conf.Repo.Url,
		})
		if err != nil {
			return nil, err
		}
		repo = r
	}

	result := newResult()
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil, err
	}
	var fileMap = make(map[string]*object.File)
	lastCommit, err := cIter.Next()
	if err != nil {
		return nil, err
	}
	files, err := lastCommit.Files()
	if err != nil {
		return nil, err
	}
	files.ForEach(func(file *object.File) error {
		fileMap[file.Name] = file
		return nil
	})
	fileList(result, conf, fileMap)

	return result, nil
}

func fileList(res *result, conf *languages.Config, fileMap map[string]*object.File) {
	for file, fileObject := range fileMap {
		extension := path.Ext(file)
		for _, language := range conf.Languages {
			if len(language.Extensions) > 0 {
				for _, ext := range language.Extensions {
					if extension == ext {
						if val, ok := res.filesMap[language.Name]; ok {
							res.filesMap[language.Name] = append(val, fileObject)
						} else {
							res.filesMap[language.Name] = []*object.File{fileObject}
						}

					}
				}
			} else {
				if val, ok := res.filesMap[language.Name]; ok {
					res.filesMap[language.Name] = append(val, fileObject)
				} else {
					res.filesMap[language.Name] = []*object.File{fileObject}
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
