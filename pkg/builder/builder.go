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
	result = BuildBlock(payload, data, 0, 0)

	return result, nil
}

// BuildBlock -
func BuildBlock(output string, block []Block, depth, lastline int) string {
	/*
		b := Block{
			Directive: " ",
			Line:      0,
			Args:      []string{},
			Includes:  []int{},
			Block:     []Block{},
			File:      " ",
			Comment:   " ",
		}*/

	var built string

	for _, stmt := range block {
		directive := stmt.Directive
		line := stmt.Line
		if directive == "#" && line == lastline {
			output += " #" + stmt.Comment
			continue
		} else if directive == "#" {
			built = "#" + "" + stmt.Comment
		}
		if stmt.Args != nil {
			built = stmt.Directive + " " + strings.Join(stmt.Args, "")
		} else {
			built = directive
		}
		if stmt.Block == nil {
			built += ";"
		} else {
			built += " {"
			built = BuildBlock(built, stmt.Block, depth+1, line)
			built += "\n " + "}"

			output += " " + built
			lastline = line
		}
		/*
			if stmt.Directive == "#" {
				b = Block{
					Directive: "#",
					Line:      stmt.Line,
					Args:      []string{},
					Includes:  []int{},
					Block:     []Block{},
					File:      " ",
					Comment:   stmt.Comment,
				}
			}
		*/
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
