/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/nginxinc/nginx-go-crossplane/internal/generator"
)

func configFromFile(path string) (generator.GenerateConfig, error) {
	var config generator.GenerateConfig
	f, err := os.Open(path)
	if err != nil {
		return config, err
	}

	defer f.Close()

	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

//nolint:funlen
func main() {
	var (
		sourceCodePath = flag.String("src-path", "",
			"The path of source code your want to generate support from, it can be either a file or a directory. (required)")
		configPath = flag.String("config-path", "", "The path of json config file.\n"+
			"The file can contain directiveMapName, matchFuncName, matchFuncComment, filter, and override.\n"+
			"They provide same functions as other arguments directive-map-name, match-func-name, match-func-comment, filter, and override.\n"+
			"It will unmarsh to generator.GenerateConfig. (optional)")
		directiveMapName = flag.String("directive-map-name", "", "Name of the generated map variable."+
			"Normally it should start with lowercase to avoid export. If this is provided, the directiveMapName in json config will be ignored.\n"+
			"You should provide it here or in json config.")
		matchFuncName = flag.String("match-func-name", "", "Name of the generated matchFunc."+
			"Normally it should start with uppercase to export. If this is provided, the match-func-name in json config will be ignored.\n"+
			"You should provide it here or in json config.")
		matchFnComment = flag.String("match-func-comment", "", "The code comment for generated matchFunc."+
			"You can add some explanations like which modules included in it. Normally it should start with match-func-name.\n"+
			"If this is provided, the matchFuncComment in json config will be ignored. (optional)")
		filterflags       filterFlag
		directiveOverride override
	)
	flag.Var(&filterflags, "filter",
		"A list of strings specifying the directives to exclude from the output. "+
			"An example is: -filter directive1 -filter directive2...\n"+
			"If this is provided, the filter in json config will be igonored. (optional)")
	flag.Var(&directiveOverride, "override",
		"A list of strings, used to override the output. "+
			"It should follow the format:{directive:bitmask00|bitmask01...,bitmask10|bitmask11...}"+"\n"+
			"An example is -override=log_format:ngxHTTPMainConf|ngxConf2More,ngxStreamMainConf|ngxConf2More"+"\n"+
			`To use | and , in command line, you may need to enclose your input in quotes, i.e. -override="directive:mask1,mask2,...".`+
			`If this is provided, the override in json config will be ignored. (optional)`)

	flag.Parse()
	var err error

	var config generator.GenerateConfig

	if *configPath != "" {
		config, err = configFromFile(*configPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *directiveMapName != "" {
		config.DirectiveMapName = *directiveMapName
	}

	if *matchFuncName != "" {
		config.MatchFuncName = *matchFuncName
	}

	if filterflags.filter != nil {
		config.Filter = filterflags.filter
	}

	if directiveOverride != nil {
		config.Override = directiveOverride
	}

	if *matchFnComment != "" {
		config.MatchFuncComment = *matchFnComment
	}

	if *sourceCodePath == "" {
		log.Fatal("src-path can't be empty")
	}

	if config.DirectiveMapName == "" {
		log.Fatal("directiveMapName can't be empty")
	}

	if config.MatchFuncName == "" {
		log.Fatal("matchFuncName can't be empty")
	}

	err = generator.Generate(*sourceCodePath, os.Stdout, config)
	if err != nil {
		log.Fatal(err)
	}
}
