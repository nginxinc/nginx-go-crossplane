/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"fmt"
	"strings"
	"unicode"
)

type included struct {
	directive *Directive
	err       error
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}

func isSpace(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func isEOL(s string) bool {
	return strings.HasSuffix(s, "\n")
}

func repr(s string) string {
	quote := fmt.Sprintf("%q", s)
	for _, char := range s {
		if char == '"' {
			quote = strings.ReplaceAll(quote, `\"`, `"`)
			quote = strings.ReplaceAll(quote, `'`, `\'`)
			quote = `'` + quote[1:len(quote)-1] + `'`
			return quote
		}
	}
	return quote
}

func validFlag(s string) bool {
	l := strings.ToLower(s)
	return l == "on" || l == "off"
}

// validExpr ensures an expression is enclused in '(' and ')' and is not empty.
func validExpr(directive *Directive) bool {
	length := len(directive.Args)
	begin := 0
	end := length - 1

	return length > 0 &&
		strings.HasPrefix(directive.Args[begin], "(") &&
		strings.HasSuffix(directive.Args[end], ")") &&
		((length == 1 && len(directive.Args[begin]) > 2) || // empty expression single arg '()'
			(length == 2 && (len(directive.Args[begin]) > 1 || len(directive.Args[end]) > 1)) || // empty expression two args '(', ')'
			(length > 2))
}

// prepareIfArgs removes parentheses from an `if` directive's arguments.
func prepareIfArgs(directive *Directive) *Directive {
	begin := 0
	end := len(directive.Args) - 1
	if len(directive.Args) > 0 && strings.HasPrefix(directive.Args[0], "(") && strings.HasSuffix(directive.Args[end], ")") {
		directive.Args[0] = strings.TrimLeftFunc(strings.TrimPrefix(directive.Args[0], "("), unicode.IsSpace)
		directive.Args[end] = strings.TrimRightFunc(strings.TrimSuffix(directive.Args[end], ")"), unicode.IsSpace)
		if len(directive.Args[0]) == 0 {
			begin++
		}
		if len(directive.Args[end]) == 0 {
			end--
		}
		directive.Args = directive.Args[begin : end+1]
	}
	return directive
}

// combineConfigs combines config files into one by using include directives.
func combineConfigs(old *Payload) (*Payload, error) {
	if len(old.Config) < 1 {
		return old, nil
	}

	status := old.Status
	if status == "" {
		status = "ok"
	}

	errors := old.Errors
	if errors == nil {
		errors = []PayloadError{}
	}

	combined := Config{
		File:   old.Config[0].File,
		Status: "ok",
		Errors: []ConfigError{},
		Parsed: Directives{},
	}

	for _, config := range old.Config {
		combined.Errors = append(combined.Errors, config.Errors...)
		if config.Status == "failed" {
			combined.Status = "failed"
		}
	}

	for incl := range performIncludes(old, combined.File, old.Config[0].Parsed) {
		if incl.err != nil {
			return nil, incl.err
		}
		combined.Parsed = append(combined.Parsed, incl.directive)
	}

	return &Payload{
		Status: status,
		Errors: errors,
		Config: []Config{combined},
	}, nil
}

func performIncludes(old *Payload, fromfile string, block Directives) chan included {
	channel := make(chan included)
	go func() {
		defer close(channel)
		for _, d := range block {
			dir := *d
			if dir.IsBlock() {
				nblock := Directives{}
				for incl := range performIncludes(old, fromfile, dir.Block) {
					if incl.err != nil {
						channel <- incl
						return
					}
					nblock = append(nblock, incl.directive)
				}
				dir.Block = nblock
			}
			if !dir.IsInclude() {
				channel <- included{directive: &dir}
				continue
			}
			for _, idx := range dir.Includes {
				if idx >= len(old.Config) {
					channel <- included{
						err: &ParseError{
							What:      fmt.Sprintf("include config with index: %d", idx),
							File:      &fromfile,
							Line:      &dir.Line,
							Statement: dir.String(),
						},
					}
					return
				}
				for incl := range performIncludes(old, old.Config[idx].File, old.Config[idx].Parsed) {
					channel <- incl
				}
			}
		}
	}()
	return channel
}
