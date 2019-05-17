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
var spacing int

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

	spacing = indent

	var body string
	body = BuildBlock(body, data, 0, 0)

	return body, nil
}

// BuildBlock -
func BuildBlock(output string, block []Block, depth, lastline int) string {
	var built string
	margin := strings.Repeat(padding, depth)
	tab := strings.Repeat("\t", spacing)

	for _, stmt := range block {
		line := 0

		if stmt.Directive == "#" && line == lastline && stmt.Line != 1 {
			output += " #" + stmt.Comment
			continue
		} else if stmt.Directive == "#" && stmt.Line == 1 {
			output = "\n" + tab + "#" + stmt.Comment
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
				built += "\n" + tab + margin + "}"
			}

			if output != " " {
				output += "\n" + tab + margin + built
			} else {
				output += " " + margin + built
			}
			lastline = line
			output = strings.Replace(output, "\t", padding, -1)
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
