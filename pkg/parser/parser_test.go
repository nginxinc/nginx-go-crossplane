package parser

import (
	"fmt"
	"reflect"
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
									Args:      []string{"120.0.0.1:8080"},
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
	}

	for _, tes := range tests {
		p := Payload{
			Status: "ok",
			Errors: []ParseErrors{},
			Config: []Config{},
		}
		q := Config{
			File:   tes.file,
			Status: "ok",
			Errors: []ParseErrors{},
			Parsed: []Block{},
		}
		gen, _ := parse(q, p, tes.testdata, tes.arg, [3]string{}, false)
		//g, e := json.Marshal(gen)
		/*if e != nil {
			panic(e)
		}*/
		fmt.Printf("%+v\n", gen[0].Args)

		//dat, err := json.Marshal(tes.config)
		fmt.Printf("%+v\n", tes.config[0].Args)
		/*if err != nil {
			panic(err)
		}*/
		isEqual := reflect.DeepEqual(tes.config[0].Args, gen[0].Args)
		fmt.Println(isEqual)
		if !isEqual {
			t.Errorf("%v Failed : Generated Data is not the same", tes.title)
		}
		// deep equals etcs

	}
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
