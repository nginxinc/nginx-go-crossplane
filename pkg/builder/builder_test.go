package builder

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/nginxinc/crossplane-go/pkg/parser"
)

func TestBuilder(t *testing.T) {
	var tests = []struct {
		title    string
		input    []parser.Block
		expected string
	}{
		{
			"basic: build with comments",
			[]parser.Block{
				{
					Directive: "http",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []parser.Block{
						{
							Directive: "server",
							Args:      []string{},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []parser.Block{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      3,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []parser.Block{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      3,
									Includes:  []int{},
									File:      "",
									Comment:   "listen",
									Block:     []parser.Block{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      4,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []parser.Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      5,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []parser.Block{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      5,
									Includes:  []int{},
									File:      "",
									Comment:   "# this is brace",
									Block:     []parser.Block{},
								},
							},
						},
					},
				},
			},
			`
				http {
					server {
						listen 127.0.0.1:8080; #listen
						server_name default_server;
						location /; ## this is brace
					}
				}`,
		},
		{
			"basic: build nested and multiple args",
			[]parser.Block{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []parser.Block{
						{
							Directive: "worker_connections",
							Args:      []string{"1024"},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []parser.Block{},
						},
					},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      4,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []parser.Block{
						{
							Directive: "server",
							Args:      []string{},
							Line:      5,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []parser.Block{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      6,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []parser.Block{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      7,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []parser.Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      8,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block: []parser.Block{
										{
											Directive: "return",
											Args:      []string{"200", "foo bar baz"},
											Line:      9,
											Includes:  []int{},
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
			`
				events {
					worker_connections 1024;
				}
				http {
					server {
						listen 127.0.0.1:8080;
						server_name default_server;
						location / {
							return 200 foo bar baz;
						}
					}
				}`,
		},
		{
			"basic: build include regular",
			[]parser.Block{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block:     []parser.Block{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []parser.Block{
						{
							Directive: "include",
							Args:      []string{"conf.d/server.conf"},
							Line:      3,
							Includes:  []int{1},
							File:      "",
							Comment:   "",
							Block:     []parser.Block{},
						},
					},
				},
			},
			`
				events;
				http {
					include conf.d/server.conf;
				}`,
		},
		{
			"basic: start with comment",
			[]parser.Block{
				{
					Directive: "#",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "comment",
					Block:     []parser.Block{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []parser.Block{
						{
							Directive: "server",
							Args:      []string{},
							Line:      3,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []parser.Block{},
						},
					},
				},
			},
			`
				#comment
				http {
					server;
				}`,
		},
	}

	for _, test := range tests {
		out, err := json.Marshal(test.input)
		if err != nil {
			t.Errorf("Error %v", err)
		}
		result, err := Build(string(out), 4, false, false)

		test.expected = strings.Replace(test.expected, "\t", padding, -1)

		if err != nil {
			t.Error(test.title)
		}
		if result != test.expected {
			t.Error(test.title)
		}
	}
}

func TestBuildFile(t *testing.T) {
	var tests = []struct {
		title    string
		file     string
		input    parser.Payload
		expected string
	}{
		{
			"basic: simple build files",
			"config/simple.conf",
			parser.Payload{

				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "config/simple.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "events",
								Line:      1,
								Args:      []string{},
								Includes:  []int{},
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
										Includes:  []int{},
										File:      "",
										Comment:   "",
										Block:     []parser.Block{},
									},
								},
							},
							{
								Directive: "http",
								Line:      5,
								Args:      []string{},
								Includes:  []int{},
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "server",
										Line:      6,
										Args:      []string{},
										Includes:  []int{},
										File:      "",
										Comment:   "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      7,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block:     []parser.Block{},
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Line:      8,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block:     []parser.Block{},
											},
											{
												Directive: "location",
												Args:      []string{"/"},
												Line:      9,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block: []parser.Block{
													{
														Directive: "return",
														Args:      []string{"200", "foo bar baz"},
														Line:      10,
														Includes:  []int{},
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
			},
			`
			events {
				worker_connections 1024;
			}
			http {
				server {
					listen 127.0.0.1:8080;
					server_name default_server;
					location / {
						return 200 foo bar baz;
					}
				}
			}`,
		},
		{
			"basic: with comments build files",
			"config/withComments.conf",
			parser.Payload{

				Status: "ok",
				Errors: []parser.ParseError{},
				Config: []parser.Config{
					{
						File:   "config/withComments.conf",
						Status: "ok",
						Errors: []parser.ParseError{},
						Parsed: []parser.Block{
							{
								Directive: "http",
								Args:      []string{},
								Line:      1,
								Includes:  []int{},
								File:      "",
								Comment:   "",
								Block: []parser.Block{
									{
										Directive: "server",
										Args:      []string{},
										Line:      2,
										Includes:  []int{},
										File:      "",
										Comment:   "",
										Block: []parser.Block{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      3,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block:     []parser.Block{},
											},
											{
												Directive: "#",
												Args:      []string{},
												Line:      3,
												Includes:  []int{},
												File:      "",
												Comment:   "listen",
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
			`
				http {
					server {
						listen 127.0.0.1:8080; #listen
					}
				}`,
		},
	}

	for _, test := range tests {
		result, err := BuildFiles(test.input, " ", 0, false, false)
		fmt.Println("HELLO : ", result)
		test.expected = strings.TrimLeft(test.expected, "\n")
		test.expected = strings.Replace(test.expected, "\t", padding, -1)
		fmt.Println()
		fmt.Println(result)

		if err != nil {
			t.Error(test.title)
		}
		if result != test.expected {
			t.Error(test.title)
		}
	}
}
