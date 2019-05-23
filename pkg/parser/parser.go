package parser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nginxinc/crossplane-go/pkg/analyzer"
	"github.com/nginxinc/crossplane-go/pkg/lexer"
)

//LexicalItem -
type LexicalItem struct {
	item    string
	lineNum int
}
//type LexicalItem lexer.LexicalItem
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
	Consume     bool
	checkCtx    bool
	checkArgs   bool
}

// ParsingError -
type ParsingError error

// Payload -
type Payload struct {
	Status string
	Errors []ParseError
	File   string
	Config []Config
}

// Config -
type Config struct {
	File   string
	Status string
	Errors []ParseError
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

// ParseError -
type ParseError struct {
	File  string
	Line  int
	Error ParsingError
}

// list of conf files to be parsed
var included []string
var includes = map[string][3]string{}

// Parse - Parses an nginx config file and returns json payload
//   :param filename: string containing the name of the config file to parse
//   :param catch_errors: bool; if False, parse stops after first error
//   :param ignore: list or slice of directives to exclude from the payload
//   :param combine: bool; if True, use includes to create a single config obj
//   :param single: bool; if True, including from other files doesn't happen
//   :param comments: bool; if True, including comments to json payload
//   :param strict: bool; if True, unrecognized directives raise errors
//   :param check_ctx: bool; if True, runs context analysis on directives
//   :param check_args: bool; if True, runs arg count analysis on directives
//   :returns: a payload that describes the parsed nginx config
func Parse(file string, catcherr bool, ignore []string, single bool, comment bool, strict bool,
	combine bool, consume bool, checkctx bool, checkargs bool) (Payload, error) {
	var e error
	a := ParseArgs{
		FileName:    file,
		CatchErrors: catcherr,
		Ignore:      ignore,
		Single:      single,
		Comments:    comment,
		Strict:      strict,
		Combine:     combine,
		Consume:     consume,
		checkCtx:    checkctx,
		checkArgs:   checkargs,
	}
	includes[a.FileName] = [3]string{}

	p := Payload{
		Status: "ok",
		Errors: []ParseError{},
		Config: []Config{},
		File:   a.FileName,
	}
	for f, r := range includes {

		c := Config{
			File:   f,
			Status: "ok",
			Errors: []ParseError{},
			Parsed: []Block{},
		}

		re, err := ioutil.ReadFile(f)
		if err != nil{
			fmt.Println(err)
			return p, nil
		}
		// we should probably pass a file?
		tokens, _ := lexer.LexScanner(string(re))
		c.Parsed, _, e = parse(c, p, tokens, a, r, false)
		if e != nil {
			return p, e
		}
		p.Config = append(p.Config, c)
	}
	if a.Combine {
		return combineParsedConfigs(p)
	}

	return p, nil

}

func parse(parsed Config, pay Payload, parsing []lexer.LexicalItem, args ParseArgs, ctx [3]string, consume bool) ([]Block, int, error) {
	var o []Block
	var e error
	p := 0
	for ; p < len(parsing); p++ {
		block := Block{
			Directive: "",
			Line:      0,
			Args:      []string{},
			File:      "",
			Comment:   "",
			Block:     []Block{},
		}
		if parsing[p].Item == "}" {
			break
		}

		if consume {
			if parsing[p].Item == "{" {
				_, i, e := parse(parsed, pay, parsing[p:], args, ctx, true)
				if e != nil {
					return o, p + i, e
				}
				p += i
			}
			continue
		}
		directive := parsing[p].Item
		if args.Combine {
			block = Block{
				Directive: directive,
				Line:      parsing[p].LineNum,
				File:      args.FileName,
				Args:      []string{},
			}
		} else {
			block = Block{
				Directive: directive,
				Line:      parsing[p].LineNum,
				Args:      []string{},
			}
		}
		// comments in file
		if strings.HasPrefix(parsing[p].Item, "#") {
			if args.Comments {
				block = Block{
					Directive: "#",
					Comment:   parsing[p].Item[1:],
					Args:      []string{},
					Block:     []Block{},
					File:      "",
					Line:      parsing[p].LineNum,
				}
				o = append(o, block)

			}
			continue

		}
		// args for directives
		var a []string

		// this is an ugly version of next() in python
		p++
		if p >= len(parsing){
			continue
		}

		for ; parsing[p].Item != ";" && parsing[p].Item != "{" && parsing[p].Item != "}"; p++{
			a = append(a, parsing[p].Item)
		}
		block.Args = a

		if len(args.Ignore) > 0 {
			for _, k := range args.Ignore {
				if k == parsing[p].Item {
					_, i, e := parse(parsed, pay, parsing[p:], args, ctx, true)
					if e != nil {
						return o, p + i, e
					}
					p += i
				}
			}
			continue
		}
		if block.Directive == "if" {
			block.Args = removeBrackets(block.Args)
		}
		stmt := analyzer.Statement{
			Directive: block.Directive,
			Args:      block.Args,
			Line:      block.Line,
		}

		if stmt.Directive != "" && stmt.Directive != "if" {
			e := analyzer.Analyze(parsed.File, stmt, ";", ctx, args.Strict, args.checkCtx, args.checkArgs)
			if e != nil {
				if args.CatchErrors {
					handleErrors(parsed, pay, e, parsing[p].LineNum)
					if strings.HasSuffix(e.Error(), "is not terminated by \";\"") {
						if parsing[p].Item != "}" {
							parse(parsed, pay, parsing[p:], args, ctx, true)
						} else {
							break
						}
					}
					continue

				} else {
					return o, p, e
				}
			}
		}


		if args.Single && block.Directive == "include" {
			configDir := filepath.Dir(args.FileName)
			pattern := block.Args[0]
			var fnames []string
			var err error
			if filepath.IsAbs(pattern) {
				pattern = filepath.Join(configDir, pattern)
			}

			hasMagic := func(pat string) bool {
				magic := []byte{'*', '?', ']', '[', '{', '}', '(', ')'}
				for _, m := range magic {
					for _, p := range pat {
						if m == byte(p) {
							return true
						}
					}
				}
				return false
			}

			if hasMagic(pattern) {
				fnames, err = filepath.Glob(pattern)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				b, e := canRead(p, pattern, args, parsed, pay, parsing)
				if e != nil {
					log.Fatal(e)
				}
				if b {
					fnames = []string{pattern}
				}
			}

			for _, fname := range fnames {
				if !checkIncluded(fname, included) {
					included = append(included, fname)
					includes[fname] = ctx
				}
			}
		}
		// try analysing the directives
		if parsing[p].Item == "{" {
			stmt := analyzer.Statement{
				Directive: block.Directive,
				Args:      block.Args,
				Line:      block.Line,
			}
			inner := analyzer.EnterBlockCTX(stmt, ctx)
			l := 0

			block.Block, l, e = parse(parsed, pay, parsing[p+1:], args, inner, false)
			if e != nil {
				return o, p + l, e
			}
			p += l
		}

		o = append(o, block)

	}
	return o, p, nil
}

func removeBrackets(s []string) []string {
	if s[0] == "(" && s[len(s)-1] == ")" {
		s = s[1 : len(s)-2]
	}
	return s
}

func checkIncluded(fname string, included []string) bool {
	for _, f := range included {
		if f == fname {
			return false
		}
	}
	return true
}

func canRead(p int, pattern string, a ParseArgs, parsed Config, pay Payload, parsing []lexer.LexicalItem) (bool, error) {
	f, err := os.Open(pattern)
	if err != nil {
		if a.CatchErrors {
			handleErrors(parsed, pay, err, parsing[p].LineNum)
		} else {
			return false, err
		}
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("error closing the file")
		}
	}()
	return true, nil
}

func handleErrors(parsed Config, pay Payload, e error, line int) {
	file := parsed.File

	parseErr := ParseError{
		Error: e,
		Line:  line,
		File:  "",
	}
	payloadErr := ParseError{
		Error: e,
		Line:  line,
		File:  file,
	}

	parsed.Status = "failed"
	parsed.Errors = append(parsed.Errors, parseErr)

	pay.Status = "failed"
	pay.Errors = append(pay.Errors, payloadErr)
}

func combineParsedConfigs(p Payload) (Payload, error) {
	if p.Config == nil {
		return Payload{}, errors.New("Input pyload config is nil")
	}
	oldConfig := p.Config
	var performIncludes func(b []Block) Block
	performIncludes = func(b []Block) Block {
		for _, block := range b {
			if len(block.Block) > 0 {
				a := performIncludes(block.Block)
				block.Block = append(block.Block, a)
			}
			if block.Directive == "include" {
				for _, f := range block.Args {
					config := findFile(f, oldConfig)
					g := performIncludes(config)
					for _, blo := range g.Block {
						return blo

					}
				}
			} else {
				return block
			}
		}
		return Block{}
	}

	combineConfig := Config{
		File:   oldConfig[0].File,
		Status: "ok",
		Errors: []ParseError{},
		Parsed: []Block{},
	}

	for _, config := range oldConfig {
		for _, e := range config.Errors {
			combineConfig.Errors = append(combineConfig.Errors, e)
		}
		if config.Status != "ok" {
			combineConfig.Status = "failed"
		}
	}
	firstConfig := oldConfig[0].Parsed
	combineConfig.Parsed = append(combineConfig.Parsed, performIncludes(firstConfig))

	combinePayload := Payload{
		Status: p.Status,
		Errors: p.Errors,
		File:   p.File,
		Config: []Config{combineConfig},
	}
	return combinePayload, nil
}

func findFile(f string, config []Config) []Block {
	for _, i := range config {
		if i.File == f {
			return i.Parsed
		}
	}
	return []Block{}
}
