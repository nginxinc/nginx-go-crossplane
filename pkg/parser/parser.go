package parser

import (
	"fmt"
	"log"
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
	Comsume     bool
	checkCtx    bool
	checkArgs   bool
}

// ParsingError -
type ParsingError string

type Payload struct {
	Status string
	Errors []ParseErrors
	Config []Config
}

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
// Parse -
func Parse(a ParseArgs) Payload {

	includes := map[string][3]string{
		a.FileName: {},
	}
	q := Payload{
		Status: "ok",
		Errors: []ParseErrors{},
		Config: []Config{},
	}
	for f, r := range includes {
		token := lex(f)
		p := Config{
			File:   f,
			Status: "ok",
			Errors: []ParseErrors{},
			Parsed: []Block{},
		}
		// data to be changed to token
		p.Parsed, _ = parse(p, q, token, a, r, false)
		q.Config = append(q.Config, p)
	}
	if a.Combine {
		return q //combineParsedConfigs(p)
	}
	fmt.Println(q)
	return q

}

func parse(parsed Config, pay Payload, parsing []LexicalItem, a ParseArgs, ctx [3]string, consume bool) ([]Block, int) {
	o := []Block{}
	p := 0
	for ; p < len(parsing); p++ {
		b := Block{
			Directive: "",
			Line:      0,
			Args:      []string{},
			Includes:  []int{},
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
				_, i := parse(parsed, pay, parsing[p:], a, ctx, true)
				p += i
			}
			continue
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
		if a.Comments {
			q := []byte{'#'}

			if q[0] == parsing[p].item[0] {
				if a.Comments {
					b = Block{
						Directive: "",
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
			continue
		}
		// args for directives
		args := []string{}
		p++
		for ; parsing[p].item != ";" && parsing[p].item != "{" && parsing[p].item != "}"; p++ {
			args = append(args, parsing[p].item)
		}
		b.Args = args

		if len(a.Ignore) > 0 {
			for _, k := range a.Ignore {
				if k == parsing[p].item {
					_, i := parse(parsed, pay, parsing[p:], a, ctx, true)
					p += i
				}
			}
			continue
		}
		stmt := analyzer.Statement{
			Directive: b.Directive,
			Args:      b.Args,
			Line:      b.Line,
		}
		e := analyzer.Analyze(parsed.File, stmt, ";", ctx, a.Strict, a.checkCtx, a.checkArgs)
		if e != nil {
			if a.CatchErrors {
				handle_errors(parsed, pay, e, parsing[p].lineNum)
				if strings.HasSuffix(e.Error(), "is not terminated by \";\"") {
					if parsing[p].item != "}" {
						parse(parsed, pay, parsing[p:], a, ctx, true)
					} else {
						break
					}
				}
				continue

			} else {
				log.Fatal(e)
			}
		}
		// try analysing the directives
		if parsing[p].item == "{" {
			stmt := analyzer.Statement{
				Directive: b.Directive,
				Args:      b.Args,
				Line:      b.Line,
			}
			inner := analyzer.EnterBlockCTX(stmt, ctx)
			l := 0
			b.Block, l = parse(parsed, pay, parsing[p+1:], a, inner, false)
			p += l
		}
		o = append(o, b)

	}
	return o, p
}

func handle_errors(parsed Config, pay Payload, e error, line int) {
	file := parsed.File
	err := e.Error()

	parseerr := ParseErrors{
		Error: err,
		Line:  line,
		File:  "",
	}
	payloaderr := ParseErrors{
		Error: err,
		Line:  line,
		File:  file,
	}

	parsed.Status = "failed"
	parsed.Errors = append(parsed.Errors, parseerr)

	pay.Status = "failed"
	pay.Errors = append(pay.Errors, payloaderr)
}
