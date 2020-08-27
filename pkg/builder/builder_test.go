package builder

import (
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
			`http {
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
			`events;
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
		result := Build(test.input, &Options{Indent: 4, Tabs: false})
		expected := strings.TrimLeft(
			strings.ReplaceAll(test.expected, "\t", strings.Repeat(" ", 4)),
			"\n",
		)

		if expected != result {
			t.Errorf("\nexpected:\n%s\ngot:\n%s\n", test.expected, result)
		}
	}
}

func BenchmarkBuild(b *testing.B) {
	// TODO: The signature of Parse is well out of hand...
	input, _ := parser.ParseFile("testdata/nginx-full.conf", nil, false, false, false)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Build(input.Config[0].Parsed, &Options{Indent: 4, Tabs: false})
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
			"testdata/simple.conf",
			parser.Payload{

				Errors: []parser.ParseError{},
				Config: []*parser.Config{
					{
						File:   "testdata/simple.conf",
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
			"testdata/with-comments.conf",
			parser.Payload{

				Errors: []parser.ParseError{},
				Config: []*parser.Config{
					{
						File:   "testdata/with-comments.conf",
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

		expected := strings.TrimLeft(
			strings.ReplaceAll(test.expected, "\t", strings.Repeat(" ", 4)),
			"\n",
		)

		if result != expected {
			t.Error(test.title)
		}
	}
}
