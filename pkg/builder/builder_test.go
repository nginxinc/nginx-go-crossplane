package builder

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBuilderUltraSimple(t *testing.T) {
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
									Line:      4,
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
			`
			http {
				server {
					listen 127.0.0.1:8080;
					#listen;
				}
			}
			`,
		},
	}

	for _, test := range tests {
		out, err := json.Marshal(test.input)
		if err != nil {
			t.Errorf("Error %v", err)
		}

		result, err := Build(string(out), 4, false, false)
		if err != nil {
			t.Error(test.title)
		}
		fmt.Println(result)
		/*
			if result != test.expected {
				t.Error(test.title)
			}
		*/
	}
}

/*
func TestBuild(t *testing.T) {
	var tests = []struct {
		title   string
		payload []Block
	}{
		{
			"Build: NestedAndMultipleArgs",
			[]Block{
				{
					Directive: "events",
					Args:      []string{" "},
					Block: []Block{
						Block{
							Directive: "worker_connections",
							Args:      []string{"1024"},
						},
					},
				},
				{
					Directive: "http",
					Args:      []string{" "},
					Block: []Block{
						Block{
							Directive: "server",
							Args:      []string{" "},
							Block: []Block{
								Block{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
								},
								Block{
									Directive: "server_name",
									Args:      []string{"default_server"},
								},
								Block{
									Directive: "location",
									Args:      []string{"/"},
									Block: []Block{
										Block{
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
		},
		{
			"Build: WithComments",
			[]Block{
				{
					Directive: "events",
					Line:      1,
					Args:      []string{" "},
					Block: []Block{
						Block{
							Directive: "worker_connections",
							Line:      2,
							Args:      []string{"1024"},
						},
					},
				},
				{
					Directive: "#",
					Line:      4,
					Args:      []string{" "},
					Comment:   "comment",
				},
				{
					Directive: "http",
					Line:      5,
					Args:      []string{" "},
					Block: []Block{
						Block{
							Directive: "server",
							Line:      6,
							Args:      []string{" "},
							Block: []Block{
								Block{
									Directive: "listen",
									Line:      7,
									Args:      []string{"127.0.0.1:8080"},
								},
								Block{
									Directive: "#",
									Line:      7,
									Args:      []string{" "},
									Comment:   "listen",
								},
								Block{
									Directive: "server_name",
									Line:      8,
									Args:      []string{"default_server"},
								},
								Block{
									Directive: "location",
									Line:      9,
									Args:      []string{"/"},
									Block: []Block{
										Block{
											Directive: "#",
											Line:      9,
											Args:      []string{" "},
											Comment:   "# this is brace",
										},
										Block{
											Directive: "#",
											Line:      10,
											Args:      []string{" "},
											Comment:   " location /",
										},
										Block{
											Directive: "#",
											Line:      11,
											Args:      []string{" "},
											Comment:   " is here",
										},
										Block{
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
		},
		{
			"Build: StartsWithComments",
			[]Block{
				{
					Directive: "#",
					Line:      1,
					Args:      []string{" "},
					Comment:   " foo",
				},
				{
					Directive: "user",
					Line:      5,
					Args:      []string{"root"},
				},
			},
		},
		{
			"Build: WithQuotedUnicode",
			[]Block{
				{
					Directive: "env",
					Line:      1,
					Args:      []string{"русский текст"},
				},
			},
		},
	}
	for _, test := range tests {
		out, err := json.Marshal(test)
		if err != nil {
			t.Errorf("Error %v", err)
		}

		c, err := Build(string(out), 4, false, false)
		if err != nil {
			t.Errorf("%v: test failed due to error being returned from Build", test.title)
		}
		if c != "built" {
			t.Errorf("expected %s but got %s", "built", c)
		}
	}
}

func TestBuildFiles(t *testing.T) {
	var tests = []struct {
		title   string
		payload []ConfFiles
	}{
		{
			"BuildFiles: WithMissingStatusAndErrors",
			[]ConfFiles{
				{
					File: "nginx.conf",
					Parsed: []Block{
						Block{
							Directive: "user",
							Line:      1,
							Args:      []string{"nginx"},
						},
					},
				},
			},
		},
		{
			"BuildFiles: WithUnicode",
			[]ConfFiles{
				{
					Status: "ok",
					Errors: " ",
					Config: []ConfFiles{
						ConfFiles{
							File:   "nginx.conf",
							Status: "ok",
							Errors: " ",
							Parsed: []Block{
								Block{
									Directive: "user",
									Line:      1,
									Args:      []string{"測試"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		out, err := json.Marshal(test)
		if err != nil {
			t.Errorf("Error %v", err)
		}

		c, err := BuildFiles(string(out), "none", 4, false, false)
		if err != nil {
			t.Errorf("%v: test failed due to error being returned from Build", test.title)
		}
		if c != "built" {
			t.Errorf("expected %s but got %s", "built", c)
		}
	}
}
*/
