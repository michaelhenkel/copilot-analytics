package languages

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type GoParser struct {
	fileList   []*object.File
	annotation string
}

func NewGoparser(fileList []*object.File, language *Language) *GoParser {
	return &GoParser{
		fileList:   fileList,
		annotation: language.Annotation,
	}
}

// +copilot
func (g *GoParser) Parse(fileList []*object.File) error {
	for _, file := range g.fileList {

		content, err := file.Contents()
		if err != nil {
			return err
		}
		if file.Name == "languages/go_parser.go" {
			fmt.Println("File:", file.Name)
			fmt.Println("content:", content)
		}
		if err := g.scan(content); err != nil {
			return err
		}

	}
	return nil
}

func (g *GoParser) scan(content string) error {
	// Specify the path to your Go source code file.

	// Initialize the counters.
	functionCount := 0
	structCount := 0
	currentCount := 0
	inFunction := false
	inStruct := false

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Iterate through each line in the file.
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, g.annotation) {
			if strings.Contains(line, "func ") {
				// This line contains a function annotated with "// +count".
				functionCount++
				inFunction = true
			} else if strings.Contains(line, "type ") {
				// This line contains a struct annotated with "// +count".
				structCount++
				inStruct = true
			}
		}

		// Count lines for the current function or struct.
		if inFunction || inStruct {
			currentCount++
		}

		// Check for the end of a function or struct.
		if strings.TrimSpace(line) == "}" {
			if inFunction {
				fmt.Printf("Lines in function %d: %d\n", functionCount, currentCount)
				inFunction = false
			} else if inStruct {
				fmt.Printf("Lines in struct %d: %d\n", structCount, currentCount)
				inStruct = false
			}
			currentCount = 0
		}
	}

	// Check for any scanner errors.
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	// Print the counts.
	fmt.Printf("Number of functions annotated with '// +count': %d\n", functionCount)
	fmt.Printf("Number of structs annotated with '// +count': %d\n", structCount)
	return nil
}
