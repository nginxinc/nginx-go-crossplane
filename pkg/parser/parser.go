package parser

import (
	"fmt"
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

func parse(a ParseArgs) (Config, error) {
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
		{item: "server_name", lineNum: 8},
		{item: "default_server", lineNum: 8},
		{item: "location", lineNum: 9},
		{item: "/", lineNum: 9},
		{item: "{", lineNum: 9},
		{item: "return", lineNum: 10},
		{item: "200", lineNum: 10},
		{item: "foo bar baz", lineNum: 10},
		{item: "}", lineNum: 11},
		{item: "}", lineNum: 12},
		{item: "}", lineNum: 13},
	}
	includes := map[string][]string{
		a.FileName: []string{},
	}
	p := Config{
		File:   "",
		Status: "ok",
		Errors: []ParseErrors{},
		Parsed: []Block{},
	}
	for f, r := range includes {
		p.File = f
		v, _ := Parsing(data, a, r)
		p.Parsed = v
	}
	return p, nil
}

// Parsing -
func Parsing(parsing []LexicalItem, a ParseArgs, ctx []string) ([]Block, int) {
	newb := Block{}
	w := 0
	for i, p := range parsing {
		w = i
		b := Block{
			Directive: "",
			Comment:   "",
			Includes:  []int{},
			Line:      p.lineNum,
			File:      "",
			Args:      []string{},
			Block:     []Block{},
		}
		if a.Combine {
			b.File = a.FileName
		}

		if s := isDirective(p.item); s {
			b.Directive = p.item
		} else if s = checkifParent(p.item); s {
			b.Directive = p.item
		}
		q := []byte{'#'}

		if q[0] == p.item[0] {
			if a.Comments {
				b = Block{
					Directive: "#",
					Comment:   string(p.item[1:]),
					Args:      []string{},
					Block:     []Block{},
					File:      "",
					Line:      p.lineNum,
					Includes:  []int{},
				}
			}
		}

		args := []string{}
		count := 0
		if p.item != "{" && p.item != ";" && p.item != "}" && isDirective(p.item) {
			count++
			p = parsing[i+count]
			args = append(args, p.item)

		} else {
			continue
		}
		b.Args = args

		if p.item == "{" {
			v, _ := Parsing(parsing[i+1:], a, ctx)
			if isDirective(v[0].Directive) {
				b.Block = v
			}

		} else if p.item == "}" {
			// cna't do his part until analyze and parser code are on the same branch
			/*stm := statement{
				directive: b.Directive,
				args:      b.Args,
				line:      b.Line,
			}
			inner := analyzer.enterBlockCTX(stm, ctx)
			b.Block, _ = Parsing(parsing, a, inner)*/
			break
		}

		newb.Block = append(newb.Block, b)

	}
	fmt.Println("newb : ", newb.Block)
	b := []Block{newb}
	return b, w
}

func checkifParent(s string) bool {
	fmt.Println(s)
	if s == "http" || s == "server" || s == "location" || s == "events" {
		return true
	}
	return false
}

func isDirective(s string) bool {
	d := []string{
		"try_files",
		"return",
		"root",
		"listen",
		"error_log",
		"default_type",
		"access_log",
		"user",
		"worker_connections",
	}
	fmt.Println(s)
	for _, t := range d {
		if t == s {
			return true
		}
	}
	return false
}
