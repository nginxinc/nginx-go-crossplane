package builder

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuilder(t *testing.T) {
	var tests = []struct {
		title    string
		input    []Block
		expected string
	}{
		{
			"basic: build with comments",
			[]Block{
				{
					Directive: "http",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []Block{
						{
							Directive: "server",
							Args:      []string{},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []Block{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      3,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      3,
									Includes:  []int{},
									File:      "",
									Comment:   "listen",
									Block:     []Block{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      4,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      5,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      5,
									Includes:  []int{},
									File:      "",
									Comment:   "# this is brace",
									Block:     []Block{},
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
			[]Block{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []Block{
						{
							Directive: "worker_connections",
							Args:      []string{"1024"},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []Block{},
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
					Block: []Block{
						{
							Directive: "server",
							Args:      []string{},
							Line:      5,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []Block{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      6,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      7,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      8,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block: []Block{
										{
											Directive: "return",
											Args:      []string{"200", "foo bar baz"},
											Line:      9,
											Includes:  []int{},
											File:      "",
											Comment:   "",
											Block:     []Block{},
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
			[]Block{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block:     []Block{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []Block{
						{
							Directive: "include",
							Args:      []string{"conf.d/server.conf"},
							Line:      3,
							Includes:  []int{1},
							File:      "",
							Comment:   "",
							Block:     []Block{},
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
			[]Block{
				{
					Directive: "#",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "comment",
					Block:     []Block{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []Block{
						{
							Directive: "server",
							Args:      []string{},
							Line:      3,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []Block{},
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
		title string
		file  string
		input []Payload
	}{
		{
			"basic: build with comments",
			"config/simple.conf",
			[]Payload{
				{
					Status: "ok",
					Errors: []ParseError{},
					Config: []Config{
						{
							File:   "nginx.conf",
							Status: "ok",
							Errors: []ParseError{},
							Parsed: []Block{
								{
									Directive: "http",
									Args:      []string{},
									Line:      1,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block: []Block{
										{
											Directive: "server",
											Args:      []string{},
											Line:      2,
											Includes:  []int{},
											File:      "",
											Comment:   "",
											Block: []Block{
												{
													Directive: "listen",
													Args:      []string{"127.0.0.1:8080"},
													Line:      3,
													Includes:  []int{},
													File:      "",
													Comment:   "",
													Block:     []Block{},
												},
												{
													Directive: "#",
													Args:      []string{},
													Line:      3,
													Includes:  []int{},
													File:      "",
													Comment:   "listen",
													Block:     []Block{},
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
		out, err := json.Marshal(test.input)
		if err != nil {
			t.Errorf("Error %v", err)
		}
		result, err := BuildFiles(string(out), test.file, 4, false, false)
		if err != nil {
			t.Error(test.title)
		}
		if result == " " {
			t.Error(test.title)
		}
	}
}
