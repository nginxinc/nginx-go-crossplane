package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		title    string
		arg      ParseArgs
		file     string
		testdata []LexicalItem
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
			[]LexicalItem{
				{item: "events", lineNum: 1},
				{item: "{", lineNum: 1},
				{item: "worker_connections", lineNum: 2},
				{item: "1024", lineNum: 2},
				{item: ";", lineNum: 2},
				{item: "}", lineNum: 3},
				{item: "http", lineNum: 5},
				{item: "{", lineNum: 5},
				{item: "server", lineNum: 6},
				{item: "{", lineNum: 6},
				{item: "listen", lineNum: 7},
				{item: "127.0.0.1:8080", lineNum: 7},
				{item: ";", lineNum: 7},
				{item: "server_name", lineNum: 8},
				{item: "default_server", lineNum: 8},
				{item: ";", lineNum: 8},
				{item: "location", lineNum: 9},
				{item: "/", lineNum: 9},
				{item: "{", lineNum: 9},
				{item: "return", lineNum: 10},
				{item: "200", lineNum: 10},
				{item: "foo bar baz", lineNum: 10},
				{item: ";", lineNum: 10},
				{item: "}", lineNum: 11},
				{item: "}", lineNum: 12},
				{item: "}", lineNum: 13},
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
			[]LexicalItem{
				{item: "http", lineNum: 1},
				{item: "{", lineNum: 1},
				{item: "server", lineNum: 2},
				{item: "{", lineNum: 2},
				{item: "listen", lineNum: 3},
				{item: "127.0.0.1:8080", lineNum: 3},
				{item: ";", lineNum: 3},
				{item: "#listen", lineNum: 4},
				{item: "}", lineNum: 5},
				{item: "}", lineNum: 6},
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
									Directive: "",
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

/*
		{

			"basic : messy test",
			"config/messy.conf",
			ParseArgs{
				FileName:    "config/messy.conf",
				CatchErrors: false,
				Ignore:      []string{},
				Single:      false,
				Strict:      false,
				Combine:     false,
			},
			[]Config{
				{
					File:   "config/messy.conf",
					Status: "ok",
					Errors: []ParseErrors{},
					Parsed: []Block{
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
							Directive: "",
							Args:      []string{},
							Line:      2,
							Includes:  []int{},
							File:      "",
							Comment:   "hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n ",
							Block:     []Block{},
						},
						{
							Directive: "events",
							Args:      []string{"worker_connections", "2048"},
							Line:      3,
							Includes:  []int{},
							File:      "",
							Comment:   "",
							Block:     []Block{},
						},
						{
							Directive: "http",
							Args:      []string{"{"},
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
			},
		},
	{

		"basic : testwith comments",
		"config/withComments.conf",
		ParseArgs{
			FileName:    "config/withComments.conf",
			CatchErrors: false,
			Ignore:      []string{},
			Single:      false,
			Strict:      false,
			Combine:     false,
		},
		[]Config{
			{
				File:   "config/withComments.conf",
				Status: "ok",
				Errors: []ParseErrors{},
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
								Line:      2,
								Args:      []string{},
								Includes:  []int{},
								File:      "",
								Comment:   "",
								Block: []Block{
									{
										Directive: "listen",
										Args:      []string{"120.0.0.1:8080"},
										Line:      3,
										Comment:   "listen",
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
		},
	},*/
