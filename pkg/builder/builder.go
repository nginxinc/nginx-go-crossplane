package builder

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Block -
type Block struct {
	Directive string
	Line      int
	Args      []string
	Includes  []int
	Block     []Block
	File      string
	Comment   string
}

// Config -
type Config struct {
	File   string
	Status string
	Errors []ParseError
	Parsed []Block
}

// ParseError -
type ParseError struct {
	File  string
	Line  string
	Error string
}

// Build takes a string representing NGINX configuration
// builds it into conf format and returns that as a string
func Build(payload string, indent int, tabs, header bool) (string, error) {
	data := []Block{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling payload: %v", err)
	}
	var result string
	var output string
	result = BuildBlock(output, data, 0, 0)

	return result, nil
}

// BuildBlock -
func BuildBlock(output string, block []Block, depth, lastline int) string {
	var built string
	for _, stmt := range block {
		//directive := stmt.Directive
		var line int
		if stmt.Directive == "#" && line == lastline {
			output += " #" + stmt.Comment
			continue
		} else if stmt.Directive == "#" {
			built = "#" + stmt.Comment
		} else {
			if stmt.Directive == "if" {
				built = "if (" + " " + strings.Join(stmt.Args, "") + ")"
			} else if stmt.Args != nil {
				built = stmt.Directive + " " + strings.Join(stmt.Args, "")
			} else {
				built = stmt.Directive
			}
			if len(stmt.Block) <= 0 {
				built += ";"
			} else {
				built += " {"
				built = BuildBlock(built, stmt.Block, depth, line)
				built += "\n }"
			}
			output += " " + built
			lastline = line
		}
	}
	return output
}

// BuildFiles -
func BuildFiles(payload string, dirname string, indent int, tabs, header bool) (string, error) {
	data := Config{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling payload: %v", err)
	}

	return "built", nil
}
