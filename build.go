/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type BuildOptions struct {
	Indent int
	Tabs   bool
	Header bool
}

const MaxIndent = 100

//nolint:gochecknoglobals
var (
	marginSpaces = strings.Repeat(" ", MaxIndent)
	marginTabs   = strings.Repeat("\t", MaxIndent)
)

const header = `# This config was built from JSON using NGINX crossplane.
# If you encounter any bugs please report them here:
# https://github.com/nginxinc/crossplane/issues

`

// BuildFiles builds all of the config files in a crossplane.Payload and
// writes them to disk.
func BuildFiles(payload Payload, dir string, options *BuildOptions) error {
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = cwd
	}

	for _, config := range payload.Config {
		path := config.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}

		// make directories that need to be made for the config to be built
		dirpath := filepath.Dir(path)
		if err := os.MkdirAll(dirpath, os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		// build then create the nginx config file using the json payload
		var buf bytes.Buffer
		if err := Build(&buf, config, options); err != nil {
			return err
		}

		file, err := os.Create(path)
		if err != nil {
			return err
		}

		output := append(bytes.TrimRightFunc(buf.Bytes(), unicode.IsSpace), '\n')
		if _, err := file.Write(output); err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Build creates an NGINX config from a crossplane.Config.
func Build(writer io.Writer, config Config, options *BuildOptions) error {
	if options.Indent == 0 {
		options.Indent = 4
	}

	if options.Header {
		_, err := writer.Write([]byte(header))
		if err != nil {
			return err
		}
	}

	body := strings.Builder{}
	buildBlock(&body, nil, config.Parsed, 0, 0, options)

	bodyStr := body.String()
	if len(bodyStr) > 0 && bodyStr[len(bodyStr)-1] == '\n' {
		bodyStr = bodyStr[:len(bodyStr)-1]
	}

	_, err := writer.Write([]byte(bodyStr))
	return err
}

//nolint:cyclop
func buildBlock(writer io.StringWriter, parent *Directive, block Directives, depth int, lastLine int, options *BuildOptions) {
	for index, stmt := range block {
		// if the this statement is a comment on the same line as the preview, do not emit EOL for this stmt
		if stmt.Line == lastLine && stmt.IsComment() {
			_, _ = writer.WriteString(" #")
			_, _ = writer.WriteString(*stmt.Comment)
			// sb.WriteString("\n")
			continue
		}

		if index != 0 || parent != nil {
			_, _ = writer.WriteString("\n")
		}

		_, _ = writer.WriteString(margin(options, depth))

		if stmt.IsComment() {
			_, _ = writer.WriteString("#")
			_, _ = writer.WriteString(*stmt.Comment)
		} else {
			directive := Enquote(stmt.Directive)
			_, _ = writer.WriteString(directive)

			// special handling for if statements
			if directive == "if" {
				_, _ = writer.WriteString(" (")
				for i, arg := range stmt.Args {
					if i > 0 {
						_, _ = writer.WriteString(" ")
					}
					_, _ = writer.WriteString(Enquote(arg))
				}
				_, _ = writer.WriteString(")")
			} else {
				for _, arg := range stmt.Args {
					_, _ = writer.WriteString(" ")
					_, _ = writer.WriteString(Enquote(arg))
				}
			}

			if !stmt.IsBlock() {
				_, _ = writer.WriteString(";")
			} else {
				_, _ = writer.WriteString(" {")
				stmt := stmt
				buildBlock(writer, stmt, stmt.Block, depth+1, stmt.Line, options)
				_, _ = writer.WriteString("\n")
				_, _ = writer.WriteString(margin(options, depth))
				_, _ = writer.WriteString("}")
			}
		}
		lastLine = stmt.Line
	}
}
func margin(options *BuildOptions, depth int) string {
	indent := depth * options.Indent
	if indent < MaxIndent {
		if options.Tabs {
			return marginTabs[:depth]
		}
		return marginSpaces[:indent]
	}

	if options.Tabs {
		return strings.Repeat("\t", depth)
	}
	return strings.Repeat(" ", options.Indent*depth)
}

func Enquote(arg string) string {
	if !needsQuote(arg) {
		return arg
	}
	return strings.ReplaceAll(repr(arg), `\\`, `\`)
}

//nolint:gocyclo,cyclop
func needsQuote(input string) bool {
	if input == "" {
		return true
	}

	// lexer should throw an error when variable expansion syntax
	// is messed up, but just wrap it in quotes for now I guess
	var char rune
	chars := escape(input)

	if len(chars) == 0 {
		return true
	}

	// get first rune
	char, off := utf8.DecodeRuneInString(chars)

	// arguments can't start with variable expansion syntax
	if unicode.IsSpace(char) || strings.ContainsRune("{};\"'", char) || strings.HasPrefix(chars, "${") {
		return true
	}

	chars = chars[off:]

	expanding := false
	var prev rune
	for _, c := range chars {
		char = c

		if prev == '\\' {
			prev = 0
			continue
		}
		if unicode.IsSpace(char) || strings.ContainsRune("{;\"'", char) {
			return true
		}

		if (expanding && (prev == '$' && char == '{')) || (!expanding && char == '}') {
			return true
		}

		if (expanding && char == '}') || (!expanding && (prev == '$' && char == '{')) {
			expanding = !expanding
		}

		prev = char
	}

	return expanding || char == '\\' || char == '$'
}

//nolint:cyclop
func escape(input string) string {
	if !strings.ContainsAny(input, "{}$;\\") {
		return input
	}

	builder := strings.Builder{}
	var prevChar, currentChar rune

	for _, r := range input {
		currentChar = r
		if prevChar == '\\' || (prevChar == '$' && currentChar == '{') {
			builder.WriteRune(prevChar)
			builder.WriteRune(currentChar)
			prevChar = 0
			continue
		}

		if prevChar == '$' {
			builder.WriteRune(prevChar)
		}
		if currentChar != '\\' && currentChar != '$' {
			builder.WriteRune(currentChar)
		}
		prevChar = currentChar
	}

	if currentChar == '\\' || currentChar == '$' {
		builder.WriteRune(currentChar)
	}

	return builder.String()
}

// BuildInto builds all of the config files in a crossplane.Payload and
// writes them to the Creator.
func BuildInto(payload *Payload, into Creator, options *BuildOptions) error {
	for _, config := range payload.Config {
		writerCloser, err := into.Create(config.File)
		if err != nil {
			return err
		}
		if err := Build(writerCloser, config, options); err != nil {
			return err
		}

		if err := writerCloser.Close(); err != nil {
			return err
		}
	}

	return nil
}
