package cmd

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/nginxinc/crossplane-go/pkg/builder"
	"github.com/nginxinc/crossplane-go/pkg/parser"
)

func TestParseAndBuild(t *testing.T) {
	var tests = []struct {
		name     string
		args     parser.ParseArgs
		expected parser.Payload
	}{
		{
			"bad-args/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/bad-args/nginx.conf",
				Status: "failed",
				Errors: []parser.ParseError{
					{
						File:  "configs/bad-args/nginx.conf",
						Line:  1,
						Error: errors.New("invalid number of arguements in user"),
					},
				},
				Config: []parser.Config{
					{
						File:   "configs/bad-args/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Args:      []string{},
								Line:      2,
								Comment:   "",
								File:      "",
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Args:      []string{},
								Line:      3,
								Comment:   "",
								Block:     []parser.Block{},
								File:      "",
							},
						},
					},
				},
			},
		},
		{
			"directive-with-space/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/directive-with-space/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/directive-with-space/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Args:      []string{},
								Comment:   "",
								File:      "",
								Line:      1,
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Args:      []string{},
								Comment:   "",
								Line:      3,
								File:      "",
								Block: []parser.Block{
									{
										Directive: "map",
										Args:      []string{"$http_user_agent", "$mobile"},
										Line:      4,
										File:      "",
										Comment:   "",
										Block: []parser.Block{
											{
												Directive: "default",
												Args:      []string{"0"},
												Line:      5,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											}, {
												Directive: "'~Opera Mini'",
												Args:      []string{"1"},
												Line:      6,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
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
		{
			"empty-value-map/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/empty-value-map/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/empty-value-map/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Args:      []string{},
								Line:      1,
								Comment:   "",
								File:      "",
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Line:      3,
								Args:      []string{},
								Comment:   "",
								File:      "",
								Block: []parser.Block{
									{
										Directive: "map",
										Args:      []string{"string", "$variable"},
										Line:      4,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "''",
												Args:      []string{"$arg"},
												Comment:   "",
												Line:      5,
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "*.example.com",
												Args:      []string{"''"},
												Line:      6,
												File:      "",
												Comment:   "",
												Block:     []parser.Block{},
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

		{
			"includes-globbed/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     true,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/includes-globbed/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/includes-globbed/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Args:      []string{},
								Comment:   "",
								File:      "configs/includes-globbed/nginx.conf",
								Line:      1,
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Args:      []string{},
								Line:      1,
								Comment:   "",
								File:      "configs/includes-globbed/http.conf",
								Block: []parser.Block{
									{
										Directive: "server",
										Args:      []string{},
										Line:      1,
										Comment:   "",
										File:      "configs/includes-globbed/servers/server1.conf",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"8080"},
												Line:      2,
												Comment:   "",
												File:      "configs/includes-globbed/servers/server1.conf",
												Block:     []parser.Block{},
											}, {
												Directive: "location",
												Args:      []string{"/foo"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-globbed/locations/location1.conf",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "'foo'"},
														Comment:   "",
														Line:      2,
														File:      "configs/includes-globbed/locations/location1.conf",
														Block:     []parser.Block{},
													},
												},
											},
											{
												Directive: "location",
												Args:      []string{"/bar"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-globbed/locations/location2.conf",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "'bar'"},
														Comment:   "",
														Line:      2,
														File:      "configs/includes-globbed/locations/location2.conf",
														Block:     []parser.Block{},
													},
												},
											},
										},
									},
									{
										Directive: "server",
										Args:      []string{},
										Line:      1,
										Comment:   "",
										File:      "configs/includes-globbed/servers/server2.conf",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"8081"},
												Line:      2,
												Comment:   "",
												File:      "configs/includes-globbed/servers/server2.conf",
												Block:     []parser.Block{},
											}, {
												Directive: "location",
												Args:      []string{"/foo"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-globbed/locations/location1.conf",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "'foo'"},
														Comment:   "",
														Line:      2,
														File:      "configs/includes-globbed/locations/location1.conf",
														Block:     []parser.Block{},
													},
												},
											}, {
												Directive: "location",
												Args:      []string{"/bar"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-globbed/locations/location2.conf",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "'bar'"},
														Comment:   "",
														Line:      2,
														File:      "configs/includes-globbed/locations/location2.conf",
														Block:     []parser.Block{},
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

		{
			"includes-regular/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     true,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/includes-regular/nginx.conf",
				Status: "failed",
				Errors: []parser.ParseError{
					{
						File:  "configs/includes-regular/conf.d/server.conf",
						Line:  5,
						Error: errors.New("open configs/includes-regular/bar.conf: no such file or directory"),
					},
				},
				Config: []parser.Config{
					{
						File:   "configs/includes-regular/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Line:      1,
								Args:      []string{},
								Comment:   "",
								File:      "configs/includes-regular/nginx.conf",
								Block:     []parser.Block{},
							},
							{
								Directive: "http",
								Args:      []string{},
								Line:      2,
								Comment:   "",
								File:      "configs/includes-regular/nginx.conf",
								Block: []parser.Block{
									{
										Directive: "server",
										Line:      1,
										Comment:   "",
										Args:      []string{},
										File:      "configs/includes-regular/conf.d/server.conf",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      2,
												Comment:   "",
												File:      "configs/includes-regular/conf.d/server.conf",
												Block:     []parser.Block{},
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Line:      3,
												Comment:   "",
												File:      "configs/includes-regular/conf.d/server.conf",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"/foo"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-regular/foo.conf",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "'foo'"},
														Line:      2,
														Comment:   "",
														File:      "configs/includes-regular/foo.conf",
														Block:     []parser.Block{},
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

		{

			"messy/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    true,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/messy/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/messy/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "user",
								Args:      []string{"nobody"},
								Line:      1,
								File:      "",
								Comment:   "",
								Block:     []parser.Block{},
							},
							{
								Directive: "#",
								Args:      []string{},
								Line:      2,
								File:      "",
								Comment:   ` hello\n\\n\\\n worlddd  \#\\#\\\# dfsf\n \\n \\\n \`,
								Block:     []parser.Block{},
							},
							{
								Directive: "\"events\"",
								Args:      []string{},
								Line:      3,
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "\"worker_connections\"",
										Args:      []string{"\"2048\""},
										Line:      3,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
								},
							},

							{
								Directive: "\"http\"",
								Args:      []string{},
								Line:      5,
								Comment:   "",
								File:      "",
								Block: []parser.Block{
									{
										Directive: "#",
										Args:      []string{},
										Line:      5,
										Comment:   "forteen",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "#",
										Args:      []string{},
										Line:      6,
										Comment:   " this is a comment",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "\"access_log\"",
										Args:      []string{"off"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "default_type",
										Args:      []string{"\"text/plain\""},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "error_log",
										Args:      []string{"\"off\""},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "server",
										Args:      []string{},
										Line:      8,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "\"listen\"",
												Args:      []string{"\"8083\""},
												Line:      9,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "\"return\"",
												Args:      []string{"200", `"Ser" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n"`},
												Line:      10,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
										},
									},
									{
										Directive: "\"server\"",
										Args:      []string{},
										Line:      12,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "\"listen\"",
												Args:      []string{"8080"},
												Comment:   "",
												Line:      12,
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "'root'",
												Args:      []string{"/usr/share/nginx/html"},
												Line:      13,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"~", "\"/hello/world;\""},
												Comment:   "",
												Line:      14,
												File:      "",
												Block: []parser.Block{
													{
														Directive: "\"return\"",
														Args:      []string{"301", "/status.html"},
														Line:      14,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},
											{
												Directive: "location",
												Args:      []string{"/foo"},
												Line:      15,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"/bar"},
												Line:      15,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"/\\{\\;\\}\\ #\\ ab"},
												Line:      16,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "#",
												Args:      []string{},
												Line:      16,
												Comment:   " hello",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "if",
												Args:      []string{"$request_method", "=", "P\\{O\\)\\###\\;ST"},
												Line:      17,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"\"/status.html\""},
												Line:      18,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "try_files",
														Args:      []string{"/abc/${uri}", "/abc/${uri}.html", "=404"},
														Line:      19,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},

											{
												Directive: "\"location\"",
												Args:      []string{"\"/sta;\n                    tus\""},
												Line:      21,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "\"return\"",
														Args:      []string{"302", "/status.html"},
														Line:      22,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},

											{
												Directive: "\"location\"",
												Args:      []string{"/upstream_conf"},
												Line:      23,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "\"return\"",
														Args:      []string{"200", "/status.html"},
														Line:      23,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},
										},
									},
									{
										Directive: "server",
										Args:      []string{},
										Line:      24,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			"missing-semicolon/broken-above.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/missing-semicolon/",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/missing-semicolon/broken-above.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "http",
								Line:      1,
								Comment:   "",
								Args:      []string{},
								File:      "",
								Block: []parser.Block{
									{
										Directive: "server",
										Line:      2,
										Comment:   "",
										Args:      []string{},
										File:      "",
										Block: []parser.Block{
											{
												Directive: "location",
												Line:      3,
												Comment:   "",
												Args:      []string{"/is-broken"},
												File:      "",
												Block: []parser.Block{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://is.broken.example"},
														Line:      4,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},
											{
												Directive: "location",
												Line:      6,
												Args:      []string{"/not-broken"},
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://not.broken.example"},
														Line:      7,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{

						File:   "configs/missing-semicolon/broken-above.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "http",
								Line:      1,
								Comment:   "",
								Args:      []string{},
								File:      "",
								Block: []parser.Block{
									{
										Directive: "server",
										Line:      2,
										Comment:   "",
										Args:      []string{},
										File:      "",
										Block: []parser.Block{
											{
												Directive: "location",
												Line:      3,
												Comment:   "",
												Args:      []string{"/not-broken"},
												File:      "",
												Block: []parser.Block{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://not.broken.example"},
														Line:      4,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},
											{
												Directive: "location",
												Line:      6,
												Args:      []string{"/is-broken"},
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://is.broken.example"},
														Line:      7,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
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

		{
			"quote-behavior/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{},
		},

		{
			"russian-text/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/russian-text/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/russian-text/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "env",
								Line:      1,
								Args:      []string{"'русский текст'"},
								Comment:   "",
								File:      "",
								Block:     []parser.Block{},
							},
							{
								Directive: "events",
								Line:      2,
								Args:      []string{},
								Comment:   "",
								File:      "",
								Block:     []parser.Block{},
							},
						},
					},
				},
			},
		},

		{
			"simple/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/simple/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/simple/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Line:      1,
								Comment:   "",
								Args:      []string{},
								File:      "",
								Block: []parser.Block{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
								},
							},
							{
								Directive: "http",
								Line:      5,
								Comment:   "",
								File:      "",
								Args:      []string{},
								Block: []parser.Block{
									{
										Directive: "server",
										Line:      6,
										Comment:   "",
										File:      "",
										Args:      []string{},
										Block: []parser.Block{
											{
												Directive: "listen",
												Line:      7,
												Comment:   "",
												Args:      []string{"127.0.0.1:8080"},
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "server_name",
												Line:      8,
												Comment:   "",
												File:      "",
												Args:      []string{"default_server"},
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Line:      9,
												Comment:   "",
												Args:      []string{"/"},
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Line:      10,
														Comment:   "",
														File:      "",
														Args:      []string{"200", "\"foo bar baz\""},
														Block:     []parser.Block{},
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
		{
			"spelling-mistake/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      true,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/spelling-mistake/nginx.conf",
				Status: "failed",
				Errors: []parser.ParseError{
					{
						File:  "configs/spelling-mistake/nginx.conf",
						Line:  7,
						Error: errors.New("unknown directive proxy_passs"),
					},
				},
				Config: []parser.Config{
					{
						File:   "configs/spelling-mistake/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Line:      1,
								Comment:   "",
								Args:      []string{},
								File:      "",
							},
							{
								Directive: "http",
								Line:      3,
								Comment:   "",
								File:      "",
								Args:      []string{},
								Block: []parser.Block{
									{
										Directive: "server",
										Args:      []string{},
										Comment:   "",
										File:      "",
										Line:      4,
										Block: []parser.Block{
											{
												Directive: "location",
												Args:      []string{"/"},
												Line:      5,
												File:      "",
												Comment:   "",
												Block:     []parser.Block{},
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
		{
			"with-comments/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    true,
				Strict:      false,
				Combine:     false,
				CheckCtx:    true,
				CheckArgs:   true,
			},
			parser.Payload{
				File:   "configs/with-comments/nginx.conf",
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/with-comments/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Args:      []string{},
								Line:      1,
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
								},
							},
							{
								Directive: "#",
								Args:      []string{},
								Comment:   "comment",
								File:      "",
								Line:      4,
								Block:     []parser.Block{},
							},
							{
								Directive: "http",
								Line:      5,
								Args:      []string{},
								Comment:   "",
								File:      "",
								Block: []parser.Block{
									{
										Directive: "server",
										Args:      []string{},
										Comment:   "",
										Line:      6,
										File:      "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Comment:   "",
												File:      "",
												Line:      7,
												Block:     []parser.Block{},
											},
											{
												Directive: "#",
												Args:      []string{},
												Comment:   "listen",
												File:      "",
												Line:      7,
												Block:     []parser.Block{},
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Comment:   "",
												File:      "",
												Line:      8,
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Line:      9,
												Args:      []string{"/"},
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "#",
														Line:      9,
														Args:      []string{},
														File:      "",
														Comment:   "# this is brace",
														Block:     []parser.Block{},
													},
													{
														Directive: "#",
														Args:      []string{},
														File:      "",
														Comment:   " location /",
														Line:      10,
														Block:     []parser.Block{},
													},
													{
														Directive: "return",
														Line:      11,
														Comment:   "",
														Args:      []string{"200", "\"foo bar baz\""},
														File:      "",
														Block:     []parser.Block{},
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
	}

	for _, test := range tests {
		test.args.FileName = "configs/" + test.name
		f := test.args.FileName
		i := test.args.Ignore
		catch := test.args.CatchErrors
		sin := test.args.Single
		com := test.args.Comments
		strict := test.args.Strict
		comb := test.args.Combine
		ctx := test.args.CheckCtx
		check := test.args.CheckArgs
		con := test.args.Consume
		parsed, err := parser.Parse(f, catch, i, sin, com, strict, comb, con, ctx, check)
		if err != nil {
			t.Errorf(err.Error())
		}

		_, err = builder.BuildFiles(parsed, "test1", 4, false, false)
		if err != nil {
			t.Errorf(err.Error())
		}
		f1 := "test1/" + test.args.FileName
		f2 := "test2/test1/" + test.args.FileName
		f = "test1/" + test.args.FileName
		parsed1, err := parser.Parse(f, catch, i, sin, com, strict, comb, con, ctx, check)
		_, err = builder.BuildFiles(parsed1, "test2", 4, false, false)
		if err != nil {
			t.Errorf(err.Error())
		}
		result, result2 := compareFiles(f1, f2)
		if len(result) != 0 && len(result) != 0 {
			t.Errorf("\n%v\n\n%v ", string(result), string(result2))
		}
	}
}

func compareFiles(inputfile, outputfile string) ([]byte, []byte) {
	f1, err := os.Open(inputfile)
	if err != nil {
		return []byte{}, []byte{}
	}

	f2, err := os.Open(outputfile)
	if err != nil {
		return []byte{'\''}, []byte{}
	}

	for {
		b1 := make([]byte, 64000)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, 64000)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return []byte{}, []byte{}
			} else if err1 == io.EOF || err2 == io.EOF {
				return b2, b1
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return b1, b2
		}
	}
}

func compareConfigs(conf parser.Config, c parser.Config) string {
	var s string
	if !reflect.DeepEqual(conf.File, c.File) {
		s += "Problems with the names of config files" + string('\n')
	}
	if len(conf.Errors) != len(c.Errors) {
		s += "Errors are not the same length" + string('\n')
	}
	if !reflect.DeepEqual(conf.Status, c.Status) {
		s += "the Status's are not the same" + string('\n')
	}

	for i := 0; i < len(c.Parsed); i++ {
		s += compareBlocks(conf.Parsed[i], c.Parsed[i])
	}
	return s
}

func compareBlocks(gen parser.Block, config parser.Block) string {
	var s string
	if !reflect.DeepEqual(gen.Directive, config.Directive) {
		s += "Error with directives : " + gen.Directive + " && " + config.Directive + string('\n')
	}
	if len(gen.Args) == len(config.Args) {
		for i := 0; i < len(gen.Args); i++ {
			if !reflect.DeepEqual(gen.Args[i], config.Args[i]) {
				s += "Problem with Args in Block " + gen.Directive + " && " + config.Directive + string('\n')
			}
		}
	} else {
		s += "Problem with Args in Block " + gen.Directive + " && " + config.Directive + string('\n')
	}
	if !reflect.DeepEqual(gen.Line, config.Line) {
		s += "Problem with Line in Block " + gen.Directive + " && " + config.Directive + string('\n')
	}
	if !reflect.DeepEqual(gen.File, config.File) {
		s += "Problem with File in Block " + gen.Directive + " && " + config.Directive + string('\n')
	}
	if !reflect.DeepEqual(gen.Comment, config.Comment) {
		s += "Problem with Comments in Block " + gen.Comment + " && " + config.Comment + string('\n')
	}
	for i := 0; i < len(config.Block); i++ {
		s += compareBlocks(gen.Block[i], config.Block[i])
	}
	return s
}

func TestExecute(t *testing.T) {

}
