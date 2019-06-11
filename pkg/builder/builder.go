package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nginxinc/crossplane-go/pkg/parser"
)

var padding string
var spacing int

// Build takes a string representing NGINX configuration
// builds it into conf format and returns that as a string
func Build(payload string, indent int, tabs, header bool) (string, error) {
	data := []parser.Block{}
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
	var b string

	body := BuildBlock(b, data, 0, 0)

	return body, nil
}

// BuildBlock -
func BuildBlock(output string, block []parser.Block, depth, lastline int) string {
	var built string
	margin := strings.Repeat(padding, depth)

	for _, stmt := range block {
		line := 0
		tab := ""
		if stmt.Directive == "#" && stmt.Line != 1 {
			output += "\n" + " #" + stmt.Comment
			continue
		} else if stmt.Directive == "#" && stmt.Line == 1 {
			output += tab + "#" + stmt.Comment
		} else {

			if stmt.Directive == "if" {
				built = "if (" + strings.Join(stmt.Args, " ") + ")"
			} else if len(stmt.Args) > 0 {
				built = stmt.Directive + " " + strings.Join(stmt.Args, " ")
			} else {
				built = stmt.Directive
			}

			if len(stmt.Block) < 1 {
				built += ";"
			} else {
				built += " {"
				built = BuildBlock(built, stmt.Block, depth+1, line)
				built += "\n" + margin + "}"
				if spacing != 0 {
					spacing -= 4
				}
			}
			if output != " " {
				output += "\n" + tab + margin + built
			} else {
				output += " " + margin + built
			}
			lastline = line
			output = strings.Replace(output, "\t", padding, -1)

		}
		tab = strings.Repeat(" ", spacing)
	}
	return output
}

// BuildFiles -
func BuildFiles(data parser.Payload, dirname string, indent int, tabs, header bool) (string, error) {

	var built string
	var err error
	var output string
	var file string
	if dirname == " " {
		dirname, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	for _, payload := range data.Config {
		path := payload.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(dirname+"/", path)
		}
		dirpath := filepath.Dir(path)
		file = filepath.Base(path)
		//if _, err = os.Stat(dirpath); os.IsNotExist(err) {
		os.MkdirAll(dirpath, 0777)
		//}

		parsed := payload.Parsed
		out, err := json.Marshal(parsed)
		if err != nil {
			return "", err
		}

		output, err = Build(string(out), 4, false, false)
		if err != nil {
			return "", err
		}
		output = strings.TrimLeft(output, "\n")
		path = dirpath + "/" + file
		err = ioutil.WriteFile(path, []byte(output), 0777)
		if err != nil {
			return "", err
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		built += string(b)
	}

	return built, nil
}
