package parser

import (
	"fmt"
	"testing"

	"github.com/nginxinc/crossplane-go/pkg/lexer"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		title    string
		arg      ParseArgs
		file     string
		testdata []lexer.LexicalItem
		config   []Block
	}{
		{

			"basic : test Parse ",
			ParseArgs{
				FileName:    "config/simple.conf",
				CatchErrors: false,
				Ignore:      []string{},
				Single:      false,
				Strict:      false,
				Combine:     false,
			},
			"config/simple.conf",
			[]lexer.LexicalItem{
				{Item: "events", LineNum: 1},
				{Item: "{", LineNum: 1},
				{Item: "worker_connections", LineNum: 2},
				{Item: "1024", LineNum: 2},
				{Item: ";", LineNum: 2},
				{Item: "}", LineNum: 3},
				{Item: "http", LineNum: 5},
				{Item: "{", LineNum: 5},
				{Item: "server", LineNum: 6},
				{Item: "{", LineNum: 6},
				{Item: "listen", LineNum: 7},
				{Item: "127.0.0.1:8080", LineNum: 7},
				{Item: ";", LineNum: 7},
				{Item: "server_name", LineNum: 8},
				{Item: "default_server", LineNum: 8},
				{Item: ";", LineNum: 8},
				{Item: "location", LineNum: 9},
				{Item: "/", LineNum: 9},
				{Item: "{", LineNum: 9},
				{Item: "return", LineNum: 10},
				{Item: "200", LineNum: 10},
				{Item: "foo bar baz", LineNum: 10},
				{Item: ";", LineNum: 10},
				{Item: "}", LineNum: 11},
				{Item: "}", LineNum: 12},
				{Item: "}", LineNum: 13},
			},
			// need payload struct
			[]Block{
				{
					Directive: "events",
					Line:      1,
					Args:      []string{},
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block: []Block{
						{
							Directive: "worker_connections",
							Line:      2,
							Args:      []string{"1024"},
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []Block{},
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
					Block: []Block{
						{
							Directive: "server",
							Line:      6,
							Args:      []string{},
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block: []Block{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      7,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      8,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      9,
									Includes:  []int{},
									File:      "",
									Comment:   "",
									Block: []Block{
										{
											Directive: "return",
											Args:      []string{"200", "foo bar baz"},
											Line:      10,
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
		},
		{
			"Test : with Comments",
			ParseArgs{
				FileName:    "config/WithComments.conf",
				CatchErrors: false,
				Ignore:      []string{},
				Single:      false,
				Strict:      false,
				Combine:     false,
				checkArgs:   false,
				checkCtx:    false,
				Comments:    true,
			},
			"config/WithComments.conf",
			[]lexer.LexicalItem{
				{Item: "http", LineNum: 1},
				{Item: "{", LineNum: 1},
				{Item: "server", LineNum: 2},
				{Item: "{", LineNum: 2},
				{Item: "listen", LineNum: 3},
				{Item: "127.0.0.1:8080", LineNum: 3},
				{Item: ";", LineNum: 3},
				{Item: "#listen", LineNum: 3},
				{Item: "}", LineNum: 4},
				{Item: "}", LineNum: 5},
			},
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
							},
						},
					},
				},
			},
		},
		{
			"basic : messy test",
			ParseArgs{
				FileName:    "config/messy.conf",
				CatchErrors: false,
				Ignore:      []string{},
				Single:      false,
				Strict:      false,
				Combine:     false,
				Comments: true,
			},
			"config/messy.conf",
			[]lexer.LexicalItem{
				{Item: "user", LineNum: 1},
				{Item: "nobody", LineNum: 1},
				{Item: ";", LineNum: 1},
				{Item: "# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n \\", LineNum: 2},
				{Item: "events", LineNum: 3},
				{Item: "{", LineNum: 3},
				{Item: "worker_connections", LineNum: 3},
				{Item: "2048", LineNum: 3},
				{Item: ";", LineNum: 3},
				{Item: "}", LineNum: 3},
			},
			[]Block{
				{
					Directive: "user",
					Args:      []string{"nobody"},
					Line:      1,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block:     []Block{},
				},
				{
					Directive: "#",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					File:      "",
					Comment:   " hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n \\",
					Block:     []Block{},
				},
				{
					Directive: "events",
					Args:      []string{},
					Line:      3,
					Includes:  []int{},
					File:      "",
					Comment:   "",
					Block:     []Block{
						{
							Directive: "worker_connections",
							Args: []string{"2048"},
							Line: 3,
							Comment: "",
							Includes: []int{},
							File: "",
							Block: []Block{},
						},
					},
				},
				{
					Directive: "http",
					Args:      []string{""},
					Line:      5,
					Comment:   "forteen",
					Includes:  []int{},
					File:      "",
					Block: []Block{
						{
							Directive: "",
							Args:      []string{},
							Line:      6,
							Comment:   "this is a comment",
							Includes:  []int{},
							File:      "",
							Block:     []Block{},
						},
						{
							Directive: "access_log",
							Args:      []string{"off"},
							Line:      7,
							Comment:   "",
							Includes:  []int{},
							File:      "",
							Block:     []Block{},
						},
						{
							Directive: "default_type",
							Args:      []string{"text/plain"},
							Line:      7,
							Comment:   "",
							Includes:  []int{},
							File:      "",
							Block:     []Block{},
						},
						{
							Directive: "error_log",
							Args:      []string{"off"},
							Line:      7,
							Comment:   "",
							Includes:  []int{},
							File:      "",
							Block:     []Block{},
						},
						{
							Directive: "server",
							Args:      []string{"{"},
							Line:      8,
							Comment:   "",
							Includes:  []int{},
							File:      "",
							Block: []Block{
								{
									Directive: "listen",
									Args:      []string{"8080"},
									Line:      9,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "return",
									Args:      []string{`200","Ser\" \' \' ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n`},
									Line:      10,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
							},
						},
						{
							Directive: "server",
							Args:      []string{"{"},
							Line:      12,
							Comment:   "",
							Includes:  []int{},
							File:      "",
							Block: []Block{
								{
									Directive: "listen",
									Args:      []string{"8080"},
									Comment:   "",
									Line:      12,
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "root",
									Args:      []string{"/usr/share/nginx/html"},
									Line:      13,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"~", "/hello/world"},
									Comment:   "",
									Line:      14,
									Includes:  []int{},
									File:      "",
									Block: []Block{
										{
											Directive: "return",
											Args:      []string{"301", "status.html"},
											Line:      14,
											Comment:   "",
											File:      "",
											Includes:  []int{},
											Block:     []Block{},
										},
									},
								},
								{
									Directive: "location",
									Args:      []string{"/foo"},
									Line:      15,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/bar"},
									Line:      15,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/\\{\\;\\}\\ #\\ ab"},
									Line:      16,
									Comment:   "hello",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "if",
									Args:      []string{"$request_method", "=", "P\\{O\\)\\###\\;ST"},
									Line:      17,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
								{
									Directive: "location",
									Args:      []string{"/status.html"},
									Line:      18,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block: []Block{
										{
											Directive: "try_files",
											Args:      []string{"/abc/${uri} /abc/${uri}.html", "=404"},
											Line:      19,
											Comment:   "",
											Includes:  []int{},
											File:      "",
											Block:     []Block{},
										},
									},
								},

								{
									Directive: "location",
									Args:      []string{"/sta;\n                    tus"},
									Line:      20,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block: []Block{
										{
											Directive: "return",
											Args:      []string{"302", "/status.html"},
											Line:      21,
											Comment:   "",
											Includes:  []int{},
											File:      "",
											Block:     []Block{},
										},
									},
								},

								{
									Directive: "location",
									Args:      []string{"/upstream_conf"},
									Line:      23,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block: []Block{
										{
											Directive: "return",
											Args:      []string{"200", "/status.html"},
											Line:      23,
											Comment:   "",
											Includes:  []int{},
											File:      "",
											Block:     []Block{},
										},
									},
								},
								{
									Directive: "server",
									Args:      []string{},
									Line:      24,
									Comment:   "",
									Includes:  []int{},
									File:      "",
									Block:     []Block{},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tes := range tests {
		p := Payload{
			Status: "ok",
			Errors: []ParseError{},
			Config: []Config{},
		}
		q := Config{
			File:   tes.file,
			Status: "ok",
			Errors: []ParseError{},
			Parsed: []Block{},
		}
		gen, _, e := parse(q, p, tes.testdata, tes.arg, [3]string{}, false)
		if e != nil {
			t.Error(e)
		}
		for p := 0; p < len(gen); p++ {
			o := compareBlocks(gen[p], tes.config[p])
			if o != "" {
				t.Error(o)
			}
		}
		var ignore []string
		parsed, err := Parse(q.File, true, ignore, false, false, false, false, false, true, true)
		fmt.Println(parsed)
		if err != nil {
			fmt.Println("something")
		}
	}
}

func compareBlocks(gen Block, config Block) string {
	s := ""
	if gen.Directive != config.Directive {
		s += "Error with directives : " + gen.Directive + " && " + config.Directive
	}
	// loop over and compare
	if len(gen.Args) == len(config.Args) {
		for i := 0; i < len(gen.Args); i++ {
			if gen.Args[i] != config.Args[i] {
				s += "Problem with Args in Block " + gen.Directive + " && " + config.Directive
			}
		}
	} else {
		s += "Problem with Args in Block " + gen.Directive + " && " + config.Directive
	}
	if gen.Line != config.Line {
		s += "Problem with Line in Block " + gen.Directive + " && " + config.Directive
	}
	if gen.File != config.File {
		s += "Problem with File in Block " + gen.Directive + " && " + config.Directive
	}
	if gen.Comment != config.Comment {
		s += "Problem with Comments in Block " + gen.Comment + " && " + config.Comment
	}
	if len(gen.Includes) == len(config.Includes) {
		for i := 0; i < len(gen.Includes); i++ {
			if gen.Includes[i] != config.Includes[i] {
				s += "Problem with Includes in Block " + gen.Directive + " && " + config.Directive
			}
		}
	} else {
		s += "Problem with Comments in Block " + gen.Directive + " && " + config.Directive
	}
	for i := 0; i < len(gen.Block); i++ {
		s += compareBlocks(gen.Block[i], config.Block[i])
	}

	return s
}
