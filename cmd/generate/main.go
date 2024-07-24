/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"flag"
	"log"
	"os"

	"github.com/nginxinc/nginx-go-crossplane/internal/generator"
)

func main() {
	var (
		sourceCodePath = flag.String("src-path", "",
			"The path of source code your want to generate support from, it can be either a file or a directory. (required)")
		directiveMapName = flag.String("directive-map-name", "", "Name of the generated map variable."+
			" Normally it should start with lowercase to avoid export. (required)")
		matchFuncName = flag.String("match-func-name", "", "Name of the generated matchFunc."+
			" Normally it should start with uppercase to export. (required)")
		matchFnComment = flag.String("match-func-comment", "", "The code comment for generated matchFunc."+
			" You can add some explanations like which modules included in it. Normally it should start with match-func-name (optional)")
		filterflags       filterFlag
		directiveOverride override
	)
	flag.Var(&filterflags, "filter",
		"A list of strings specifying the directives to exclude from the output. "+
			"An example is: -filter directive1 -filter directive2... (optional)")
	flag.Var(&directiveOverride, "override",
		"A list of strings, used to override the output. "+
			"It should follow the format:{directive:bitmask00|bitmask01...,bitmask10|bitmask11...}"+"\n"+
			"An example is -override=log_format:ngxHTTPMainConf|ngxConf2More,ngxStreamMainConf|ngxConf2More"+"\n"+
			`To use | and , in command line, you may need to enclose your input in quotes, i.e. -override="directive:mask1,mask2,...". (optional)`)

	flag.Parse()

	if *sourceCodePath == "" {
		log.Fatal("src-path can't be empty")
	}

	if *directiveMapName == "" {
		log.Fatal("directive-map can't be empty")
	}

	if *matchFuncName == "" {
		log.Fatal("match-func can't be empty")
	}

	config := generator.GenerateConfig{
		Filter:           filterflags.filter,
		Override:         directiveOverride,
		DirectiveMapName: *directiveMapName,
		MatchFuncName:    *matchFuncName,
		MatchFuncComment: *matchFnComment,
	}

	err := generator.Generate(*sourceCodePath, os.Stdout, config)
	if err != nil {
		log.Fatal(err)
	}
}
