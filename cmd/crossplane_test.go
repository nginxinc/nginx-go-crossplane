// build skip

package cmd

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"testing"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/builder"
	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
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
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/bad-args/nginx.conf",
				Errors: []parser.ParseError{
					{
						File: "configs/bad-args/nginx.conf",
						Line: 1,
						Fail: errors.New("invalid number of arguements in user"),
					},
				},
				Config: []*parser.Config{
					{
						File:   "configs/bad-args/nginx.conf",
						Errors: []parser.ParseError{},
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      2,
							}, {
								Directive: "http",
								Line:      3,
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
			},
			parser.Payload{
				File: "configs/directive-with-space/nginx.conf",
				Config: []*parser.Config{
					{
						File: "configs/directive-with-space/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
							}, {
								Directive: "http",
								Line:      3,
								Block: []*parser.Directive{
									{
										Directive: "map",
										Args:      []string{"$http_user_agent", "$mobile"},
										Line:      4,
										Block: []*parser.Directive{
											{
												Directive: "default",
												Args:      []string{"0"},
												Line:      5,
											}, {
												Directive: "'~Opera Mini'",
												Args:      []string{"1"},
												Line:      6,
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
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/empty-value-map/nginx.conf",
				Config: []*parser.Config{
					{
						File: "configs/empty-value-map/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
							}, {
								Directive: "http",
								Line:      3,
								Block: []*parser.Directive{
									{
										Directive: "map",
										Args:      []string{"string", "$variable"},
										Line:      4,
										Block: []*parser.Directive{
											{
												Directive: "''",
												Args:      []string{"$arg"},
												Line:      5,
												Block:     []*parser.Directive{},
											},
											{
												Directive: "*.example.com",
												Args:      []string{"''"},
												Line:      6,
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
				CatchErrors: true,
			},
			parser.Payload{
				File:   "configs/includes-globbed/nginx.conf",
				Errors: []parser.ParseError{},
				Config: []*parser.Config{
					{
						File:   "configs/includes-globbed/nginx.conf",
						Errors: []parser.ParseError{},
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								File:      "configs/includes-globbed/nginx.conf",
								Line:      1,
								Block:     []*parser.Directive{},
							}, {
								Directive: "http",
								Line:      1,
								File:      "configs/includes-globbed/http.conf",
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      1,
										File:      "configs/includes-globbed/servers/server1.conf",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"8080"},
												Line:      2,
												Comment:   "",
												File:      "configs/includes-globbed/servers/server1.conf",
												Block:     []*parser.Directive{},
											}, {
												Directive: "location",
												Args:      []string{"/foo"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-globbed/locations/location1.conf",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "'foo'"},
														Comment:   "",
														Line:      2,
														File:      "configs/includes-globbed/locations/location1.conf",
													},
												},
											},
											{
												Directive: "location",
												Args:      []string{"/bar"},
												Line:      1,
												File:      "configs/includes-globbed/locations/location2.conf",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "'bar'"},
														Line:      2,
														File:      "configs/includes-globbed/locations/location2.conf",
													},
												},
											},
										},
									},
									{
										Directive: "server",
										Line:      1,
										File:      "configs/includes-globbed/servers/server2.conf",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"8081"},
												Line:      2,
												File:      "configs/includes-globbed/servers/server2.conf",
												Block:     []*parser.Directive{},
											}, {
												Directive: "location",
												Args:      []string{"/foo"},
												Line:      1,
												File:      "configs/includes-globbed/locations/location1.conf",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "'foo'"},
														Line:      2,
														File:      "configs/includes-globbed/locations/location1.conf",
													},
												},
											}, {
												Directive: "location",
												Args:      []string{"/bar"},
												Comment:   "",
												Line:      1,
												File:      "configs/includes-globbed/locations/location2.conf",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "'bar'"},
														Comment:   "",
														Line:      2,
														File:      "configs/includes-globbed/locations/location2.conf",
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
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/includes-regular/nginx.conf",
				Errors: []parser.ParseError{
					{
						File: "configs/includes-regular/conf.d/server.conf",
						Line: 5,
						Fail: errors.New("open configs/includes-regular/bar.conf: no such file or directory"),
					},
				},
				Config: []*parser.Config{
					{
						File:   "configs/includes-regular/nginx.conf",
						Errors: []parser.ParseError{},
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
								File:      "configs/includes-regular/nginx.conf",
								Block:     []*parser.Directive{},
							},
							{
								Directive: "http",
								Line:      2,
								File:      "configs/includes-regular/nginx.conf",
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      1,
										File:      "configs/includes-regular/conf.d/server.conf",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      2,
												File:      "configs/includes-regular/conf.d/server.conf",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Line:      3,
												File:      "configs/includes-regular/conf.d/server.conf",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "location",
												Args:      []string{"/foo"},
												Line:      1,
												File:      "configs/includes-regular/foo.conf",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "'foo'"},
														Line:      2,
														File:      "configs/includes-regular/foo.conf",
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
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/messy/nginx.conf",
				Config: []*parser.Config{
					{
						File: "configs/messy/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "user",
								Args:      []string{"nobody"},
								Line:      1,
							},
							{
								Directive: "#",
								Line:      2,
								Comment:   ` hello\n\\n\\\n worlddd  \#\\#\\\# dfsf\n \\n \\\n \`,
							},
							{
								Directive: "\"events\"",
								Line:      3,
								Block: []*parser.Directive{
									{
										Directive: "\"worker_connections\"",
										Args:      []string{"\"2048\""},
										Line:      3,
									},
								},
							},

							{
								Directive: "\"http\"",
								Line:      5,
								Block: []*parser.Directive{
									{
										Directive: "#",
										Line:      5,
										Comment:   "forteen",
									},
									{
										Directive: "#",
										Line:      6,
										Comment:   " this is a comment",
									},
									{
										Directive: "\"access_log\"",
										Args:      []string{"off"},
										Line:      7,
									},
									{
										Directive: "default_type",
										Args:      []string{"\"text/plain\""},
										Line:      7,
										Comment:   "",
										File:      "",
									},
									{
										Directive: "error_log",
										Args:      []string{"\"off\""},
										Line:      7,
									},
									{
										Directive: "server",
										Line:      8,
										Block: []*parser.Directive{
											{
												Directive: "\"listen\"",
												Args:      []string{"\"8083\""},
												Line:      9,
											},
											{
												Directive: "\"return\"",
												Args:      []string{"200", `"Ser" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n"`},
												Line:      10,
											},
										},
									},
									{
										Directive: "\"server\"",
										Line:      12,
										Block: []*parser.Directive{
											{
												Directive: "\"listen\"",
												Args:      []string{"8080"},
												Line:      12,
											},
											{
												Directive: "'root'",
												Args:      []string{"/usr/share/nginx/html"},
												Line:      13,
											},
											{
												Directive: "location",
												Args:      []string{"~", "\"/hello/world;\""},
												Line:      14,
												Block: []*parser.Directive{
													{
														Directive: "\"return\"",
														Args:      []string{"301", "/status.html"},
														Line:      14,
													},
												},
											},
											{
												Directive: "location",
												Args:      []string{"/foo"},
												Line:      15,
											},
											{
												Directive: "location",
												Args:      []string{"/bar"},
												Line:      15,
											},
											{
												Directive: "location",
												Args:      []string{"/\\{\\;\\}\\ #\\ ab"},
												Line:      16,
											},
											{
												Directive: "#",
												Line:      16,
												Comment:   " hello",
											},
											{
												Directive: "if",
												Args:      []string{"$request_method", "=", "P\\{O\\)\\###\\;ST"},
												Line:      17,
											},
											{
												Directive: "location",
												Args:      []string{"\"/status.html\""},
												Line:      18,
												Block: []*parser.Directive{
													{
														Directive: "try_files",
														Args:      []string{"/abc/${uri}", "/abc/${uri}.html", "=404"},
														Line:      19,
													},
												},
											},

											{
												Directive: "\"location\"",
												Args:      []string{"\"/sta;\n                    tus\""},
												Line:      21,
												Block: []*parser.Directive{
													{
														Directive: "\"return\"",
														Args:      []string{"302", "/status.html"},
														Line:      22,
													},
												},
											},

											{
												Directive: "\"location\"",
												Args:      []string{"/upstream_conf"},
												Line:      23,
												Block: []*parser.Directive{
													{
														Directive: "\"return\"",
														Args:      []string{"200", "/status.html"},
														Line:      23,
													},
												},
											},
										},
									},
									{
										Directive: "server",
										Line:      24,
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
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/missing-semicolon/",
				Config: []*parser.Config{
					{
						File: "configs/missing-semicolon/broken-above.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "http",
								Line:      1,
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      2,
										Block: []*parser.Directive{
											{
												Directive: "location",
												Line:      3,
												Args:      []string{"/is-broken"},
												Block: []*parser.Directive{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://is.broken.example"},
														Line:      4,
													},
												},
											},
											{
												Directive: "location",
												Line:      6,
												Args:      []string{"/not-broken"},
												Block: []*parser.Directive{
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
					{

						File:   "configs/missing-semicolon/broken-above.conf",
						Errors: []parser.ParseError{},
						Parsed: []*parser.Directive{
							{
								Directive: "http",
								Line:      1,
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      2,
										Block: []*parser.Directive{
											{
												Directive: "location",
												Line:      3,
												Args:      []string{"/not-broken"},
												Block: []*parser.Directive{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://not.broken.example"},
														Line:      4,
													},
												},
											},
											{
												Directive: "location",
												Line:      6,
												Args:      []string{"/is-broken"},
												Block: []*parser.Directive{
													{
														Directive: "proxy_pass",
														Args:      []string{"http://is.broken.example"},
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
			},
		},

		{
			"quote-behavior/nginx.conf",
			parser.ParseArgs{
				CatchErrors: true,
			},
			parser.Payload{},
		},

		{
			"russian-text/nginx.conf",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/russian-text/nginx.conf",
				Config: []*parser.Config{
					{
						File: "configs/russian-text/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "env",
								Line:      1,
								Args:      []string{"'русский текст'"},
							},
							{
								Directive: "events",
								Line:      2,
							},
						},
					},
				},
			},
		},

		{
			"simple/nginx.conf",
			parser.ParseArgs{
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/simple/nginx.conf",
				Config: []*parser.Config{
					{
						File: "configs/simple/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
								Block: []*parser.Directive{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
									},
								},
							},
							{
								Directive: "http",
								Line:      5,
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      6,
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Line:      7,
												Args:      []string{"127.0.0.1:8080"},
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
												Block: []*parser.Directive{
													{
														Directive: "return",
														Line:      10,
														Args:      []string{"200", "\"foo bar baz\""},
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
				CatchErrors: true,
			},
			parser.Payload{
				File: "configs/spelling-mistake/nginx.conf",
				Errors: []parser.ParseError{
					{
						File: "configs/spelling-mistake/nginx.conf",
						Line: 7,
						Fail: errors.New("unknown directive proxy_passs"),
					},
				},
				Config: []*parser.Config{
					{
						File: "configs/spelling-mistake/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
							},
							{
								Directive: "http",
								Line:      3,
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      4,
										Block: []*parser.Directive{
											{
												Directive: "location",
												Args:      []string{"/"},
												Line:      5,
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
			},
			parser.Payload{
				File: "configs/with-comments/nginx.conf",
				Config: []*parser.Config{
					{
						File:   "configs/with-comments/nginx.conf",
						Errors: []parser.ParseError{},
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
								Block: []*parser.Directive{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
									},
								},
							},
							{
								Directive: "#",
								Comment:   "comment",
								Line:      4,
							},
							{
								Directive: "http",
								Line:      5,
								Block: []*parser.Directive{
									{
										Directive: "server",
										Comment:   "",
										Line:      6,
										File:      "",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      7,
												Block:     []*parser.Directive{},
											},
											{
												Directive: "#",
												Args:      []string{},
												Comment:   "listen",
												Line:      7,
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Line:      8,
											},
											{
												Directive: "location",
												Line:      9,
												Args:      []string{"/"},
												Block: []*parser.Directive{
													{
														Directive: "#",
														Line:      9,
														Args:      []string{},
														File:      "",
														Comment:   "# this is brace",
														Block:     []*parser.Directive{},
													},
													{
														Directive: "#",
														Args:      []string{},
														File:      "",
														Comment:   " location /",
														Line:      10,
														Block:     []*parser.Directive{},
													},
													{
														Directive: "return",
														Line:      11,
														Args:      []string{"200", "\"foo bar baz\""},
														Block:     []*parser.Directive{},
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
		parsed, err := parser.ParseFile(f, i, catch, sin, com)
		if err != nil {
			t.Errorf(err.Error())
		}

		_, err = builder.BuildFiles(*parsed, "test1", 4, false, false)
		if err != nil {
			t.Errorf(err.Error())
		}
		f1 := "test1/" + test.args.FileName
		f2 := "test2/test1/" + test.args.FileName
		f = "test1/" + test.args.FileName
		parsed1, err2 := parser.ParseFile(f, i, catch, sin, com)
		if err2 != nil {
			t.Fatal(err)
		}
		if len(parsed1.Config) < 1 {
			t.Errorf("No configurations parsed for %s", test.args.FileName)
		}
		_, err = builder.BuildFiles(*parsed1, "test2", 4, false, false)
		if err != nil {
			t.Errorf(err.Error())
		}
		result, result2 := compareFiles(f1, f2)
		if len(result) != 0 && len(result2) != 0 {
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
