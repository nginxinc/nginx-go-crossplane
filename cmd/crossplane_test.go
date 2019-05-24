package cmd

import (
	"fmt"
	"log"
	"testing"

	"github.com/nginxinc/crossplane-go/pkg/parser"
)

// \ char causing problems in lexing to parsing
func TestParseAndBuild(t *testing.T) {
	var tests = []struct {
		name     string
		args     parser.ParseArgs
		expected parser.Payload
	}{
		{
			"bad-args",
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
				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "configs/bad-args/nginx.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "user",
								Args:      []string{},
								Line:      1,
								File:      "configs/bad-args/nginx.conf",
								Comment:   "",
								Block:     []parser.Block{},
							}, {
								Directive: "events",
								Args:      []string{},
								Line:      2,
								Comment:   "",
								File:      "configs/bad-args/nginx.conf",
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Args:      []string{},
								Line:      3,
								Comment:   "",
								Block:     []parser.Block{},
								File:      "configs/bad-args/nginx.conf",
							},
						},
					},
				},
			},
		},
		{
			"directive-with-space",
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
								File:      "configs/directive-with-space/nginx.conf",
								Line:      1,
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Args:      []string{},
								Comment:   "",
								Line:      3,
								File:      "configs/directive-with-space/nginx.conf",
								Block: []parser.Block{
									{
										Directive: "map",
										Args:      []string{"$http_user_agent", "$mobile"},
										Line:      4,
										File:      "configs/directive-with-space/nginx.conf",
										Comment:   "",
										Block: []parser.Block{
											{
												Directive: "default",
												Args:      []string{"0"},
												Line:      5,
												Comment:   "",
												File:      "configs/directive-with-space/nginx.conf",
												Block:     []parser.Block{},
											}, {
												Directive: "\\'~Opera Mini\\'",
												Args:      []string{"1"},
												Line:      6,
												Comment:   "",
												File:      "configs/directive-with-space/nginx.conf",
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
			"empty-value-map",
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
								File:      "configs/empty-value-map/nginx.conf",
								Block:     []parser.Block{},
							}, {
								Directive: "http",
								Line:      3,
								Args:      []string{},
								Comment:   "",
								File:      "configs/empty-value-map/nginx.conf",
								Block: []parser.Block{
									{
										Directive: "map",
										Args:      []string{"string", "$variable"},
										Line:      4,
										Comment:   "",
										File:      "configs/empty-value-map/nginx.conf",
										Block: []parser.Block{
											{
												Directive: "\\'\\'",
												Args:      []string{"$arg"},
												Comment:   "",
												Line:      5,
												File:      "configs/empty-value-map/nginx.conf",
												Block:     []parser.Block{},
											},
											{
												Directive: "*.example.com",
												Args:      []string{"\\'\\'"},
												Line:      6,
												File:      "configs/empty-value-map/nginx.conf",
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
			"includes-globbed",
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
								File:      "configs/includes-globbed/http.conf ",
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
														File:      "configs/includes-globbed/nginx.conf",
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
		}, /*
				{
					"includes-regular",
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
				},
			{
				"lua-block-larger",
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
			},
			{
				"lua-block-tricky",
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
			},
			{
				"lua-block-tricky",
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
			},*/
		{
			"messy",
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
								Comment:   " hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n",
								Block:     []parser.Block{},
							},
							{
								Directive: "events",
								Args:      []string{},
								Line:      3,
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "worker_connections",
										Args:      []string{"2048"},
										Line:      3,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
								},
							},

							{
								Directive: "http",
								Args:      []string{""},
								Line:      5,
								Comment:   "",
								File:      "",
								Block: []parser.Block{
									{
										Directive: "#",
										Args:      []string{},
										Line:      6,
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
										Directive: "access_log",
										Args:      []string{"off"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "default_type",
										Args:      []string{"text/plain"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "error_log",
										Args:      []string{"off"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "server",
										Args:      []string{"{"},
										Line:      8,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"8083"},
												Line:      9,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "return",
												Args:      []string{`200","Ser\" \' \' ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n`},
												Line:      10,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
										},
									},
									{
										Directive: "server",
										Args:      []string{"{"},
										Line:      12,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"8080"},
												Comment:   "",
												Line:      12,
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "root",
												Args:      []string{"/usr/share/nginx/html"},
												Line:      13,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"~", "/hello/world"},
												Comment:   "",
												Line:      14,
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"301", "status.html"},
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
												Args:      []string{"/status.html"},
												Line:      18,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "try_files",
														Args:      []string{"/abc/${uri} /abc/${uri}.html", "=404"},
														Line:      19,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},

											{
												Directive: "location",
												Args:      []string{"/sta;\n                    tus"},
												Line:      20,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"302", "/status.html"},
														Line:      21,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},

											{
												Directive: "location",
												Args:      []string{"/upstream_conf"},
												Line:      23,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "/status.html"},
														Line:      23,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
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
			},
		},
	}

	/*
			{
				"lua-block-tricky",
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
			},*/
		{
			"messy",
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
								Comment:   " hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n",
								Block:     []parser.Block{},
							},
							{
								Directive: "events",
								Args:      []string{},
								Line:      3,
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "worker_connections",
										Args:      []string{"2048"},
										Line:      3,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
								},
							},

							{
								Directive: "http",
								Args:      []string{""},
								Line:      5,
								Comment:   "",
								File:      "",
								Block: []parser.Block{
									{
										Directive: "#",
										Args:      []string{},
										Line:      6,
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
										Directive: "access_log",
										Args:      []string{"off"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "default_type",
										Args:      []string{"text/plain"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "error_log",
										Args:      []string{"off"},
										Line:      7,
										Comment:   "",
										File:      "",
										Block:     []parser.Block{},
									},
									{
										Directive: "server",
										Args:      []string{"{"},
										Line:      8,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"8083"},
												Line:      9,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "return",
												Args:      []string{`200","Ser\" \' \' ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n`},
												Line:      10,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
										},
									},
									{
										Directive: "server",
										Args:      []string{"{"},
										Line:      12,
										Comment:   "",
										File:      "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"8080"},
												Comment:   "",
												Line:      12,
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "root",
												Args:      []string{"/usr/share/nginx/html"},
												Line:      13,
												Comment:   "",
												File:      "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"~", "/hello/world"},
												Comment:   "",
												Line:      14,
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"301", "status.html"},
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
												Args:      []string{"/status.html"},
												Line:      18,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "try_files",
														Args:      []string{"/abc/${uri} /abc/${uri}.html", "=404"},
														Line:      19,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},

											{
												Directive: "location",
												Args:      []string{"/sta;\n                    tus"},
												Line:      20,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"302", "/status.html"},
														Line:      21,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
													},
												},
											},

											{
												Directive: "location",
												Args:      []string{"/upstream_conf"},
												Line:      23,
												Comment:   "",
												File:      "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "/status.html"},
														Line:      23,
														Comment:   "",
														File:      "",
														Block:     []parser.Block{},
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
			},
		},
	}

	/*
			{
				"missing-semicolon",
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
			},
		{
			"quote-behavior",
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
		},
		{
			"quoted-right-brace",
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
		},
		{
			"russian-text",
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
		},
		{
			"simple",
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
		},
		{
			"spelling-mistake",
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
		},
		{
			"with-comments",
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
		},*/

	for _, t := range tests {
		t.args.FileName = "configs/" + t.name + "/nginx.conf"
		parsed, err := parser.Parse(t.args)
		fmt.Println("OUTPUT : ", parsed)
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range parsed.Config {
			for _, q := range p.Parsed {
				fmt.Println(q)

		if parsed.File != t.expected.File {
			//fmt.Println(parsed.File)
			//fmt.Println(t.expected.File)
			log.Fatal("Payload filenames not the same")
		}
		if parsed.Status != t.expected.Status {
			log.Fatal("status not teh same ")
		}
		if len(parsed.Errors) != 0 {
			for p := 0; p < len(parsed.Errors); p++ {
				if parsed.Errors[p] != t.expected.Errors[p] {
					log.Fatal("Error")
				}
			}
		}
		if len(parsed.Config) != len(t.expected.Config) {
			log.Fatal("Configs arent same length")
		} else {
			var w string
			for i := 0; i < len(parsed.Config)-1; i++ {
				w += compareConfigs(parsed.Config[i], t.expected.Config[i])
			}
			if w != "" {
				log.Fatal(w)
			}
		}

	}

}

func compareConfigs(conf parser.Config, c parser.Config) string {
	var s string
	if conf.File != c.File {
		s = "Problems with the names of config files" + string('\n')
	}
	if len(conf.Errors) != len(c.Errors) {
		s = "Errors are not the same length" + string('\n')
	}
	if conf.Status != c.Status {
		s = "the Status's are not the same" + string('\n')
	}
	for i := 0; i < len(c.Parsed)-1; i++ {
		s += compareBlocks(conf.Parsed[i], c.Parsed[i])
	}
	return s
}

func compareBlocks(gen parser.Block, config parser.Block) string {
	s := ""
	if gen.Directive != config.Directive {
		s += "Error with directives : " + gen.Directive + " && " + config.Directive + string('\n')
		//fmt.Println("gen : ", gen.Directive)
		//fmt.Println("expected : ", config.Directive)
	}
	// loop over and compare
	if len(gen.Args) == len(config.Args) {
		for i := 0; i < len(gen.Args); i++ {
			if gen.Args[i] != config.Args[i] {
				s += "Problem with Args in Block " + gen.Directive + " && " + config.Directive + string('\n')
				//fmt.Println("gen args : ", gen.Args)
				//fmt.Println("expected args : ", config.Args)

			}
		}
	} else {
		s += "Problem with Args in Block " + gen.Directive + " && " + config.Directive + string('\n')
		//fmt.Println("gen args : ", gen.Args)
		//fmt.Println("expected args : ", config.Args)
	}
	if gen.Line != config.Line {
		s += "Problem with Line in Block " + gen.Directive + " && " + config.Directive + string('\n')
		//fmt.Println("gen line : ", gen.Line, gen)
		//fmt.Println("expected line : ", config.Line, config)
	}
	if gen.File != config.File {
		s += "Problem with File in Block " + gen.Directive + " && " + config.Directive + string('\n')
		/*fmt.Println("gen file : ", gen.File)
		fmt.Println(gen)
		fmt.Println()
		fmt.Println("expected file : ", config.File)
		fmt.Println(config)
		fmt.Println()*/
	}
	if gen.Comment != config.Comment {
		s += "Problem with Comments in Block " + gen.Comment + " && " + config.Comment + string('\n')
		//fmt.Println("gen comments : ", gen.Comment)
		//fmt.Println("expected comments: ", config.Comment)
	}
	for i := 0; i < len(gen.Block)-1; i++ {
		s += compareBlocks(gen.Block[i], config.Block[i])
	}

	return s
}

func TestExecute(t *testing.T) {

}
