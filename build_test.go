/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type buildFixture struct {
	name     string
	options  BuildOptions
	parsed   Directives
	expected string
}

type buildFilesFixture struct {
	name     string
	options  BuildOptions
	payload  Payload
	expected string
}

type compareFixture struct {
	name    string
	options ParseOptions
}

//nolint:gochecknoglobals
var buildFixtures = []buildFixture{
	{
		name:    "nested-and-multiple-args",
		options: BuildOptions{},
		parsed: Directives{
			{
				Directive: "events",
				Args:      []string{},
				Block: Directives{
					{
						Directive: "worker_connections",
						Args:      []string{"1024"},
					},
				},
			},
			{
				Directive: "http",
				Args:      []string{},
				Block: Directives{
					{
						Directive: "server",
						Args:      []string{},
						Block: Directives{
							{
								Directive: "listen",
								Args:      []string{"127.0.0.1:8080"},
							},
							{
								Directive: "server_name",
								Args:      []string{"default_server"},
							},
							{
								Directive: "location",
								Args:      []string{"/"},
								Block: Directives{
									{
										Directive: "return",
										Args:      []string{"200", "foo bar baz"},
									},
								},
							},
						},
					},
				},
			},
		},
		expected: strings.Join([]string{
			"events {",
			"    worker_connections 1024;",
			"}",
			"http {",
			"    server {",
			"        listen 127.0.0.1:8080;",
			"        server_name default_server;",
			"        location / {",
			`            return 200 "foo bar baz";`,
			"        }",
			"    }",
			"}",
		}, "\n"),
	},
	{
		name:    "with-comments",
		options: BuildOptions{},
		parsed: Directives{
			{
				Directive: "events",
				Line:      1,
				Args:      []string{},
				Block: Directives{
					{
						Directive: "worker_connections",
						Line:      2,
						Args:      []string{"1024"},
					},
				},
			},
			{
				Directive: "#",
				Line:      4,
				Args:      []string{},
				Comment:   pStr("comment"),
			},
			{
				Directive: "http",
				Line:      5,
				Args:      []string{},
				Block: Directives{
					{
						Directive: "server",
						Line:      6,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "listen",
								Line:      7,
								Args:      []string{"127.0.0.1:8080"},
							},
							{
								Directive: "#",
								Line:      7,
								Args:      []string{},
								Comment:   pStr("listen"),
							},
							{
								Directive: "server_name",
								Line:      8,
								Args:      []string{"default_server"},
							},
							{
								Directive: "location",
								Line:      9,
								Args:      []string{"/"},
								Block: Directives{
									{
										Directive: "#",
										Line:      9,
										Args:      []string{},
										Comment:   pStr("# this is brace"),
									},
									{
										Directive: "#",
										Line:      10,
										Args:      []string{},
										Comment:   pStr(" location /"),
									},
									{
										Directive: "#",
										Line:      11,
										Args:      []string{},
										Comment:   pStr(" is here"),
									},
									{
										Directive: "return",
										Line:      12,
										Args:      []string{"200", "foo bar baz"},
									},
								},
							},
						},
					},
				},
			},
		},
		expected: strings.Join([]string{
			"events {",
			"    worker_connections 1024;",
			"}",
			"#comment",
			"http {",
			"    server {",
			"        listen 127.0.0.1:8080; #listen",
			"        server_name default_server;",
			"        location / { ## this is brace",
			"            # location /",
			"            # is here",
			`            return 200 "foo bar baz";`,
			"        }",
			"    }",
			"}",
		}, "\n"),
	},
	{
		name:    "starts-with-comments",
		options: BuildOptions{},
		parsed: Directives{
			{
				Directive: "#",
				Line:      1,
				Args:      []string{},
				Comment:   pStr(" foo"),
			},
			{
				Directive: "user",
				Line:      5,
				Args:      []string{"root"},
			},
		},
		expected: "# foo\nuser root;",
	},
	{
		name:    "with-quoted-unicode",
		options: BuildOptions{},
		parsed: Directives{
			{
				Directive: "env",
				Line:      1,
				Args:      []string{"русский текст"},
			},
		},
		expected: `env "русский текст";`,
	},
	{
		name:    "multiple-comments-on-one-line",
		options: BuildOptions{},
		parsed: Directives{
			{
				Directive: "#",
				Line:      1,
				Args:      []string{},
				Comment:   pStr("comment1"),
			},
			{
				Directive: "user",
				Line:      2,
				Args:      []string{"root"},
			},
			{
				Directive: "#",
				Line:      2,
				Args:      []string{},
				Comment:   pStr("comment2"),
			},
			{
				Directive: "#",
				Line:      2,
				Args:      []string{},
				Comment:   pStr("comment3"),
			},
		},
		expected: "#comment1\nuser root; #comment2 #comment3",
	},
}

func TestBuild(t *testing.T) {
	t.Parallel()
	for _, fixture := range buildFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			if err := Build(&buf, Config{Parsed: fixture.parsed}, &fixture.options); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			if got != fixture.expected {
				t.Fatalf("expected: %#v\nbut got: %#v", fixture.expected, got)
			}
		})
	}
}

//nolint:gochecknoglobals
var buildFilesFixtures = []buildFilesFixture{
	{
		name:    "with-missing-status-and-errors",
		options: BuildOptions{},
		payload: Payload{
			Config: []Config{
				{
					File: "nginx.conf",
					Parsed: Directives{
						{
							Directive: "user",
							Line:      1,
							Args:      []string{"nginx"},
						},
					},
				},
			},
		},
		expected: "user nginx;\n",
	},
	{
		name:    "with-unicode",
		options: BuildOptions{},
		payload: Payload{
			Status: "ok",
			Errors: []PayloadError{},
			Config: []Config{
				{
					File:   "nginx.conf",
					Status: "ok",
					Errors: []ConfigError{},
					Parsed: Directives{
						{
							Directive: "user",
							Line:      1,
							Args:      []string{"測試"},
						},
					},
				},
			},
		},
		expected: "user 測試;\n",
	},
}

func TestBuildFiles(t *testing.T) {
	t.Parallel()
	for _, fixture := range buildFilesFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			fixture := fixture
			tmpdir := t.TempDir()

			if err := BuildFiles(fixture.payload, tmpdir, &fixture.options); err != nil {
				t.Fatal(err)
			}

			content, err := os.ReadFile(filepath.Join(tmpdir, "nginx.conf"))
			if err != nil {
				t.Fatal(err)
			}

			got := string(content)
			if got != fixture.expected {
				t.Fatalf("expected: %#v\nbut got: %#v", fixture.expected, got)
			}
		})
	}
}

//nolint:gochecknoglobals
var compareFixtures = []compareFixture{
	{"simple", ParseOptions{}},
	{"messy", ParseOptions{}},
	{"with-comments", ParseOptions{ParseComments: true}},
	{"empty-value-map", ParseOptions{}},
	{"russian-text", ParseOptions{}},
	{"quoted-right-brace", ParseOptions{}},
	{"directive-with-space", ParseOptions{ErrorOnUnknownDirectives: true}},
	{"empty-config", ParseOptions{}},
}

//nolint:gocognit
func TestCompareParsedAndBuilt(t *testing.T) {
	t.Parallel()
	for _, fixture := range compareFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			tmpdir := t.TempDir()

			origPayload, err2 := Parse(getTestConfigPath(fixture.name, "nginx.conf"), &fixture.options)
			if err2 != nil {
				t.Fatal(err2)
			}

			var build1Buffer bytes.Buffer
			if err := Build(&build1Buffer, origPayload.Config[0], &BuildOptions{}); err != nil {
				t.Fatal(err)
			}
			build1File := filepath.Join(tmpdir, "build1.conf")
			build1Config := build1Buffer.Bytes()
			if err := os.WriteFile(build1File, build1Config, os.ModePerm); err != nil {
				t.Fatal(err)
			}
			build1Payload, err2 := Parse(build1File, &fixture.options)
			if err2 != nil {
				t.Fatal(err2)
			}

			if !equalPayloads(t, *origPayload, *build1Payload) {
				b1, _ := json.Marshal(origPayload)
				b2, _ := json.Marshal(build1Payload)
				if string(b1) != string(b2) {
					t.Fatalf("expected: %s\nbut got: %s", b1, b2)
				}
			}

			var build2Buffer bytes.Buffer
			if err := Build(&build2Buffer, build1Payload.Config[0], &BuildOptions{}); err != nil {
				t.Fatal(err)
			}
			build2File := filepath.Join(tmpdir, "build2.conf")
			build2Config := build2Buffer.Bytes()
			if err := os.WriteFile(build2File, build2Config, os.ModePerm); err != nil {
				t.Fatal(err)
			}
			build2Payload, err2 := Parse(build2File, &fixture.options)
			if err2 != nil {
				t.Fatal(err2)
			}

			if !equalPayloads(t, *build1Payload, *build2Payload) {
				b1, _ := json.Marshal(build1Payload)
				b2, _ := json.Marshal(build2Payload)
				if string(b1) != string(b2) {
					t.Fatalf("expected: %s\nbut got: %s", b1, b2)
				}
			}
		})
	}
}

func equalPayloads(t *testing.T, p1, p2 Payload) bool {
	return p1.Status == p2.Status &&
		equalPayloadErrors(t, p1.Errors, p2.Errors) &&
		equalPayloadConfigs(p1.Config, p2.Config)
}

func equalPayloadErrors(t *testing.T, e1, e2 []PayloadError) bool {
	if len(e1) != len(e2) {
		return false
	}
	for i := 0; i < len(e1); i++ {
		if (e1[i].File != e2[i].File) ||
			(e1[i].Error != nil && e2[i].Error != nil && e1[i].Error.Error() != e2[i].Error.Error()) ||
			(e1[i].Line == nil) != (e2[i].Line == nil) ||
			(e1[i].Line != nil && *e1[i].Line != *e2[i].Line) {
			t.Log(e1[i].Error.Error())
			t.Log(e2[i].Error.Error())
			return false
		}
	}
	return true
}

func equalPayloadConfigs(c1, c2 []Config) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := 0; i < len(c1); i++ {
		if !equalConfigs(c1[i], c2[i]) {
			return false
		}
	}
	return true
}

func equalConfigs(c1, c2 Config) bool {
	return c1.Status == c2.Status &&
		equalConfigErrors(c1.Errors, c2.Errors) &&
		equalBlocks(c1.Parsed, c2.Parsed)
}

func equalConfigErrors(e1, e2 []ConfigError) bool {
	if len(e1) != len(e2) {
		return false
	}
	for i := 0; i < len(e1); i++ {
		if (e1[i].Error != nil && e2[i].Error != nil && e1[i].Error.Error() != e2[i].Error.Error()) ||
			(e1[i].Line == nil) != (e2[i].Line == nil) ||
			(e1[i].Line != nil && *e1[i].Line != *e2[i].Line) {
			return false
		}
	}
	return true
}

func equalBlocks(b1, b2 Directives) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := 0; i < len(b1); i++ {
		if !equalDirectives(b1[i], b2[i]) {
			return false
		}
	}
	return true
}

func equalArgs(b1, b2 []string) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := 0; i < len(b1); i++ {
		if Enquote(b1[i]) != Enquote(b2[i]) {
			return false
		}
	}
	return true
}

func equalIncludes(b1, b2 []int) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

func equalDirectives(d1, d2 *Directive) bool {
	if d1.Directive != d2.Directive ||
		d1.File != d2.File ||
		!equalArgs(d1.Args, d2.Args) ||
		!equalIncludes(d1.Includes, d2.Includes) ||
		!equalBlocks(d1.Block, d2.Block) ||
		(d1.Comment == nil) != (d2.Comment == nil) ||
		(d1.Comment != nil && *d1.Comment != *d2.Comment) {
		return false
	}

	return true
}

func TestBuildInto(t *testing.T) {
	t.Parallel()
	for _, fixture := range buildFilesFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			fixture := fixture
			sc := new(StringsCreator)

			if err := BuildInto(&fixture.payload, sc, &fixture.options); err != nil {
				t.Fatal(err)
			}

			if len(sc.Files) == 0 {
				t.Fatal("no buffer written")
			}

			got := sc.Files[0].String()
			if got != fixture.expected {
				t.Fatalf("expected: %#v\nbut got: %#v", fixture.expected, got)
			}
		})
	}
}
