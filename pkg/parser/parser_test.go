package parser

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/lexer"
)

var tests = []struct {
	title    string
	arg      ParseArgs
	file     string
	testdata []lexer.LexicalItem
	config   []*Directive
}{
	{
		"basic : test Parse ",
		ParseArgs{
			FileName:    "config/simple.conf",
			CatchErrors: true,
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
		[]*Directive{
			{
				Directive: "events",
				Line:      1,
				Block: []*Directive{
					{
						Directive: "worker_connections",
						Line:      2,
						Args:      []string{"1024"},
					},
				},
			},
			{
				Directive: "http",
				Line:      5,
				Block: []*Directive{
					{
						Directive: "server",
						Line:      6,
						Block: []*Directive{
							{
								Directive: "listen",
								Args:      []string{"127.0.0.1:8080"},
								Line:      7,
							},
							{
								Directive: "server_name",
								Args:      []string{"default_server"},
								Line:      8,
							},
							{
								Directive: "location",
								Args:      []string{"/"},
								Line:      9,
								Block: []*Directive{
									{
										Directive: "return",
										Args:      []string{"200", "foo bar baz"},
										Line:      10,
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
			FileName:    "config/withComments.conf",
			CatchErrors: true,
		},
		"config/withComments.conf",
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
		// TODO: this should be an interpolation of lexer.LexicalItem
		//       use the above definitions to build the tree below
		[]*Directive{
			{
				Directive: "http",
				Line:      1,
				Block: []*Directive{
					{
						Directive: "server",
						Line:      2,
						Block: []*Directive{
							{
								Directive: "listen",
								Args:      []string{"127.0.0.1:8080"},
								Line:      3,
							},
							{
								Directive: "#",
								Line:      3,
								Comment:   "listening",
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
			CatchErrors: true,
		},
		"config/messy.conf",
		[]lexer.LexicalItem{
			{Item: "user", LineNum: 1},
			{Item: "nobody", LineNum: 1},
			{Item: ";", LineNum: 1},
			{Item: "# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n", LineNum: 2},
			{Item: "events", LineNum: 3},
			{Item: "{", LineNum: 3},
			{Item: "worker_connections", LineNum: 3},
			{Item: "2048", LineNum: 3},
			{Item: ";", LineNum: 3},
			{Item: "}", LineNum: 3},
		},
		[]*Directive{
			{
				Directive: "user",
				Args:      []string{"nobody"},
				Line:      1,
			},
			{
				Directive: "#",
				Line:      2,
				Comment:   " hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n ", // removed last 2 "\\"
			},
			{
				Directive: "events",
				Line:      3,
				Block: []*Directive{
					{
						Directive: "worker_connections",
						Args:      []string{"2048"},
						Line:      3,
					},
				},
			},
			{
				Directive: "http",
				Line:      5,
				Block: []*Directive{
					{
						Directive: "#",
						Line:      5,
						Comment:   "forteen",
					},
					{
						Directive: "#",
						Line:      6,
						Comment:   " this is a comment",
					},
					{
						Directive: "access_log",
						Args:      []string{"off"},
						Line:      7,
					},
					{
						Directive: "default_type",
						Args:      []string{"text/plain"},
						Line:      7,
					},
					{
						Directive: "error_log",
						Args:      []string{"off"},
						Line:      7,
					},
					{
						Directive: "server",
						Line:      8,
						Block: []*Directive{
							{
								Directive: "listen",
								Args:      []string{"8083"},
								Line:      9,
							},
							{
								Directive: "return",
								Args:      []string{"200", `Ser" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n`},
								Line:      10,
							},
						},
					},
					{
						Directive: "server",
						Line:      12,
						Block: []*Directive{
							{
								Directive: "listen",
								Args:      []string{"8080"},
								Line:      12,
							},
							{
								Directive: "root",
								Args:      []string{"/usr/share/nginx/html"},
								Line:      13,
							},
							{
								Directive: "location",
								Args:      []string{"~", "/hello/world;"},
								Line:      14,
								Block: []*Directive{
									{
										Directive: "return",
										Args:      []string{"301", "/status.html"},
										Line:      14,
									},
								},
							},
							{
								Directive: "location",
								Args:      []string{"/foo"},
								Line:      15,
							},
							{
								Directive: "location",
								Args:      []string{"/bar"},
								Line:      15,
							},
							{
								Directive: "location",
								Args:      []string{"/\\{\\;\\}\\ #\\ ab"},
								Line:      16,
							},
							{
								Directive: "#",
								Line:      16,
								Comment:   " hello",
							},
							{
								Directive: "if",
								Args:      []string{"$request_method", "=", "P\\{O\\)\\###\\;ST", ""}, // PAUL added empty string to array
								Line:      17,
							},
							{
								Directive: "location",
								Args:      []string{"/status.html"},
								Line:      18,
								Block: []*Directive{
									{
										Directive: "try_files",
										Args:      []string{"/abc/${uri}", "/abc/${uri}.html", "=404"},
										Line:      19,
									},
								},
							},

							{
								Directive: "location",
								Args:      []string{"/sta;\n                    tus"},
								Line:      21,
								Block: []*Directive{
									{
										Directive: "return",
										Args:      []string{"302", "/status.html"},
										Line:      22,
									},
								},
							},

							{
								Directive: "location",
								Args:      []string{"/upstream_conf"},
								Line:      23,
								Block: []*Directive{
									{
										Directive: "return",
										Args:      []string{"200", "/status.html"},
										Line:      23,
									},
								},
							},
						},
					},
					{
						Directive: "server",
						Line:      24,
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {
	for _, tes := range tests {
		t.Logf("testing config: %s\n", tes.arg.FileName)
		parsed, err := ParseFile(tes.arg.FileName, tes.arg.Ignore, tes.arg.CatchErrors, tes.arg.Single, tes.arg.Comments)
		if parsed.Errors != nil {
			t.Errorf("Errors encountered: %v", parsed.Errors)
		}
		if len(parsed.Config) < 1 {
			t.Errorf("No configurations parsed for %s", tes.arg.FileName)
		}
		par := parsed.Config[0].Parsed
		for p := 0; p < len(par); p++ {
			compareDirectives(t, par[p], tes.config[p])
		}

		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseString(t *testing.T) {
	for _, tes := range tests {
		file, err := ioutil.ReadFile(tes.arg.FileName)
		if err != nil {
			t.Error(err)
			continue
		}
		t.Logf("parsing: %s\n", tes.arg.FileName)
		parsed, _ := ParseString(tes.arg.FileName, string(file), tes.arg.Ignore, tes.arg.CatchErrors, tes.arg.Single, tes.arg.Comments)
		if parsed.Errors != nil {
			t.Errorf("Errors encountered: %v", parsed.Errors)
		}
		if len(parsed.Config) < 1 {
			t.Errorf("No configurations parsed for %s", tes.arg.FileName)
		}
		par := parsed.Config[0].Parsed
		for p := 0; p < len(par); p++ {
			compareDirectives(t, par[p], tes.config[p])
		}
	}
}

func TestParseDump(t *testing.T) {
	const fileName = "config/simple.conf"
	var catcherr, single, comments bool
	parsed, err := ParseFile(fileName, nil, catcherr, single, comments)
	if err != nil {
		t.Fatal(err)
	}
	var w io.Writer = ioutil.Discard
	if testing.Verbose() {
		w = os.Stdout
	}
	parsed.Dump(w)
}

func TestRetag(t *testing.T) {
	t.Skip("No file config/tags.conf")
	const fileName = "config/tags.conf"
	var catcherr, single, comments bool
	Debugging = testing.Verbose()
	parsed, err := ParseFile(fileName, nil, catcherr, single, comments)
	if err != nil {
		t.Fatal(err)
	}
	t0 := tagCount(parsed)
	if t0 == 0 {
		t.Fatal("no tags found in original file")
	}
	t.Logf("original has %d tags\n", t0)

	// replicate marshaling payload to another process/server
	b, _ := json.Marshal(parsed)
	dupe := new(Payload)
	if err := json.Unmarshal(b, dupe); err != nil {
		t.Fatal(err)
	}

	t1 := tagCount(dupe)
	if t1 > 0 {
		t.Fatalf("expected no tags in dupe but found: %d", t1)
	}

	dupe.Retag()
	t1 = tagCount(dupe)
	if t1 != t0 {
		t.Fatalf("expected %d tags but got %d", t0, t1)
	}
}

func countTags(dirs []*Directive) int {
	count := 0
	for _, dir := range dirs {
		if dir.tag != "" {
			count++
		}
		count += countTags(dir.Block)
	}
	return count
}

func tagCount(p *Payload) int {
	count := 0
	for _, conf := range p.Config {
		count += countTags(conf.Parsed)
	}
	return count
}

func compareDirectives(t *testing.T, gen, config *Directive) {
	t.Helper()
	if gen.Directive != config.Directive {
		t.Errorf("directives want:%q got:%q\n", config.Directive, gen.Directive)
	}
	// loop over and compare
	if len(gen.Args) == len(config.Args) {
		for i, arg := range config.Args {
			if gen.Args[i] != arg {
				t.Errorf("(directive: %s) arg #%d want:%q got:%q\n", gen.Directive, i+1, arg, gen.Args[i])
			}
		}
	} else {
		t.Errorf("mismatched arg count - want(%d):%v got(%d):%v\n",
			len(config.Args), config.Args, len(gen.Args), gen.Args)
	}
	if gen.Line != config.Line {
		t.Errorf("(%s) line want:%d got:%d\n", config.Directive, config.Line, gen.Line)
	}
	if gen.File != config.File {
		t.Errorf("File want:%q got:%q\n", config.File, gen.File)
	}
	if gen.Comment != config.Comment {
		t.Errorf("Comment \nwant:%q\n got:%q\n", config.Comment, gen.Comment)
	}

	if len(gen.Block) == len(config.Block) {
		for i := 0; i < len(gen.Block); i++ {
			compareDirectives(t, gen.Block[i], config.Block[i])
		}
	} else {
		t.Errorf("(%s/%d) Mismatched blocks want:%d got:%d\nwant: %v\n got: %v\n",
			config.Directive, config.Line, len(config.Directive), len(gen.Directive), config.Directive, gen.Directive)
	}
}
