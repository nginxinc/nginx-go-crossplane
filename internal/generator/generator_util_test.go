/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var (
	update = flag.Bool("update", false,
		`update the expected output of these tests, 
 only use when the expected output is outdated and you are sure your output is correct`)
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func getProjectRootAbsPath() (string, error) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("can't get path of generator_util_test.go through runtime")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	// get the project root directory
	rootDir := filepath.Dir(absPath)
	rootDir = filepath.Dir(rootDir)
	rootDir = filepath.Dir(rootDir)

	return rootDir, nil
}

func getTestSrcCodePath(relativePath string) (string, error) {
	root, err := getProjectRootAbsPath()
	if err != nil {
		return "", err
	}
	return path.Join(root, "internal", "generator", "testdata", "source_codes", relativePath), nil
}

func getExpectedFilePath(relativePath string) (string, error) {
	root, err := getProjectRootAbsPath()
	if err != nil {
		return "", err
	}
	relativePath = strings.TrimSuffix(relativePath, ".c")
	relativePath = strings.TrimSuffix(relativePath, ".cpp")
	return path.Join(root, "internal", "generator", "testdata", "expected", relativePath), nil
}

//nolint:funlen,gocognit
func TestGenFromSrcCode(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		relativePath string
		wantErr      bool
		config       GenerateConfig
	}{
		"normalDirectiveDefinition_pass": {
			relativePath: "normalDefinition",
			config: GenerateConfig{
				DirectiveMapName: "myDirectives",
				MatchFuncName:    "MyMatchFn",
			},
			wantErr: false,
		},
		"unknownBitmask_fail": {
			relativePath: "unknownBitmask",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: true,
		},
		"noDirectivesDefinition_fail": {
			relativePath: "noDirectives",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: true,
		},
		// If one directive was defined in several files, we should keep all
		// of the bitmask definitions
		"directiveRepeatDefine_pass": {
			relativePath: "repeatDefine",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: false,
		},
		// If there are comments in directive definition, we should ignore them
		"commentsInDefinition_pass": {
			relativePath: "commentsInDefinition",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: false,
		},
		"genFromSingleFile_pass": {
			relativePath: "single_file.c",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: false,
		},
		"fullNgxBitmaskCover_pass": {
			relativePath: "fullNgxBitmaskCover",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: false,
		},
		"testFilter_pass": {
			relativePath: "filter",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
				Filter:           map[string]struct{}{"my_directive_2": {}, "my_directive_3": {}},
			},
		},
		"testOverride_pass": {
			relativePath: "override",
			config: GenerateConfig{
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
				Override: map[string][]Mask{
					"my_directive_1": {
						Mask{"ngxHTTPMainConf", "ngxConfTake1"},
						Mask{"ngxHTTPMainConf", "ngxConfTake2"},
					},
					"my_directive_3": {
						Mask{"ngxHTTPMainConf", "ngxConfTake2"},
						Mask{"ngxHTTPMainConf", "ngxConfTake3"},
					},
				},
			},
			wantErr: false,
		},
		"testFilterAndOverride_pass": {
			relativePath: "filterAndOverride",
			config: GenerateConfig{
				Filter: map[string]struct{}{
					"my_directive_1": {},
					"my_directive_2": {},
					"my_directive_3": {},
				},
				Override: map[string][]Mask{
					"my_directive_1": {
						Mask{"ngxHTTPMainConf", "ngxConfTake1"},
						Mask{"ngxHTTPMainConf", "ngxConfTake2"},
					},
					"my_directive_3": {
						Mask{"ngxHTTPMainConf", "ngxConfTake2"},
						Mask{"ngxHTTPMainConf", "ngxConfTake3"},
					},
					"my_directive_4": {
						Mask{"ngxHTTPMainConf", "ngxConfTake2"},
						Mask{"ngxHTTPMainConf", "ngxConfTake3"},
					},
				},
				DirectiveMapName: "directives",
				MatchFuncName:    "Match",
			},
			wantErr: false,
		},
		"withMatchFuncComment_pass": {
			relativePath: "withMatchFuncComment",
			config: GenerateConfig{
				DirectiveMapName: "myDirectives",
				MatchFuncName:    "MyMatchFn",
				MatchFuncComment: "This is a matchFunc",
			},
			wantErr: false,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var err error
			codePath, err := getTestSrcCodePath(tc.relativePath)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer

			err = genFromSrcCode(codePath, &buf, tc.config)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// If the testcase wants an error and there is an error, skip the output file validation,
			// since there may not be an output file while error occurs in generation.
			if err != nil {
				return
			}

			expectedFilePth, err := getExpectedFilePath(tc.relativePath)
			if err != nil {
				t.Fatal(err)
			}

			expectedFile, err := os.OpenFile(expectedFilePth, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if err = expectedFile.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			if *update {
				_, err = expectedFile.WriteString(buf.String())
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			expected, err := io.ReadAll(expectedFile)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, string(expected), buf.String())
		})
	}
}
