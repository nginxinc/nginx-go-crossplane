/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	"encoding/json"
	"io"
)

type Filters map[string]struct{}

func (fs *Filters) UnmarshalJSON(data []byte) error {
	var v []string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if *fs == nil {
		*fs = make(Filters)
	}

	for _, s := range v {
		(*fs)[s] = struct{}{}
	}
	return nil
}

type GenerateConfig struct {
	// Filter is a map used to exclude directives from generator.
	// The key of it is the directive names.
	Filter Filters `json:"filter"`

	// Override is a map used to override the masks from source code of directives.
	// The key of it is the directive name. The value is the masks we want.
	// If a directive exists in Override, generator won't consider its definition
	// in source code.
	Override map[string][]Mask `json:"override"`

	// DirectiveMapName is the name assigned to the variable containing the directives
	// discovered by generator. The variable name generally starts with a
	// lowercase to avoid export. Users will use the generated function named by
	// MatchFuncName to validate the module directives in nginx configurations.
	// It should not be empty.
	DirectiveMapName string `json:"directiveMapName"`

	// MatchFuncName is the name assigned to the matchFunc generated by the generator.
	// It should generally start with a uppercase to export.
	// Users will use the generated function named by MatchFuncName
	// to validate the module directives in nginx configurations.
	// It should not be empty.
	MatchFuncName string `json:"matchFuncName"`

	// MatchFuncComment is the comment appears above the generated MatchFunc.
	// It may contain some information like what modules are included
	// in the generated MatchFunc. Generally it should start with MatchFuncName.
	// If it is empty, no comments will appear above the generated MatchFunc.
	MatchFuncComment string `json:"matchFuncComment"`
}

// Generate receives a string sourcePath, an io.Writer writer, and a
// GenerateConfig config. It will extract all the directives definitions
// from the .c and .cpp files in sourcePath and its subdirectories,
// then output the corresponding directive masks map and matchFunc via writer.
func Generate(sourcePath string, writer io.Writer, config GenerateConfig) error {
	return genFromSrcCode(sourcePath, writer, config)
}
