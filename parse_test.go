/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type parseFixture struct {
	name     string
	suffix   string
	options  ParseOptions
	expected Payload
}

func pInt(i int) *int {
	return &i
}

func pStr(s string) *string {
	return &s
}

func noSuchFileErrMsg() string {
	if runtime.GOOS == "windows" {
		return "The system cannot find the file specified."
	}
	return "no such file or directory"
}

func getTestConfigPath(parts ...string) string {
	return filepath.Join("testdata", "configs", filepath.Join(parts...))
}

//nolint:gochecknoglobals
var lua = &Lua{}

//nolint:gochecknoglobals,exhaustruct
var parseFixtures = []parseFixture{
	{"includes-regular", "", ParseOptions{}, Payload{
		Status: "failed",
		Errors: []PayloadError{
			{
				File: getTestConfigPath("includes-regular", "conf.d", "server.conf"),
				Error: &ParseError{
					fmt.Sprintf("open %s: %s",
						getTestConfigPath("includes-regular", "bar.conf"),
						noSuchFileErrMsg(),
					),
					pStr(getTestConfigPath("includes-regular", "conf.d", "server.conf")),
					pInt(5),
					"include bar.conf",
					"server",
					nil,
				},
				Line: pInt(5),
			},
		},
		Config: []Config{
			{
				File:   getTestConfigPath("includes-regular", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      2,
						Block: Directives{
							{
								Directive: "include",
								Args:      []string{"conf.d/server.conf"},
								Line:      3,
								Includes:  []int{1},
							},
						},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-regular", "conf.d", "server.conf"),
				Status: "failed",
				Errors: []ConfigError{
					{
						Error: &ParseError{
							fmt.Sprintf("open %s: %s",
								getTestConfigPath("includes-regular", "bar.conf"),
								noSuchFileErrMsg(),
							),
							pStr(getTestConfigPath("includes-regular", "conf.d", "server.conf")),
							pInt(5),
							"include bar.conf",
							"server",
							nil,
						},
						Line: pInt(5),
					},
				},
				Parsed: Directives{
					{
						Directive: "server",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "listen",
								Args:      []string{"127.0.0.1:8080"},
								Line:      2,
							},
							{
								Directive: "server_name",
								Args:      []string{"default_server"},
								Line:      3,
							},
							{
								Directive: "include",
								Args:      []string{"foo.conf"},
								Line:      4,
								Includes:  []int{2},
							},
							{
								Directive: "include",
								Args:      []string{"bar.conf"},
								Line:      5,
							},
						},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-regular", "foo.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "location",
						Args:      []string{"/foo"},
						Line:      1,
						Block: Directives{
							{
								Directive: "return",
								Args:      []string{"200", "foo"},
								Line:      2,
							},
						},
					},
				},
			},
		},
	}},
	{"includes-regular", "-single-file", ParseOptions{SingleFile: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("includes-regular", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block:     Directives{},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      2,
						Block: Directives{
							{
								Directive: "include",
								Args:      []string{"conf.d/server.conf"},
								Line:      3,
								// no Includes key
							},
						},
					},
				},
			},
			// single config parsed
		},
	}},
	{"includes-globbed", "", ParseOptions{}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("includes-globbed", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
					},
					{
						Directive: "include",
						Args:      []string{"http.conf"},
						Line:      2,
						Includes:  []int{1},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-globbed", "http.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "include",
								Args:      []string{"servers/*.conf"},
								Line:      2,
								Includes:  []int{2, 3},
							},
						},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-globbed", "servers", "server1.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "server",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "listen",
								Args:      []string{"8080"},
								Line:      2,
							},
							{
								Directive: "include",
								Args:      []string{"locations/*.conf"},
								Line:      3,
								Includes:  []int{4, 5},
							},
						},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-globbed", "servers", "server2.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "server",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "listen",
								Args:      []string{"8081"},
								Line:      2,
							},
							{
								Directive: "include",
								Args:      []string{"locations/*.conf"},
								Line:      3,
								Includes:  []int{4, 5},
							},
						},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-globbed", "locations", "location1.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "location",
						Args:      []string{"/foo"},
						Line:      1,
						Block: Directives{
							{
								Directive: "return",
								Args:      []string{"200", "foo"},
								Line:      2,
							},
						},
					},
				},
			},
			{
				File:   getTestConfigPath("includes-globbed", "locations", "location2.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "location",
						Args:      []string{"/bar"},
						Line:      1,
						Block: Directives{
							{
								Directive: "return",
								Args:      []string{"200", "bar"},
								Line:      2,
							},
						},
					},
				},
			},
		},
	}},
	{"includes-globbed", "-combine-configs", ParseOptions{CombineConfigs: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("includes-globbed", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						File:      getTestConfigPath("includes-globbed", "nginx.conf"),
						Block:     Directives{},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						File:      getTestConfigPath("includes-globbed", "http.conf"),
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      1,
								File:      getTestConfigPath("includes-globbed", "servers", "server1.conf"),
								Block: Directives{
									{
										Directive: "listen",
										Args:      []string{"8080"},
										Line:      2,
										File:      getTestConfigPath("includes-globbed", "servers", "server1.conf"),
									},
									{
										Directive: "location",
										Args:      []string{"/foo"},
										Line:      1,
										File:      getTestConfigPath("includes-globbed", "locations", "location1.conf"),
										Block: Directives{
											{
												Directive: "return",
												Args:      []string{"200", "foo"},
												Line:      2,
												File:      getTestConfigPath("includes-globbed", "locations", "location1.conf"),
											},
										},
									},
									{
										Directive: "location",
										Args:      []string{"/bar"},
										Line:      1,
										File:      getTestConfigPath("includes-globbed", "locations", "location2.conf"),
										Block: Directives{
											{
												Directive: "return",
												Args:      []string{"200", "bar"},
												Line:      2,
												File:      getTestConfigPath("includes-globbed", "locations", "location2.conf"),
											},
										},
									},
								},
							},
							{
								Directive: "server",
								Args:      []string{},
								Line:      1,
								File:      getTestConfigPath("includes-globbed", "servers", "server2.conf"),
								Block: Directives{
									{
										Directive: "listen",
										Args:      []string{"8081"},
										Line:      2,
										File:      getTestConfigPath("includes-globbed", "servers", "server2.conf"),
									},
									{
										Directive: "location",
										Args:      []string{"/foo"},
										Line:      1,
										File:      getTestConfigPath("includes-globbed", "locations", "location1.conf"),
										Block: Directives{
											{
												Directive: "return",
												Args:      []string{"200", "foo"},
												Line:      2,
												File:      getTestConfigPath("includes-globbed", "locations", "location1.conf"),
											},
										},
									},
									{
										Directive: "location",
										Args:      []string{"/bar"},
										Line:      1,
										File:      getTestConfigPath("includes-globbed", "locations", "location2.conf"),
										Block: Directives{
											{
												Directive: "return",
												Args:      []string{"200", "bar"},
												Line:      2,
												File:      getTestConfigPath("includes-globbed", "locations", "location2.conf"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"simple", "-ignore-directives-1", ParseOptions{IgnoreDirectives: []string{"listen", "server_name"}}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("simple", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "worker_connections",
								Args:      []string{"1024"},
								Line:      2,
							},
						},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      5,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      6,
								Block: Directives{
									{
										Directive: "location",
										Args:      []string{"/"},
										Line:      9,
										Block: Directives{
											{
												Directive: "return",
												Args:      []string{"200", "foo bar baz"},
												Line:      10,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"simple", "-ignore-directives-2", ParseOptions{IgnoreDirectives: []string{"events", "server"}}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("simple", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      5,
						Block:     Directives{},
					},
				},
			},
		},
	}},
	{"simple-variable-with-braces", "-ignore-directives-1", ParseOptions{IgnoreDirectives: []string{"listen", "server_name"}}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("simple-variable-with-braces", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "worker_connections",
								Args:      []string{"1024"},
								Line:      2,
							},
						},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      5,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      6,
								Block: Directives{
									{
										Directive: "location",
										Args:      []string{"/proxy"},
										Line:      9,
										Block: Directives{
											{
												Directive: "set",
												Args:      []string{"$backend_protocol", "http"},
												Line:      10,
											},
											{
												Directive: "set",
												Args:      []string{"$backend_host", "bar"},
												Line:      11,
											},
											{
												Directive: "set",
												Args:      []string{"$foo", ""},
												Line:      12,
											},
											{
												Directive: "proxy_pass",
												Args:      []string{"$backend_protocol://$backend_host${foo}"},
												Line:      13,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"with-comments", "-true", ParseOptions{ParseComments: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("with-comments", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "worker_connections",
								Args:      []string{"1024"},
								Line:      2,
							},
						},
					},
					{
						Directive: "#",
						Args:      []string{},
						Line:      4,
						Comment:   pStr("comment"),
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      5,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      6,
								Block: Directives{
									{
										Directive: "listen",
										Args:      []string{"127.0.0.1:8080"},
										Line:      7,
									},
									{
										Directive: "#",
										Args:      []string{},
										Line:      7,
										Comment:   pStr("listen"),
									},
									{
										Directive: "server_name",
										Args:      []string{"default_server"},
										Line:      8,
									},
									{
										Directive: "location",
										Args:      []string{"/"},
										Line:      9,
										Block: Directives{
											{
												Directive: "#",
												Args:      []string{},
												Line:      9,
												Comment:   pStr("# this is brace"),
											},
											{
												Directive: "#",
												Args:      []string{},
												Line:      10,
												Comment:   pStr(" location /"),
											},
											{
												Directive: "return",
												Args:      []string{"200", "foo bar baz"},
												Line:      11,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"with-comments", "-false", ParseOptions{ParseComments: false}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("with-comments", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "worker_connections",
								Args:      []string{"1024"},
								Line:      2,
							},
						},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      5,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      6,
								Block: Directives{
									{
										Directive: "listen",
										Args:      []string{"127.0.0.1:8080"},
										Line:      7,
									},
									{
										Directive: "server_name",
										Args:      []string{"default_server"},
										Line:      8,
									},
									{
										Directive: "location",
										Args:      []string{"/"},
										Line:      9,
										Block: Directives{
											{
												Directive: "return",
												Args:      []string{"200", "foo bar baz"},
												Line:      11,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"spelling-mistake", "", ParseOptions{ParseComments: true, ErrorOnUnknownDirectives: true}, Payload{
		Status: "failed",
		Errors: []PayloadError{
			{
				File: getTestConfigPath("spelling-mistake", "nginx.conf"),
				Error: &ParseError{
					`unknown directive "proxy_passs"`,
					pStr(getTestConfigPath("spelling-mistake", "nginx.conf")),
					pInt(7),
					"proxy_passs http://foo.bar",
					"location",
					nil,
				},
				Line: pInt(7),
			},
		},
		Config: []Config{
			{
				File:   getTestConfigPath("spelling-mistake", "nginx.conf"),
				Status: "failed",
				Errors: []ConfigError{
					{
						Error: &ParseError{
							`unknown directive "proxy_passs"`,
							pStr(getTestConfigPath("spelling-mistake", "nginx.conf")),
							pInt(7),
							"proxy_passs http://foo.bar",
							"location",
							nil,
						},
						Line: pInt(7),
					},
				},
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block:     Directives{},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      3,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      4,
								Block: Directives{
									{
										Directive: "location",
										Args:      []string{"/"},
										Line:      5,
										Block: Directives{
											{
												Directive: "#",
												Args:      []string{},
												Line:      6,
												Comment:   pStr("directive is misspelled"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"missing-semicolon-above", "", ParseOptions{}, Payload{
		Status: "failed",
		Errors: []PayloadError{
			{
				File: getTestConfigPath("missing-semicolon-above", "nginx.conf"),
				Error: &ParseError{
					`directive "proxy_pass" is not terminated by ";"`,
					pStr(getTestConfigPath("missing-semicolon-above", "nginx.conf")),
					pInt(4),
					`proxy_pass http://is.broken.example`,
					`location`,
					nil,
				},
				Line: pInt(4),
			},
		},
		Config: []Config{
			{
				File:   getTestConfigPath("missing-semicolon-above", "nginx.conf"),
				Status: "failed",
				Errors: []ConfigError{
					{
						Error: &ParseError{
							`directive "proxy_pass" is not terminated by ";"`,
							pStr(getTestConfigPath("missing-semicolon-above", "nginx.conf")),
							pInt(4),
							`proxy_pass http://is.broken.example`,
							"location",
							nil,
						},
						Line: pInt(4),
					},
				},
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      2,
								Block: Directives{
									{
										Directive: "location",
										Args:      []string{"/is-broken"},
										Line:      3,
										Block:     Directives{},
									},
									{
										Directive: "location",
										Args:      []string{"/not-broken"},
										Line:      6,
										Block: Directives{
											{
												Directive: "proxy_pass",
												Args:      []string{"http://not.broken.example"},
												Line:      7,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"missing-semicolon-below", "", ParseOptions{}, Payload{
		Status: "failed",
		Errors: []PayloadError{
			{
				File: getTestConfigPath("missing-semicolon-below", "nginx.conf"),
				Error: &ParseError{
					`directive "proxy_pass" is not terminated by ";"`,
					pStr(getTestConfigPath("missing-semicolon-below", "nginx.conf")),
					pInt(7),
					`proxy_pass http://is.broken.example`,
					"location",
					nil,
				},
				Line: pInt(7),
			},
		},
		Config: []Config{
			{
				File:   getTestConfigPath("missing-semicolon-below", "nginx.conf"),
				Status: "failed",
				Errors: []ConfigError{
					{
						Error: &ParseError{
							`directive "proxy_pass" is not terminated by ";"`,
							pStr(getTestConfigPath("missing-semicolon-below", "nginx.conf")),
							pInt(7),
							`proxy_pass http://is.broken.example`,
							"location",
							nil,
						},
						Line: pInt(7),
					},
				},
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "server",
								Args:      []string{},
								Line:      2,
								Block: Directives{
									{
										Directive: "location",
										Args:      []string{"/not-broken"},
										Line:      3,
										Block: Directives{
											{
												Directive: "proxy_pass",
												Args:      []string{"http://not.broken.example"},
												Line:      4,
											},
										},
									},
									{
										Directive: "location",
										Args:      []string{"/is-broken"},
										Line:      6,
										Block:     Directives{},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"comments-between-args", "", ParseOptions{ParseComments: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("comments-between-args", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "#",
								Args:      []string{},
								Line:      1,
								Comment:   pStr("comment 1"),
							},
							{
								Directive: "log_format",
								Args:      []string{"\\#arg\\ 1", "#arg 2"},
								Line:      2,
							},
							{
								Directive: "#",
								Args:      []string{},
								Line:      2,
								Comment:   pStr("comment 2"),
							},
							{
								Directive: "#",
								Args:      []string{},
								Line:      2,
								Comment:   pStr("comment 3"),
							},
							{
								Directive: "#",
								Args:      []string{},
								Line:      2,
								Comment:   pStr("comment 4"),
							},
							{
								Directive: "#",
								Args:      []string{},
								Line:      2,
								Comment:   pStr("comment 5"),
							},
						},
					},
				},
			},
		},
	}},
	{"comments-between-args-disable-parse", "", ParseOptions{ParseComments: false}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("comments-between-args-disable-parse", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "log_format",
								Args:      []string{"\\#arg\\ 1", "#arg 2"},
								Line:      2,
							},
						},
					},
				},
			},
		},
	}},
	{"premature-eof", "", ParseOptions{}, Payload{
		Status: "failed",
		Errors: []PayloadError{
			{
				File: getTestConfigPath("premature-eof", "nginx.conf"),
				Error: &ParseError{
					`premature end of file`,
					pStr(getTestConfigPath("premature-eof", "nginx.conf")),
					pInt(3),
					"",
					"",
					ErrPrematureLexEnd,
				},
				Line: pInt(3),
			},
		},
		Config: []Config{
			{
				File:   getTestConfigPath("premature-eof", "nginx.conf"),
				Status: "failed",
				Errors: []ConfigError{
					{
						Error: &ParseError{
							`premature end of file`,
							pStr(getTestConfigPath("premature-eof", "nginx.conf")),
							pInt(3),
							"",
							"",
							ErrPrematureLexEnd,
						},
						Line: pInt(3),
					},
				},
				Parsed: Directives{},
			},
		},
	}},
	{"directive-with-space", "", ParseOptions{ErrorOnUnknownDirectives: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("directive-with-space", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block:     Directives{},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      3,
						Block: Directives{
							{
								Directive: "map",
								Args:      []string{"$http_user_agent", "$mobile"},
								Line:      4,
								Block: Directives{
									{
										Directive: "default",
										Args:      []string{"0"},
										Line:      5,
										Block:     Directives{},
									},
									{
										Directive: "~Opera Mini",
										Args:      []string{"1"},
										Line:      6,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "charset_map",
								Args:      []string{"koi8-r", "utf-8"},
								Line:      9,
								Block: Directives{
									{
										Directive: "C0",
										Args:      []string{"D18E"},
										Line:      10,
										Block:     Directives{},
									},
									{
										Directive: "C1",
										Args:      []string{"D0B0"},
										Line:      11,
										Block:     Directives{},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"invalid-map", "", ParseOptions{ErrorOnUnknownDirectives: true}, Payload{
		Status: "failed",
		Errors: []PayloadError{
			{
				File: getTestConfigPath("invalid-map", "nginx.conf"),
				Error: &ParseError{
					`unexpected "{"`,
					pStr(getTestConfigPath("invalid-map", "nginx.conf")),
					pInt(7),
					"i_am_lost ",
					"map",
					nil,
				},
				Line: pInt(7),
			},
			{
				File: getTestConfigPath("invalid-map", "nginx.conf"),
				Error: &ParseError{
					`invalid number of parameters`,
					pStr(getTestConfigPath("invalid-map", "nginx.conf")),
					pInt(10),
					"too many params",
					"map",
					nil,
				},
				Line: pInt(10),
			},
			{
				File: getTestConfigPath("invalid-map", "nginx.conf"),
				Error: &ParseError{
					`invalid number of parameters`,
					pStr(getTestConfigPath("invalid-map", "nginx.conf")),
					pInt(14),
					"C0 ",
					"charset_map",
					nil,
				},
				Line: pInt(14),
			},
		},
		Config: []Config{
			{
				File:   getTestConfigPath("invalid-map", "nginx.conf"),
				Status: "failed",
				Errors: []ConfigError{
					{
						Error: &ParseError{
							`unexpected "{"`,
							pStr(getTestConfigPath("invalid-map", "nginx.conf")),
							pInt(7),
							"i_am_lost ",
							"map",
							nil,
						},
						Line: pInt(7),
					},
					{
						Error: &ParseError{
							`invalid number of parameters`,
							pStr(getTestConfigPath("invalid-map", "nginx.conf")),
							pInt(10),
							"too many params",
							"map",
							nil,
						},
						Line: pInt(10),
					},
					{
						Error: &ParseError{
							`invalid number of parameters`,
							pStr(getTestConfigPath("invalid-map", "nginx.conf")),
							pInt(14),
							"C0 ",
							"charset_map",
							nil,
						},
						Line: pInt(14),
					},
				},
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block:     Directives{},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      3,
						Block: Directives{
							{
								Directive: "map",
								Args:      []string{"$http_user_agent", "$mobile"},
								Line:      4,
								Block: Directives{
									{
										Directive: "default",
										Args:      []string{"0"},
										Line:      5,
										Block:     Directives{},
									},
									{
										Directive: "~Opera Mini",
										Args:      []string{"1"},
										Line:      6,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "charset_map",
								Args:      []string{"koi8-r", "utf-8"},
								Line:      13,
								Block:     Directives{},
							},
						},
					},
				},
			},
		},
	}},
	{"geo", "", ParseOptions{ErrorOnUnknownDirectives: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("geo", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "events",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "worker_connections",
								Args:      []string{"1024"},
								Line:      2,
							},
						},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      5,
						Block: Directives{
							{
								Directive: "geo",
								Args:      []string{"$geo"},
								Line:      6,
								Block: Directives{
									{
										Directive: "ranges",
										Args:      []string{},
										Line:      7,
										Block:     Directives{},
									},
									{
										Directive: "default",
										Args:      []string{"0"},
										Line:      8,
										Block:     Directives{},
									},
									{
										Directive: "192.168.1.0/24",
										Args:      []string{"1"},
										Line:      9,
										Block:     Directives{},
									},
									{
										Directive: "127.0.0.1",
										Args:      []string{"2"},
										Line:      10,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "server",
								Args:      []string{},
								Line:      12,
								Block: Directives{
									{
										Directive: "listen",
										Args:      []string{"127.0.0.1:8080"},
										Line:      13,
										Block:     Directives{},
									},
									{
										Directive: "server_name",
										Args:      []string{"default_server"},
										Line:      14,
										Block:     Directives{},
									},
									{
										Directive: "location",
										Args:      []string{"/"},
										Line:      15,
										Block: Directives{
											{
												Directive: "if",
												Args:      []string{"$geo", "=", "2"},
												Line:      16,
												Block: Directives{
													{
														Directive: "return",
														Args:      []string{"403"},
														Line:      17,
														Block:     Directives{},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"types", "", ParseOptions{ErrorOnUnknownDirectives: true}, Payload{
		Status: "ok",
		Config: []Config{
			{
				File:   getTestConfigPath("types", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
							{
								Directive: "types",
								Line:      2,
								Block: Directives{
									{
										Directive: "text/html",
										Args:      []string{"html", "htm", "shtml"},
										Line:      3,
										Block:     Directives{},
									},
									{
										Directive: "text/css",
										Args:      []string{"css"},
										Line:      4,
										Block:     Directives{},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"nap-waf-v4", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchAppProtectWAFv4},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("nap-waf-v4", "nginx.conf"),
				Status: "ok",
				Errors: []ConfigError{},
				Parsed: Directives{
					{
						Directive: "user",
						Args:      []string{"nginx"},
						Line:      1,
					},
					{
						Directive: "worker_processes",
						Line:      2,
						Args:      []string{"4"},
					},
					{
						Directive: "load_module",
						Line:      4,
						Args:      []string{"modules/ngx_http_app_protect_module.so"},
					},
					{
						Directive: "error_log",
						Line:      6,
						Args:      []string{"/var/log/nginx/error.log", "debug"},
					},
					{
						Directive: "events",
						Line:      8,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "worker_connections",
								Line:      9,
								Args:      []string{"65536"},
							},
						},
					},
					{
						Directive: "http",
						Line:      12,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "include",
								Line:      13,
								Args:      []string{"/etc/nginx/mime.types"},
							},
							{
								Directive: "default_type",
								Line:      14,
								Args:      []string{"application/octet-stream"},
							},
							{
								Directive: "sendfile",
								Line:      15,
								Args:      []string{"on"},
							},
							{
								Directive: "keepalive_timeout",
								Line:      16,
								Args:      []string{"65"},
							},
							{
								Directive: "app_protect_enable",
								Line:      18,
								Args:      []string{"on"},
							},
							{
								Directive: "app_protect_policy_file",
								Line:      19,
								Args: []string{
									"/etc/app_protect/conf/NginxDefaultPolicy.json",
								},
							},
							{
								Directive: "app_protect_security_log_enable",
								Line:      20,
								Args:      []string{"on"},
							},
							{
								Directive: "app_protect_security_log",
								Line:      21,
								Args: []string{
									"/etc/app_protect/conf/log_default.json",
									"syslog:server=127.0.0.1:515",
								},
							},
							{
								Directive: "server",
								Line:      23,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      24,
										Args:      []string{"80"},
									},
									{
										Directive: "server_name",
										Line:      25,
										Args:      []string{"localhost"},
									},
									{
										Directive: "proxy_http_version",
										Line:      26,
										Args:      []string{"1.1"},
									},
									{
										Directive: "location",
										Line:      28,
										Args:      []string{"/"},
										Block: Directives{
											{
												Directive: "client_max_body_size",
												Line:      29,
												Args:      []string{"0"},
											},
											{
												Directive: "default_type",
												Line:      30,
												Args:      []string{"text/html"},
											},
											{
												Directive: "proxy_pass",
												Line:      31,
												Args:      []string{"http://172.29.38.211:80$request_uri"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"nap-waf-v5", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchAppProtectWAFv5},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("nap-waf-v5", "nginx.conf"),
				Status: "ok",
				Errors: []ConfigError{},
				Parsed: Directives{
					{
						Directive: "user",
						Args:      []string{"nginx"},
						Line:      1,
					},
					{
						Directive: "worker_processes",
						Line:      2,
						Args:      []string{"4"},
					},
					{
						Directive: "load_module",
						Line:      4,
						Args:      []string{"modules/ngx_http_app_protect_module.so"},
					},
					{
						Directive: "error_log",
						Line:      6,
						Args:      []string{"/var/log/nginx/error.log", "debug"},
					},
					{
						Directive: "events",
						Line:      8,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "worker_connections",
								Line:      9,
								Args:      []string{"65536"},
							},
						},
					},
					{
						Directive: "http",
						Line:      12,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "include",
								Line:      13,
								Args:      []string{"/etc/nginx/mime.types"},
							},
							{
								Directive: "default_type",
								Line:      14,
								Args:      []string{"application/octet-stream"},
							},
							{
								Directive: "sendfile",
								Line:      15,
								Args:      []string{"on"},
							},
							{
								Directive: "keepalive_timeout",
								Line:      16,
								Args:      []string{"65"},
							},
							{
								Directive: "app_protect_enforcer_address",
								Line:      18,
								Args:      []string{"127.0.0.1:50000"},
							},
							{
								Directive: "app_protect_enable",
								Line:      19,
								Args:      []string{"on"},
							},
							{
								Directive: "app_protect_policy_file",
								Line:      20,
								Args: []string{
									"/policies/policy1.tgz",
								},
							},
							{
								Directive: "app_protect_security_log_enable",
								Line:      21,
								Args:      []string{"on"},
							},
							{
								Directive: "app_protect_security_log",
								Line:      22,
								Args: []string{
									"log_all",
									"syslog:server=127.0.0.1:515",
								},
							},
							{
								Directive: "server",
								Line:      24,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      25,
										Args:      []string{"80"},
									},
									{
										Directive: "server_name",
										Line:      26,
										Args:      []string{"localhost"},
									},
									{
										Directive: "proxy_http_version",
										Line:      27,
										Args:      []string{"1.1"},
									},
									{
										Directive: "location",
										Line:      29,
										Args:      []string{"/"},
										Block: Directives{
											{
												Directive: "client_max_body_size",
												Line:      30,
												Args:      []string{"0"},
											},
											{
												Directive: "default_type",
												Line:      31,
												Args:      []string{"text/html"},
											},
											{
												Directive: "proxy_pass",
												Line:      32,
												Args:      []string{"http://172.29.38.211/"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"lua-basic", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchLuaLatest},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("lua-basic", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "init_by_lua",
								Line:      2,
								Args:      []string{"\n        print(\"I need no extra escaping here, for example: \\r\\nblah\")\n    "},
								Block:     Directives{},
							},
							{
								Directive: "lua_shared_dict",
								Line:      5,
								Args:      []string{"dogs", "1m"},
								Block:     Directives{},
							},
							{
								Directive: "server",
								Line:      6,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      7,
										Args:      []string{"8080"},
										Block:     Directives{},
									},
									{
										Directive: "location",
										Line:      8,
										Args:      []string{"/"},
										Block: Directives{
											{
												Directive: "set_by_lua",
												Line:      9,
												Args:      []string{"$res", " return 32 + math.cos(32) "},
											},
											{
												Directive: "access_by_lua_file",
												Line:      10,
												Args:      []string{"/path/to/lua/access.lua"},
												Block:     Directives{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"lua-block-simple", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchLuaLatest},
		LexOptions: LexOptions{
			Lexers: []RegisterLexer{lua.RegisterLexer()},
		},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("lua-block-simple", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "init_by_lua_block",
								Line:      2,
								Args:      []string{"\n        print(\"Lua block code with curly brace str {\")\n    "},
								Block:     Directives{},
							},
							{
								Directive: "init_worker_by_lua_block",
								Line:      5,
								Args:      []string{"\n        print(\"Work that every worker\")\n    "},
								Block:     Directives{},
							},
							{
								Directive: "body_filter_by_lua_block",
								Line:      8,
								Args:      []string{"\n        local data, eof = ngx.arg[1], ngx.arg[2]\n    "},
								Block:     Directives{},
							},
							{
								Directive: "header_filter_by_lua_block",
								Line:      11,
								Args:      []string{"\n        ngx.header[\"content-length\"] = nil\n    "},
								Block:     Directives{},
							},
							{
								Directive: "server",
								Line:      14,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      15,
										Args:      []string{"127.0.0.1:8080"},
										Block:     Directives{},
									},
									{
										Directive: "location",
										Line:      16,
										Args:      []string{"/"},
										Block: Directives{
											{
												Directive: "content_by_lua_block",
												Line:      17,
												Args:      []string{"\n                ngx.say(\"I need no extra escaping here, for example: \\r\\nblah\")\n            "},
											},
											{
												Directive: "return",
												Args:      []string{"200", "foo bar baz"},
												Line:      20,
											},
										},
									},
									{
										Directive: "ssl_certificate_by_lua_block",
										Line:      22,
										Args:      []string{"\n            print(\"About to initiate a new SSL handshake!\")\n        "},
										Block:     Directives{},
									},
									{
										Directive: "log_by_lua_block",
										Line:      25,
										Args:      []string{"\n            print(\"I need no extra escaping here, for example: \\r\\nblah\")\n        "},
										Block:     Directives{},
									},
									{
										Directive: "location",
										Line:      28,
										Args:      []string{"/a"},
										Block: Directives{
											{
												Directive: "client_max_body_size",
												Line:      29,
												Args:      []string{"100k"},
											},
											{
												Directive: "client_body_buffer_size",
												Line:      30,
												Args:      []string{"100k"},
												Block:     Directives{},
											},
										},
									},
								},
							},
							{
								Directive: "upstream",
								Line:      34,
								Args:      []string{"foo"},
								Block: Directives{
									{
										Directive: "server",
										Line:      35,
										Args:      []string{"127.0.0.1"},
										Block:     Directives{},
									},
									{
										Directive: "balancer_by_lua_block",
										Line:      36,
										Args:      []string{"\n            -- use Lua to do something interesting here\n        "},
										Block:     Directives{},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"lua-block-cert-slim", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchLuaLatest},
		LexOptions: LexOptions{
			Lexers: []RegisterLexer{lua.RegisterLexer()},
		},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("lua-block-cert-slim", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "server",
								Line:      2,
								Args:      []string{},
								Block:     Directives{
									// TODO
								},
							},
						},
					},
				},
			},
		},
	}},
	{"lua-block-cert-double-server", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchLuaLatest},
		LexOptions: LexOptions{
			Lexers: []RegisterLexer{lua.RegisterLexer()},
		},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("lua-block-cert-double-server", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "server",
								Line:      2,
								Args:      []string{},
								Block:     Directives{
									// TODO
								},
							},
						},
					},
				},
			},
		},
	}},
	{"lua-block-larger", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchLuaLatest},
		LexOptions: LexOptions{
			Lexers: []RegisterLexer{lua.RegisterLexer()},
		},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("lua-block-larger", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "access_by_lua_block",
								Line:      2,
								Args: []string{"\n        -- check the client IP address is in our black list" +
									"\n        if ngx.var.remote_addr == \"132.5.72.3\" then" +
									"\n            ngx.exit(ngx.HTTP_FORBIDDEN)" +
									"\n        end" +
									"\n" +
									"\n        -- check if the URI contains bad words" +
									"\n        if ngx.var.uri and" +
									"\n               string.match(ngx.var.request_body, \"evil\")" +
									"\n        then" +
									"\n            return ngx.redirect(\"/terms_of_use.html\")" +
									"\n        end" +
									"\n" +
									"\n        -- tests passed" +
									"\n    "},
								Block: Directives{},
							},
							{
								Directive: "server",
								Line:      17,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      18,
										Args:      []string{"127.0.0.1:8080"},
										Block:     Directives{},
									},
									{
										Directive: "location",
										Line:      19,
										Args:      []string{"/"},
										Block: Directives{
											{
												Directive: "content_by_lua_block",
												Line:      20,
												Args: []string{"\n                ngx.req.read_body()  -- explicitly read the req body" +
													"\n                local data = ngx.req.get_body_data()" +
													"\n                if data then" +
													"\n                    ngx.say(\"body data:\")" +
													"\n                    ngx.print(data)" +
													"\n                    return" +
													"\n                end" +
													"\n" +
													"\n                -- body may get buffered in a temp file:" +
													"\n                local file = ngx.req.get_body_file()" +
													"\n                if file then" +
													"\n                    ngx.say(\"body is in file \", file)" +
													"\n                else" +
													"\n                    ngx.say(\"no body found\")" +
													"\n                end" +
													"\n            "},
												Block: Directives{},
											},
											{
												Directive: "return",
												Args:      []string{"200", "foo bar baz"},
												Line:      37,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"lua-block-tricky", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		ParseComments:            true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchLuaLatest},
		LexOptions: LexOptions{
			Lexers: []RegisterLexer{lua.RegisterLexer()},
		},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("lua-block-tricky", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "server",
								Line:      2,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      3,
										Args:      []string{"127.0.0.1:8080"},
										Block:     Directives{},
									},
									{
										Directive: "server_name",
										Line:      4,
										Args:      []string{"content_by_lua_block"},
										Block:     Directives{},
									},
									{
										Directive: "#",
										Args:      []string{},
										Line:      4,
										Comment:   pStr(" make sure this doesn't trip up lexers"),
									},
									{
										Directive: "set_by_lua_block",
										Line:      5,
										Args: []string{"$res", " -- irregular lua block directive" +
											"\n            local a = 32" +
											"\n            local b = 56" +
											"\n" +
											"\n            ngx.var.diff = a - b;  -- write to $diff directly" +
											"\n            return a + b;          -- return the $sum value normally" +
											"\n        "},
										Block: Directives{},
									},
									{
										Directive: "rewrite_by_lua_block",
										Line:      12,
										Args: []string{" -- have valid braces in Lua code and quotes around directive" +
											"\n            do_something(\"hello, world!\\nhiya\\n\")" +
											"\n            a = { 1, 2, 3 }" +
											"\n            btn = iup.button({title=\"ok\"})" +
											"\n        "},
										Block: Directives{},
									},
								},
							},
							{
								Directive: "upstream",
								Line:      18,
								Args:      []string{"content_by_lua_block"},
								Block: Directives{
									{
										Directive: "#",
										Args:      []string{},
										Line:      19,
										Comment:   pStr(" stuff"),
									},
								},
							},
						},
					},
				},
			},
		},
	}},
	{"limit-req-zone", "", ParseOptions{SingleFile: true}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("limit-req-zone", "nginx.conf"),
				Status: "ok",
				Errors: []ConfigError{},
				Parsed: Directives{
					{
						Directive: "user",
						Args:      []string{"nginx"},
						Line:      1,
						Block:     nil,
					},
					{
						Directive: "worker_processes",
						Args:      []string{"auto"},
						Line:      2,
						Block:     nil,
					},
					{
						Directive: "error_log",
						Args:      []string{"/var/log/nginx/error.log", "notice"},
						Line:      4,
						Block:     nil,
					},
					{
						Directive: "pid",
						Args:      []string{"/var/run/nginx.pid"},
						Line:      5,
						Block:     nil,
					},
					{
						Directive: "events",
						Args:      []string{},
						Line:      7,
						Block: Directives{
							{
								Directive: "worker_connections",
								Args:      []string{"1024"},
								Line:      8,
							},
						},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      11,
						Block: Directives{
							{
								Directive: "include",
								Args:      []string{"/etc/nginx/mime.types"},
								Line:      12,
							},
							{
								Directive: "default_type",
								Args:      []string{"application/octet-stream"},
								Line:      13,
							},
							{
								Directive: "limit_req_zone",
								Args:      []string{"$binary_remote_addr", "zone=one:10m", "rate=1r/s", "sync"},
								Line:      15,
							},
							{
								Directive: "log_format",
								Args: []string{
									"main",
									"$remote_addr - $remote_user [$time_local] \"$request\" ",
									"$status $body_bytes_sent \"$http_referer\" ",
									"\"$http_user_agent\" \"$http_x_forwarded_for\"",
								},
								Line: 17,
							},
							{
								Directive: "access_log",
								Args:      []string{"/var/log/nginx/access.log", "main"},
								Line:      21,
							},
							{
								Directive: "sendfile",
								Args:      []string{"on"},
								Line:      23,
							},
							{
								Directive: "keepalive_timeout",
								Args:      []string{"65"},
								Line:      25,
							},
						},
					},
				},
			},
		},
	}},
	{"geoip2", "", ParseOptions{
		SingleFile:               true,
		ErrorOnUnknownDirectives: true,
		DirectiveSources:         []MatchFunc{MatchNginxPlusLatest, MatchGeoip2Latest},
	}, Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{
			{
				File:   getTestConfigPath("geoip2", "nginx.conf"),
				Status: "ok",
				Parsed: Directives{
					{
						Directive: "http",
						Line:      1,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "geoip2",
								Line:      2,
								Args:      []string{"/etc/Geo/GeoLite2-City.mmdb"},
								Block: Directives{
									{
										Directive: "auto_reload",
										Args:      []string{"5s"},
										Line:      3,
										Block:     Directives{},
									},
									{
										Directive: "$geoip2_city_name",
										Args:      []string{"city", "names", "en"},
										Line:      4,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "geoip2_proxy",
								Line:      6,
								Args:      []string{"203.0.113.0/24"},
								Block:     Directives{},
							},
							{
								Directive: "geoip2_proxy_recursive",
								Line:      7,
								Args:      []string{"on"},
								Block:     Directives{},
							},
							{
								Directive: "server",
								Line:      8,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      9,
										Args:      []string{"80"},
										Block:     Directives{},
									},
									{
										Directive: "server_name",
										Line:      10,
										Args:      []string{"localhost"},
										Block:     Directives{},
									},
									{
										Directive: "location",
										Line:      11,
										Args:      []string{"/"},
										Block: Directives{
											{
												Directive: "return",
												Line:      12,
												Args: []string{
													"200",
													"Hello $geoip2_city_name",
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Directive: "stream",
						Line:      18,
						Args:      []string{},
						Block: Directives{
							{
								Directive: "geoip2",
								Line:      19,
								Args:      []string{"/etc/Geo/GeoLite2-Country.mmdb"},
								Block: Directives{
									{
										Directive: "$geoip2_country_name",
										Args:      []string{"country", "names", "en"},
										Line:      20,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "map",
								Line:      23,
								Args:      []string{"$geoip2_country_name", "$backend"},
								Block: Directives{
									{
										Directive: "United States",
										Args:      []string{"us_backend"},
										Line:      24,
										Block:     Directives{},
									},
									{
										Directive: "default",
										Args:      []string{"default_backend"},
										Line:      25,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "server",
								Line:      28,
								Args:      []string{},
								Block: Directives{
									{
										Directive: "listen",
										Line:      29,
										Args:      []string{"12345"},
										Block:     Directives{},
									},
									{
										Directive: "proxy_pass",
										Args:      []string{"$backend"},
										Line:      30,
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "upstream",
								Line:      33,
								Args:      []string{"us_backend"},
								Block: Directives{
									{
										Directive: "server",
										Line:      34,
										Args:      []string{"192.168.0.1:12345"},
										Block:     Directives{},
									},
								},
							},
							{
								Directive: "upstream",
								Line:      37,
								Args:      []string{"default_backend"},
								Block: Directives{
									{
										Directive: "server",
										Line:      39,
										Args:      []string{"192.168.0.2:12345"},
										Block:     Directives{},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
}

func TestParse(t *testing.T) {
	t.Parallel()
	for _, fixture := range parseFixtures {
		fixture := fixture
		t.Run(fixture.name+fixture.suffix, func(t *testing.T) {
			t.Parallel()
			path := getTestConfigPath(fixture.name, "nginx.conf")
			payload, err := Parse(path, &fixture.options)
			if err != nil {
				t.Fatal(err)
			}
			if !equalPayloads(t, *payload, fixture.expected) {
				b1, _ := json.Marshal(fixture.expected)
				b2, _ := json.Marshal(payload)
				t.Fatalf("expected: %s\nbut got: %s", b1, b2)
			}
		})
	}
}

//nolint:errchkjson
func TestParseVarArgs(t *testing.T) {
	t.Parallel()
	tcs := map[string]struct {
		fn string
	}{
		"simple resolver":                        {fn: "simple"},
		"multiple resolver addresses":            {fn: "multiple_resolvers"},
		"multiple resolver addresses with ports": {fn: "multiple_resolvers_with_ports"},
	}

	for n, tc := range tcs {
		t.Log(n)
		path := getTestConfigPath("upstream_resolver_directive", tc.fn+".conf")
		golden := getTestConfigPath("upstream_resolver_directive", tc.fn+".conf.golden")

		payload, err := Parse(path, &ParseOptions{SingleFile: true, StopParsingOnError: true})
		require.NoError(t, err, "parsing error when reading test file")
		require.Len(t, payload.Config, 1)

		gpayload, err := Parse(golden, &ParseOptions{SingleFile: true, StopParsingOnError: true})
		require.NoError(t, err, "parsing error when reading golden file")
		require.Len(t, gpayload.Config, 1)

		b1, _ := json.Marshal(payload.Config[0].Parsed)
		b2, _ := json.Marshal(gpayload.Config[0].Parsed)
		require.Equal(t, string(b1), string(b2))
	}
}

func TestParseIfExpr(t *testing.T) {
	t.Parallel()
	tcs := map[string]struct {
		fn  string
		err bool
	}{
		"valid if expr":         {fn: "nginx"},
		"spaced parens":         {fn: "spaced-parens"},
		"missing opening paren": {fn: "missing-opening-paren", err: true},
		"missing closing paren": {fn: "missing-closing-paren", err: true},
		"no parens":             {fn: "no-parens", err: true},
		"empty parens":          {fn: "empty-parens", err: true},
		"empty spaced parens":   {fn: "empty-spaced-parens", err: true},
	}

	for n, tc := range tcs {
		t.Log(n)
		path := getTestConfigPath("if-expr", tc.fn+".conf")

		payload, err := Parse(path, &ParseOptions{SingleFile: true, StopParsingOnError: true})
		if tc.err {
			require.Error(t, err, "expected parsing error when reading test file: %s", path)
		} else {
			require.NoError(t, err, "unexpected parsing error when reading test file: %s", path)
			require.Len(t, payload.Config, 1)
		}
	}
}

func TestBalancingBraces(t *testing.T) {
	t.Parallel()
	tcs := map[string]struct {
		fn  string
		err string
	}{
		"extra brace":   {fn: "extra-brace", err: `unexpected "}"`},
		"missing brace": {fn: "missing-brace", err: `unexpected end of file, expecting "}"`},
	}

	for n, tc := range tcs {
		t.Log(n)
		path := getTestConfigPath("braces", tc.fn+".conf")

		payload, err := Parse(path, &ParseOptions{SingleFile: true, StopParsingOnError: true})
		require.Nil(t, payload, "expected nil payload when reading bad test file: %s", path)
		require.Error(t, err, "expected parsing error when reading test file: %s", path)
		require.Contains(t, err.Error(), tc.err+" in "+path)
	}
}

func TestCombinedIncludes(t *testing.T) {
	t.Parallel()
	tcs := map[string]struct {
		fn  string
		err bool
	}{
		"nested includes": {fn: "valid"},
		"include cycle":   {fn: "invalid", err: true},
	}

	for n, tc := range tcs {
		t.Log(n)
		path := getTestConfigPath("includes-cycle", tc.fn, "nginx.conf")

		payload, err := Parse(path, &ParseOptions{CombineConfigs: true, StopParsingOnError: true})
		if tc.err {
			require.Nil(t, payload, "expected nil payload when reading bad test file: %s", path)
			require.Error(t, err, "expected parsing error when reading test file: %s", path)
		} else {
			require.NoError(t, err, "unexpected parsing error when reading test file: %s", path)
			require.Len(t, payload.Config, 1)
		}
	}
}

func TestDefaultUbuntu(t *testing.T) {
	t.Parallel()
	path := getTestConfigPath("ubuntu-default", "nginx.conf")
	_, err := Parse(path, &ParseOptions{SingleFile: false, StopParsingOnError: true})
	require.NoError(t, err, "unexpected parsing error when reading test file: %s", path)
}
