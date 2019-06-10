package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	tab := strings.Repeat("\t", spacing)

	for _, stmt := range block {
		line := 0

		if stmt.Directive == "#" && stmt.Line != 1 {
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

			if len(stmt.Block) < 1 {
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
func BuildFiles(data parser.Payload, dirname string, indent int, tabs, header bool) (string, error) {

	var built string
	var err error
	var output string
	if dirname == " " {
		dirname, _ = os.Getwd()
	}

	for _, payload := range data.Config {

		path := payload.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(dirname, path)
		}

		parts := strings.Split(payload.File, "/")
		dirpath := parts[0]
		if _, err = os.Stat(dirpath); os.IsNotExist(err) {
			os.Mkdir(dirpath, 0777)
		}

		parsed := payload.Parsed
		out, err := json.Marshal(parsed)
		if err != nil {
			log.Fatal("Problem json-ing data : ", err)
		}

		output, err = Build(string(out), 4, false, false)
		if err != nil {
			log.Fatal("Build fail ", err)
		}
		output = strings.TrimLeft(output, "\n")

		err = ioutil.WriteFile("nginx.conf", []byte(output), 0777)
		if err != nil {
			log.Fatal("Can't write to file ", err)
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("Couldn't read from file : ", err)
		}
		built = string(b)
	}

	return built, nil
}
