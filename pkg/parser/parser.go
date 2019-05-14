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

// ParseArgs -
type ParseArgs struct {
	FileName    string
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
	//Includes  []int
	Block   []Block
	File    string
	Comment string
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
func Parse(a ParseArgs) (Payload, error) {
	var e error
	included = []string{}
	includes = map[string][3]string{}
	includes[a.FileName] = [3]string{}

	payload = Payload{
		Status: "ok",
		Errors: []ParseError{},
		Config: []Config{},
		File:   a.FileName,
	}
	for f, r := range includes {
		//fmt.Println("file name : ", f)
		p := Config{
			File:   f,
			Status: "ok",
			Errors: []ParseError{},
			Parsed: []Block{},
		}
		contents, err := ioutil.ReadFile(f)
		if err != nil {
			return Payload{}, err
		}
		token, err := lexer.LexScanner(string(contents))
		if err != nil {
			return Payload{}, err
		}
		p.Parsed, _, e = parse(p, q, token, a, r, false)
		if e != nil {
			log.Println("error parsing")
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
				File:      args.FileName,
				Args:      []string{},
			}
		} else {
			block = Block{
				Directive: directive,
				Line:      token.LineNum,
				Args:      []string{},
			}
		}
		//fmt.Println("im directive : ", block.Directive)
		if string(parsing[p+1].Item) == "{" {
			stmt := analyzer.Statement{
				Directive: block.Directive,
				Args:      block.Args,
				Line:      block.Line,
			}
			inner := analyzer.EnterBlockCTX(stmt, ctx)
			l := 0
			block.Block, l, e = parse(parsed, pay, parsing[p+2:], args, inner, false)
			if e != nil {
				return o, p + l, e
			}
			p += l
			o = append(o, block)
			continue
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
		a := block.Args
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
		//fmt.Println("wbnfe : ", block.Directive)
		if stmt.Directive != "" && stmt.Directive != "if" && args.Strict {
			e := analyzer.Analyze(parsed.File, stmt, ";", ctx, args.Strict, args.CheckCtx, args.CheckArgs)
			if e != nil {
				if args.CatchErrors {
					handleErrors(parsed, pay, e, parsing[p].LineNum)
					if strings.HasSuffix(e.Error(), "is not terminated by \";\"") {
						if parsing[p].Item != "}" {
							parse(parsed, pay, parsing[p:], args, ctx, true)
						} else {
							continue
						}
					}
					continue

				} else {
					return o, e
				}
			}
		}
		if args.Single && block.Directive == "include" {
			configDir := filepath.Dir(args.FileName)
			pattern := a[0]
			var fnames []string
			var err error

			if !filepath.IsAbs(pattern) {
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
				b, e := canRead(pattern, args, parsing, token.LineNum)
				if e != nil {
					log.Fatal(e)
				}
				if b {

					fnames = []string{pattern}
				}
			}

			for _, fname := range fnames {
				if checkIncluded(fname, included) {
					//fmt.Println(fname)
					included = append(included, fname)
					includes[fname] = ctx

				}
			}
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
			return false
		}
	}
	return true
}

func canRead(pattern string, a ParseArgs, parsed Config, lineNumber int) (bool, error) {
	f, err := os.Open(pattern)
	if err != nil {
		if a.CatchErrors {
			handleErrors(parsed, err, lineNumber)
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
	var performIncludes func(b []Block) Block
	performIncludes = func(b []Block) Block {
		for _, block := range b {
			if len(block.Block) > 0 {
				a := performIncludes(block.Block)
				block.Block = append(block.Block, a)
			}
			if block.Directive == "include" {
				for _, f := range block.Args {
					fmt.Println(f)
					files, err := filepath.Glob(f)
					if err != nil {
						continue
					}
					for _, file := range files {
						config := findFile(file, oldConfig)
						g := performIncludes(config)
						for _, blo := range g.Block {
							return blo

						}
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
