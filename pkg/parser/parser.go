package parser

import (
	"github.com/nginxinc/crossplane-go/pkg/analyzer"
)

// LexicalItem -
type LexicalItem struct {
	item    string
	lineNum int
}

// ParseArgs -
type ParseArgs struct {
	FileName string
	//onerror
	CatchErrors bool
	Ignore      []string
	Single      bool
	Comments    bool
	Strict      bool
	Combine     bool
	// checkCtx    bool
	// checkArgs bool
}

// ParsingError -
type ParsingError string

// Config -
type Config struct {
	File   string
	Status string
	Errors []ParseErrors
	Parsed []Block
}

// Block -
type Block struct {
	Directive string
	Line      int
	Args      []string
	Includes  []int
	Block     []Block
	File      string
	Comment   string
}

//ParseErrors -
type ParseErrors struct {
	File  string
	Line  int
	Error string
}

/*
   Parses an nginx config file and returns json payload

   :param filename: string contianing the name of the config file to parse
   :param catch_errors: bool; if False, parse stops after first error
   :param ignore: list or slice of directives to exclude from the payload
   :param combine: bool; if True, use includes to create a single config obj
   :param single: bool; if True, including from other files doesn't happen
   :param comments: bool; if True, including comments to json payload
   :param strict: bool; if True, unrecognized directives raise errors
   :param check_ctx: bool; if True, runs context analysis on directives
   :param check_args: bool; if True, runs arg count analysis on directives
   :returns: a payload that describes the parsed nginx config
*/

func parse(a ParseArgs) Config {
	data := []LexicalItem{
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
	}
	includes := map[string][3]string{
		a.FileName: {},
	}
	p := Config{
		File:   "",
		Status: "ok",
		Errors: []ParseErrors{},
		Parsed: []Block{},
	}
	for f, r := range includes {
		//token := lex(f)
		p := Config{
			File:   f,
			Status: "ok",
			Errors: []ParseErrors{},
			Parsed: []Block{},
		}
		// data to be changed to token
		p.Parsed, _ = Parsing(data, a, r)
	}
	if a.Combine {
		return p //combineParsedConfigs(p)
	}
	return p

}

// Parsing -
func Parsing(parsing []LexicalItem, a ParseArgs, ctx [3]string) ([]Block, int) {
	o := []Block{}
	p := 0
	for ; p < len(parsing); p++ {
		b := Block{}
		if parsing[p].item == "}" {
			p++
			break
		}
		directive := parsing[p].item
		if a.Combine {
			b = Block{
				Directive: directive,
				Line:      parsing[p].lineNum,
				File:      a.FileName,
				Args:      []string{},
			}
		} else {
			b = Block{
				Directive: directive,
				Line:      parsing[p].lineNum,
				Args:      []string{},
			}
		}
		// comments in file
		q := []byte{'#'}

		if q[0] == parsing[p].item[0] {
			if a.Comments {
				b = Block{
					Directive: "#",
					Comment:   string(parsing[p].item[1:]),
					Args:      []string{},
					Block:     []Block{},
					File:      "",
					Line:      parsing[p].lineNum,
					Includes:  []int{},
				}
			}
			continue
		}
		// args for directives
		args := []string{}
		p++
		for ; parsing[p].item != ";" && parsing[p].item != "{" && parsing[p].item != "}"; p++ {
			args = append(args, parsing[p].item)
		}
		b.Args = args

		if parsing[p].item == "{" {
			stmt := analyzer.Statement{
				Directive: b.Directive,
				Args:      b.Args,
				Line:      b.Line,
			}
			inner := analyzer.EnterBlockCTX(stmt, ctx)
			l := 0
			b.Block, l = Parsing(parsing[p+1:], a, inner)
			p += l
		}
		o = append(o, b)

	}
	return o, p
}
