package parser

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

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
var included = []string{}
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
	q := Payload{
		Status: "ok",
		Errors: []ParseError{},
		Config: []Config{},
		File:   a.FileName,
	}
	for f, r := range includes {
		p := Config{
			File:   f,
			Status: "ok",
			Errors: []ParseError{},
			Parsed: []Block{},
		}

		token := []LexicalItem{} // LexScanner(string(s))

		p.Parsed, _, e = parse(p, q, token, a, r, false)
		if e != nil {
			return q, e
		}
		q.Config = append(q.Config, p)
	}
	if a.Combine {
		return combineParsedConfigs(q)
	}

	return q, nil

}

func parse(parsed Config, pay Payload, parsing []LexicalItem, args ParseArgs, ctx [3]string, consume bool) ([]Block, int, error) {
	o := []Block{}
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
		if parsing[p].item == "}" {
			p++
			break
		}

		if consume {
			if parsing[p].item == "}" {
				_, i, e := parse(parsed, pay, parsing[p:], args, ctx, true)
				if e != nil {
					return o, p + i, e
				}
				p += i
			}
			continue
		}
		directive := parsing[p].item
		if args.Combine {
			block = Block{
				Directive: directive,
				Line:      parsing[p].lineNum,
				File:      args.FileName,
				Args:      []string{},
			}
		} else {
			block = Block{
				Directive: directive,
				Line:      parsing[p].lineNum,
				Args:      []string{},
			}
		}
		// comments in file
		if args.Comments {
			q := []byte{'#'}

			if q[0] == parsing[p].item[0] {

				block = Block{
					Directive: "",
					Comment:   parsing[p].item[1:],
					Args:      []string{},
					Block:     []Block{},
					File:      "",
					Line:      parsing[p].lineNum,
				}

			}

		}
		// args for directives
		a := []string{}
		p++
		for ; parsing[p].item != ";" && parsing[p].item != "{" && parsing[p].item != "}"; p++ {
			a = append(a, parsing[p].item)
		}
		block.Args = a

		if len(args.Ignore) > 0 {
			for _, k := range args.Ignore {
				if k == parsing[p].item {
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
					handleErrors(parsed, pay, e, parsing[p].lineNum)
					if strings.HasSuffix(e.Error(), "is not terminated by \";\"") {
						if parsing[p].item != "}" {
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
		// try analysing the directives
		if parsing[p].item == "{" {
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

		if args.Single && block.Directive == "include" {
			configDir := filepath.Dir(args.FileName)
			pattern := block.Args[0]
			fnames := []string{}
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

func canRead(p int, pattern string, a ParseArgs, parsed Config, pay Payload, parsing []LexicalItem) (bool, error) {
	f, err := os.Open(pattern)
	if err != nil {
		if a.CatchErrors {
			handleErrors(parsed, pay, err, parsing[p].lineNum)
		} else {
			return false, err
		}
	}
	f.Close()
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
