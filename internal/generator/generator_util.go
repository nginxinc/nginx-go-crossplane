/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A Mask is a list of string, includes several variable names,
// which specify a behavior of a directive.
// An example is []string{"ngxHTTPMainConf", "ngxConfFlag",}.
// A directive can have several masks.
type Mask []string

type supportFileTmplStruct struct {
	Directive2Masks map[string][]Mask
	MapVariableName string
	MatchFnName     string
}

var (
	// Extract single directive definition block
	// static ngx_command_t  {name}[] = {definition}
	// this regex extracts {name} and {definition}.
	directivesDefineBlockExtracter = regexp.MustCompile(`ngx_command_t\s+(\w+)\[\]\s*=\s*{(.*?)}\s*;`)

	// Extract one directive definition and attributes from extracted block
	// { ngx_string({directive_name}),
	//   {bitmask1|bitmask2|...},
	//   ... },
	// this regex extracts {directive_name} and {bitmask1|bitmask2|...}.
	singleDirectiveExtracter = regexp.MustCompile(`ngx_string\("(.*?)"\).*?,(.*?),`)

	singleLineCommentExtracter = regexp.MustCompile(`//.*`)

	multiLineCommentExtracter = regexp.MustCompile(`/\*[\s\S]*?\*/`)
)

// Template of support file. A support file contains a map from
// diective to its bitmask definitions, and a MatchFunc for it.
//
//go:embed tmpl/support_file.tmpl
var supportFileTmplStr string

//nolint:gochecknoglobals
var supportFileTmpl = template.Must(template.New("supportFile").
	Funcs(template.FuncMap{"Join": strings.Join}).Parse(supportFileTmplStr))

//nolint:gochecknoglobals
var ngxVarNameToGo = map[string]string{
	"NGX_MAIL_MAIN_CONF":   "ngxMailMainConf",
	"NGX_STREAM_MAIN_CONF": "ngxStreamMainConf",
	"NGX_CONF_TAKE1":       "ngxConfTake1",
	"NGX_STREAM_UPS_CONF":  "ngxStreamUpsConf",
	"NGX_HTTP_LIF_CONF":    "ngxHTTPLifConf",
	"NGX_CONF_TAKE2":       "ngxConfTake2",
	"NGX_HTTP_UPS_CONF":    "ngxHTTPUpsConf",
	"NGX_CONF_TAKE23":      "ngxConfTake23",
	"NGX_CONF_TAKE12":      "ngxConfTake12",
	"NGX_HTTP_MAIN_CONF":   "ngxHTTPMainConf",
	"NGX_HTTP_LMT_CONF":    "ngxHTTPLmtConf",
	"NGX_CONF_TAKE1234":    "ngxConfTake1234",
	"NGX_MAIL_SRV_CONF":    "ngxMailSrvConf",
	"NGX_CONF_FLAG":        "ngxConfFlag",
	"NGX_HTTP_SRV_CONF":    "ngxHTTPSrvConf",
	"NGX_CONF_1MORE":       "ngxConf1More",
	"NGX_ANY_CONF":         "ngxAnyConf",
	"NGX_CONF_TAKE123":     "ngxConfTake123",
	"NGX_MAIN_CONF":        "ngxMainConf",
	"NGX_CONF_NOARGS":      "ngxConfNoArgs",
	"NGX_CONF_2MORE":       "ngxConf2More",
	"NGX_CONF_TAKE3":       "ngxConfTake3",
	"NGX_HTTP_SIF_CONF":    "ngxHTTPSifConf",
	"NGX_EVENT_CONF":       "ngxEventConf",
	"NGX_CONF_BLOCK":       "ngxConfBlock",
	"NGX_HTTP_LOC_CONF":    "ngxHTTPLocConf",
	"NGX_STREAM_SRV_CONF":  "ngxStreamSrvConf",
	"NGX_DIRECT_CONF":      "ngxDirectConf",
	"NGX_CONF_TAKE13":      "ngxConfTake13",
	"NGX_CONF_ANY":         "ngxConfAny",
	"NGX_CONF_TAKE4":       "ngxConfTake4",
	"NGX_CONF_TAKE5":       "ngxConfTake5",
	"NGX_CONF_TAKE6":       "ngxConfTake6",
	"NGX_CONF_TAKE7":       "ngxConfTake7",
}

//nolint:nonamedreturns
func masksFromFile(path string) (directive2Masks map[string][]Mask, err error) {
	directive2Masks = make(map[string][]Mask, 0)
	byteContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	strContent := string(byteContent)

	// Remove comments
	strContent = singleLineCommentExtracter.ReplaceAllString(strContent, "")
	strContent = multiLineCommentExtracter.ReplaceAllString(strContent, "")
	strContent = strings.ReplaceAll(strContent, "\r\n", "")
	strContent = strings.ReplaceAll(strContent, "\n", "")

	// Extract directives definition code blocks, each code block contains a list of directives definition
	blocks := directivesDefineBlockExtracter.FindAllStringSubmatch(strContent, -1)

	for _, block := range blocks {
		// Extract directives and their attributes in the code block, the first dimension of subBlocks
		// is index of directive, the second dimension is index of attributes
		subBlocks := singleDirectiveExtracter.FindAllStringSubmatch(block[2], -1)

		// Iterate through every directive
		for _, attributes := range subBlocks {
			// Extract attributes from the directive
			directiveName := strings.TrimSpace(attributes[1])
			directiveMask := strings.Split(attributes[2], "|")

			// Transfer C-style mask to go style
			for idx, ngxVarName := range directiveMask {
				goVarName, found := ngxVarNameToGo[strings.TrimSpace(ngxVarName)]
				if !found {
					return nil, fmt.Errorf("parsing directive %s, bitmask %s in source code not found in crossplane", directiveName, ngxVarName)
				}
				directiveMask[idx] = goVarName
			}

			directive2Masks[directiveName] = append(directive2Masks[directiveName], directiveMask)
		}
	}
	return directive2Masks, nil
}

//nolint:nonamedreturns
func getMasksFromPath(path string) (directive2Masks map[string][]Mask, err error) {
	directive2Masks = make(map[string][]Mask, 0)

	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the entry is a C/C++ file
		// Some dynamic modules are written in C++, like otel
		if d.IsDir() {
			return nil
		}

		if !(strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".cpp")) {
			return nil
		}

		directive2MasksInFile, err := masksFromFile(path)
		if err != nil {
			return err
		}

		for directive, masksInFile := range directive2MasksInFile {
			directive2Masks[directive] = append(directive2Masks[directive], masksInFile...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(directive2Masks) == 0 {
		return nil, errors.New("can't find any directives in the directory and subdirectories, please check the path")
	}

	return directive2Masks, nil
}

func genFromSrcCode(codePath string, mapVariableName string, matchFnName string, writer io.Writer,
	filter map[string]struct{}, override map[string][]Mask) error {
	directive2Masks, err := getMasksFromPath(codePath)
	if err != nil {
		return err
	}

	if len(filter) > 0 {
		for d := range directive2Masks {
			if _, found := filter[d]; found {
				delete(directive2Masks, d)
			}
		}
	}

	if override != nil {
		for d := range directive2Masks {
			if newMasks, found := override[d]; found {
				directive2Masks[d] = newMasks
			}
		}
	}

	err = supportFileTmpl.Execute(writer, supportFileTmplStruct{
		Directive2Masks: directive2Masks,
		MapVariableName: mapVariableName,
		MatchFnName:     matchFnName,
	})
	if err != nil {
		return err
	}

	return nil
}
