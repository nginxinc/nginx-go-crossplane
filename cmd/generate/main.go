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
		filterflags       filterFlag
		directiveOverride override
	)
	flag.Var(&filterflags, "filter",
		"A list of strings specifying the directives to exclude from the output. "+
			"An example is: -filter directive1 -filter directive2...(optional)")
	flag.Var(&directiveOverride, "override",
		"A list of strings, used to override the output. "+
			"It should follow the format:{directive:bitmask00|bitmask01...,bitmask10|bitmask11...}"+"\n"+
			"An example is -override=log_format:ngxHTTPMainConf|ngxConf2More,ngxStreamMainConf|ngxConf2More"+"\n"+
			`To use | and , in command line, you may need to enclose your input in quotes, i.e. -override="directive:mask1,mask2,...". (optional)`)

	flag.Parse()

	err := generator.Generate(*sourceCodePath, os.Stdout, filterflags.filter, directiveOverride)
	if err != nil {
		log.Fatal(err)
	}
}
