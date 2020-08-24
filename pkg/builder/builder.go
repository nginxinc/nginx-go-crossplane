package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

type Options struct {
	Indent int
	Tabs   bool
}

// Build takes a parsed NGINX configuration and builds an NGINX configuration
func Build(parsed []*parser.Directive, opts *Options) string {
	b := strings.Builder{}
	buildBlock(&b, parsed, 0, 0, opts)
	return b.String()
}

// BuildFromJSON takes NGINX configuration JSON and builds an NGINX configuration
func BuildFromJSON(payload string, opts *Options) (string, error) {
	data := []*parser.Directive{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return "", fmt.Errorf("error unmarshalling payload: %v", err)
	}
	return Build(data, opts), nil
}

// buildBlock recursively builds NGINX configuration blocks using a strings.Builder
func buildBlock(
	builder *strings.Builder,
	block []*parser.Directive,
	depth,
	lastLine int,
	opts *Options,
) {
	for _, stmt := range block {
		var built strings.Builder
		if stmt.IsComment() && stmt.Line == lastLine {
			builder.WriteString(" #" + stmt.Comment)
			continue
		} else if stmt.IsComment() {
			built.WriteString("#" + stmt.Comment)
		} else {
			if stmt.IsIf() {
				built.WriteString("if (" + strings.Join(stmt.Args, " ") + ")")
			} else if len(stmt.Args) > 0 {
				built.WriteString(stmt.Directive + " " + strings.Join(stmt.Args, " "))
			} else {
				built.WriteString(stmt.Directive)
			}

			if stmt.Block == nil || len(stmt.Block) < 1 {
				built.WriteString(";")
			} else {
				built.WriteString(" {")
				buildBlock(&built, stmt.Block, depth+1, stmt.Line, opts)
				built.WriteString("\n" + indent(depth, opts) + "}")
			}
		}

		if builder.Len() > 0 {
			builder.WriteString("\n")
		}

		builder.WriteString(indent(depth, opts) + built.String())
		lastLine = stmt.Line
	}
}

func indent(depth int, opts *Options) string {
	if opts.Tabs {
		return strings.Repeat("\t", opts.Indent*depth)
	}
	return strings.Repeat(" ", opts.Indent*depth)
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
		err := os.MkdirAll(dirpath, 0777)
		if err != nil {
			return "", err
		}

		parsed := payload.Parsed
		out, err := json.Marshal(parsed)
		if err != nil {
			return "", err
		}

		output, err = BuildFromJSON(string(out), &Options{Indent: 4})
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

// NewPayload -
func NewPayload(payloadBytes []byte) (data parser.Payload, err error) {
	err = json.Unmarshal(payloadBytes, &data)
	return data, err
}
