package parser

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nginxinc/crossplane-go/pkg/analyzer"
	"github.com/nginxinc/crossplane-go/pkg/lexer"
)

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
	CheckCtx    bool
	CheckArgs   bool
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
var includes map[string][3]string
var payload Payload

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
	included = []string{file}
	includes = map[string][3]string{}
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
		CheckCtx:    checkctx,
		CheckArgs:   checkargs,
	}
	includes[a.FileName] = [3]string{}
	payload = Payload{
		Status: "ok",
		Errors: []ParseError{},
		Config: []Config{},
		File:   a.FileName,
	}
	for i := 0; i < len(included); i++ { //f, r := range includes {
		f := included[i]
		c := Config{
			File:   f,
			Status: "ok",
			Errors: []ParseError{},
			Parsed: []Block{},
		}

		re, err := ioutil.ReadFile(f)
		if err != nil {
			if a.CatchErrors {
				handleErrors(c, err, 0)
				continue
			} else {
				return payload, nil
			}
		}
		// we should probably pass a file?
		tokens := lexer.LexScanner(string(re))
		c.Parsed, e = parse(c, tokens, a, includes[f], false)
		if e != nil {
			return payload, e
		}
		payload.Config = append(payload.Config, c)

	}

	if a.Combine {
		return combineParsedConfigs(payload)
	}
	return payload, nil

}

func parse(parsing Config, tokens <-chan lexer.LexicalItem, args ParseArgs, ctx [3]string, consume bool) ([]Block, error) {
	var o []Block
	var e error
	for token := range tokens {
		block := Block{
			Directive: "",
			Line:      0,
			Args:      []string{},
			File:      "",
			Comment:   "",
			Block:     []Block{},
		}
		if token.Item == "}" {
			break
		}
		if consume {
			if token.Item == "{" {
				_, _ = parse(parsing, tokens, args, ctx, true)
			}
			continue
		}
		directive := token.Item
		if args.Combine {
			block = Block{
				Directive: directive,
				Line:      token.LineNum,
				File:      parsing.File,
				Args:      []string{},
			}
		} else {
			block = Block{
				Directive: directive,
				Line:      token.LineNum,
				Args:      []string{},
			}
		}
		// comments in file
		if strings.HasPrefix(directive, "#") {
			if args.Comments {
				block = Block{
					Directive: "#",
					Comment:   token.Item[1:],
					Args:      []string{},
					Block:     []Block{},
					File:      "",
					Line:      token.LineNum,
				}
				o = append(o, block)
			}
			continue

		}
		// args for directives
		token := <-tokens
		for token.Item != ";" && token.Item != "{" && token.Item != "}" {
			block.Args = append(block.Args, token.Item)
			token = <-tokens
		}

		if len(args.Ignore) > 0 {
			for _, k := range args.Ignore {
				if k == token.Item {
					o, e = parse(parsing, tokens, args, ctx, true)
					if e != nil {
						return o, e
					}
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
			e := analyzer.Analyze(parsing.File, stmt, ";", ctx, args.Strict, args.CheckCtx, args.CheckArgs)
			if e != nil {
				if args.CatchErrors {
					handleErrors(parsing, e, token.LineNum)
					if strings.HasSuffix(e.Error(), "is not terminated by \";\"") {
						if token.Item != "}" {
							parse(parsing, tokens, args, ctx, true)
						} else {
							break
						}
					}
					continue

				} else {
					return o, e
				}
			}
		}
		// check if file can be opened
		if !args.Single && block.Directive == "include" {
			a := block.Args
			configDir := filepath.Dir(args.FileName)
			pattern := a[0]
			var fnames []string
			var err error

			pattern = filepath.Join(configDir, pattern)

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
					if args.CatchErrors {
						handleErrors(parsing, e, token.LineNum)
						continue
					} else {
						log.Fatal(err)
					}
				}
			} else {
				b, e := canRead(pattern, args, parsing, token.LineNum)
				if e != nil {
					if args.CatchErrors {
						handleErrors(parsing, e, token.LineNum)
						continue
					} else {
						log.Fatal(e)
					}
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
			block.Args = fnames
		}
		// try analysing the directives
		if token.Item == "{" {
			stmt := analyzer.Statement{
				Directive: block.Directive,
				Args:      block.Args,
				Line:      block.Line,
			}
			inner := analyzer.EnterBlockCTX(stmt, ctx)
			block.Block, e = parse(parsing, tokens, args, inner, false)
			if e != nil {
				return o, e
			}
		}

		o = append(o, block)
	}
	return o, nil
}

func removeBrackets(s []string) []string {
	if strings.HasPrefix(s[0], "(") && strings.HasSuffix(s[len(s)-1], ")") {
		s[0] = strings.TrimPrefix(s[0], "(")
		s[len(s)-1] = strings.TrimSuffix(s[len(s)-1], ")")
		if s[len(s)-1] == "" {
			s = s[:len(s)-1]
		}
	}
	return s
}

func checkIncluded(fname string, included []string) bool {
	for _, f := range included {
		if f == fname {
			return true
		}
	}
	return false
}

func canRead(pattern string, a ParseArgs, parsed Config, lineNumber int) (bool, error) {
	f, err := os.Open(pattern)
	if err != nil {
		if a.CatchErrors {
			handleErrors(parsed, err, lineNumber)
			return false, nil
		}
		return false, err

	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("error closing the file")
		}
	}()
	return true, nil
}

func handleErrors(parsed Config, e error, line int) {
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

	payload.Status = "failed"
	payload.Errors = append(payload.Errors, payloadErr)
}

func combineParsedConfigs(p Payload) (Payload, error) {
	if p.Config == nil {
		return Payload{}, errors.New("Input pyload config is nil")
	}
	oldConfig := p.Config
	var performIncludes func(b []Block) []Block
	performIncludes = func(b []Block) []Block {
		var returnBlock []Block
		for _, block := range b {
			if len(block.Block) > 0 {
				block.Block = performIncludes(block.Block)
			}
			if block.Directive == "include" {
				for _, f := range block.Args {
					config := findFile(f, oldConfig)
					g := performIncludes(config)
					for _, bl := range g {
						returnBlock = append(returnBlock, bl)
					}
				}
				continue
			}
			returnBlock = append(returnBlock, block)
		}
		return returnBlock
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
	combineConfig.Parsed = performIncludes(firstConfig)

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
