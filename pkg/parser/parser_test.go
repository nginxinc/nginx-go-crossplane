package parser

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		title    string
		filename string
		parargs  ParseArgs
		config   []Config
	}{
		{

			"basic : test Parse ",
			"config/simple.conf",
			ParseArgs{
				FileName:    "config/simple.conf",
				CatchErrors: false,
				Ignore:      []string{},
				Single:      false,
				Strict:      false,
				Combine:     false,
			},
			[]Config{
				{
					File:   "simple.conf",
					Status: "ok",
					Errors: []ParseErrors{},
					Parsed: []Block{
						{
							Directive: "http",
							Line:      1,
							Args:      []string{},
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
											Includes:  []int{},
											File:      "",
											Comment:   "",
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
	}

	for _, tes := range tests {

		gen := parse(tes.parargs)

		dat, err := json.Marshal(tes.config)
		if err != nil {
			panic(err)
		}
		isEqual := reflect.DeepEqual(dat, gen)
		if !isEqual {
			t.Errorf("%v Failed : Generated Data is not the same", tes.title)
		}
		// deep equals etcs

	}
}
