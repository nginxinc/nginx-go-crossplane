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
)

type Payload struct {
	Status string         `json:"status"`
	Errors []PayloadError `json:"errors"`
	Config []Config       `json:"config"`
}

type PayloadError struct {
	File     string      `json:"file"`
	Line     *int        `json:"line"`
	Error    error       `json:"error"`
	Callback interface{} `json:"callback,omitempty"`
}

type Config struct {
	File   string        `json:"file"`
	Status string        `json:"status"`
	Errors []ConfigError `json:"errors"`
	Parsed Directives    `json:"parsed"`
}

type ConfigError struct {
	Line  *int  `json:"line"`
	Error error `json:"error"`
}

type Directive struct {
	Directive string     `json:"directive"`
	Line      int        `json:"line"`
	Args      []string   `json:"args"`
	File      string     `json:"file,omitempty"`
	Includes  []int      `json:"includes,omitempty"`
	Block     Directives `json:"block,omitempty"`
	Comment   *string    `json:"comment,omitempty"`
}
type Directives []*Directive

// IsBlock returns true if this is a block directive.
func (d Directive) IsBlock() bool {
	return d.Block != nil
}

// IsInclude returns true if this is an include directive.
func (d Directive) IsInclude() bool {
	return d.Directive == "include" && d.Includes != nil
}

// IsComment returns true iff the directive represents a comment.
func (d Directive) IsComment() bool {
	return d.Directive == "#" && d.Comment != nil
}

func equals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if x != b[i] {
			return false
		}
	}
	return true
}

// strPtrEqual returns true if the content of the provided string pointer are equal.
func strPtrEqual(str1, str2 *string) bool {
	if str1 == str2 {
		return true
	}
	if str1 == nil || str2 == nil {
		return false
	}
	return *str1 == *str2
}

// Equal returns true if both blocks are functionally equivalent.
//
//nolint:cyclop
func (d *Directive) Equal(directive *Directive) bool {
	if d == directive {
		// same ptr, or both nil
		return true
	}
	if d == nil || directive == nil {
		return false
	}
	switch {
	case directive.Directive != d.Directive:
		return false
	case !equals(directive.Args, d.Args):
		return false
	case len(directive.Block) != len(d.Block):
		return false
	case len(directive.Includes) != len(d.Includes):
		return false
	case !strPtrEqual(directive.Comment, d.Comment):
		return false
	case directive.Line != d.Line:
		return false
	case directive.File != d.File:
		return false
	}
	for i, inc := range directive.Includes {
		if inc != d.Includes[i] {
			return false
		}
	}
	for i, dir := range directive.Block {
		if !dir.Equal(d.Block[i]) {
			return false
		}
	}
	return true
}

// String makes this a Stringer, returning a string representation of the Directive. The string representation is a
// peak at the content of the Directive, does not represent a valid config rendering of the Directive in question.
func (d *Directive) String() string {
	if len(d.Block) == 0 {
		return fmt.Sprintf("%s %s", d.Directive, strings.Join(d.Args, " "))
	}
	return fmt.Sprintf("%s %s {...}", d.Directive, strings.Join(d.Args, " "))
}

// Combined returns a new Payload that is the same except that the inluding
// logic is performed on its configs. This means that the resulting Payload
// will always have 0 or 1 configs in its Config field.
func (p *Payload) Combined() (*Payload, error) {
	return combineConfigs(p)
}
