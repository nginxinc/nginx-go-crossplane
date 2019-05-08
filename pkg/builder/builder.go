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

var padding string

// Build takes a string representing NGINX configuration
// builds it into conf format and returns that as a string
func Build(payload string, indent int, tabs, header bool) (string, error) {
	data := []Block{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling payload: %v", err)
	}

	if tabs {
		padding = "\t"
	} else {
		padding = strings.Repeat(" ", indent)
	}

	var body string
	body = BuildBlock(body, data, 0, 0)

	return body, nil
}

// BuildBlock -
func BuildBlock(output string, block []Block, depth, lastline int) string {
	var built string
	margin := strings.Repeat(padding, depth)

	for _, stmt := range block {
		line := stmt.Line
		if stmt.Directive == "#" && line == lastline {
			output += " #" + stmt.Comment
			continue
		} else if stmt.Directive == "#" {
			built = "#" + stmt.Comment
		} else {

			if stmt.Directive == "if" {
				built = "if (" + strings.Join(stmt.Args, " ") + ")"
			} else if len(stmt.Args) > 0 {
				built = stmt.Directive + " " + strings.Join(stmt.Args, " ")
			} else {
				built = stmt.Directive
			}

			if len(stmt.Block) <= 0 {
				built += ";"
			} else {
				built += " {"
				built = BuildBlock(built, stmt.Block, depth+1, line)
				built += "\n" + margin + "}"
			}

			if output != " " {
				output += "\n" + margin + built
			} else {
				output += " " + margin + built
			}
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
