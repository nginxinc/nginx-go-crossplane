package builder

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

func TestBuilder(t *testing.T) {
	var tests = []struct {
		title    string
		input    []*parser.Directive
		expected string
	}{
		{
			"basic: build with comments",
			[]*parser.Directive{
				{
					Directive: "http",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "server",
							Args:      []string{},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []*parser.Directive{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      3,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      3,
									Includes:  []int{},
									File:      "",
									Comment:   "listen",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      4,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      5,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      5,
									Includes:  []int{},
									File:      "",
									Comment:   "# this is brace",
									Block:     []*parser.Directive{},
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
			[]*parser.Directive{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "worker_connections",
							Args:      []string{"1024"},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []*parser.Directive{},
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
					Block: []*parser.Directive{
						{
							Directive: "server",
							Args:      []string{},
							Line:      5,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []*parser.Directive{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      6,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      7,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      8,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block: []*parser.Directive{
										{
											Directive: "return",
											Args:      []string{"200", "foo bar baz"},
											Line:      9,
											Includes:  []int{},
											File:      "",
											Comment:   "",
											Block:     []*parser.Directive{},
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
			[]*parser.Directive{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block:     []*parser.Directive{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "include",
							Args:      []string{"conf.d/server.conf"},
							Line:      3,
							Includes:  []int{1},
							File:      "",
							Comment:   "",
							Block:     []*parser.Directive{},
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
			[]*parser.Directive{
				{
					Directive: "#",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "comment",
					Block:     []*parser.Directive{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "server",
							Args:      []string{},
							Line:      3,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []*parser.Directive{},
						},
					},
				},
			},
			`#comment
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
		if err != nil {
			t.Errorf(test.title, err)
		}
		test.expected = strings.Replace(test.expected, "\t", padding, -1)

		for i := 0; i < len(test.expected); i++ {
			if test.expected[i] != result[i] {
				t.Error(test.title)
			}
		}

		if !reflect.DeepEqual(result, test.expected) {
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
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
								Args:      []string{},
								Includes:  []int{},
								File:      "",
								Comment:   "",
								Block: []*parser.Directive{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
										Includes:  []int{},
										File:      "",
										Comment:   "",
										Block:     []*parser.Directive{},
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
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      6,
										Args:      []string{},
										Includes:  []int{},
										File:      "",
										Comment:   "",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      7,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Line:      8,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "location",
												Args:      []string{"/"},
												Line:      9,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "foo bar baz"},
														Line:      10,
														Includes:  []int{},
														File:      "",
														Comment:   "",
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
			`events {
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
						Parsed: []*parser.Directive{
							{
								Directive: "http",
								Args:      []string{},
								Line:      1,
								Includes:  []int{},
								File:      "",
								Comment:   "",
								Block: []*parser.Directive{
									{
										Directive: "server",
										Args:      []string{},
										Line:      2,
										Includes:  []int{},
										File:      "",
										Comment:   "",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      3,
												Includes:  []int{},
												File:      "",
												Comment:   "",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "#",
												Args:      []string{},
												Line:      3,
												Includes:  []int{},
												File:      "",
												Comment:   "listen",
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
			`http {
	server {
		listen 127.0.0.1:8080; #listen
	}
}`,
		},
	}

	for _, test := range tests {
		result, err := BuildFiles(test.input, " ", 0, false, false)
		if err != nil {
			t.Error(test.title)
		}

		test.expected = strings.TrimLeft(test.expected, "\n")
		test.expected = strings.Replace(test.expected, "\t", padding, -1)

		if result != test.expected {
			t.Error(test.title)
		}
	}
}
