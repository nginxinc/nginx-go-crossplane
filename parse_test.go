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

// nolint:gochecknoglobals
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
						Block:     Directives{},
					},
					{
						Directive: "http",
						Args:      []string{},
						Line:      1,
						Block: Directives{
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
			if !equalPayloads(*payload, fixture.expected) {
				b1, _ := json.Marshal(fixture.expected)
				b2, _ := json.Marshal(payload)
				t.Fatalf("expected: %s\nbut got: %s", b1, b2)
			}
		})
	}
}

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
		require.Contains(t, err.Error(), tc.err+" in lexer")
	}
}
