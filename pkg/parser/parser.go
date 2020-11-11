package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/lexer"
)

type Opener func(string) (io.Reader, error)
type Globber func(string) ([]string, error)

type ParseArgs struct {
	FileName    string
	PrefixPath  string // The directory where the config was loaded
	ConfigDir   string // The absolute path
	Ignore      []string
	Single      bool
	CatchErrors bool
	Comments    bool
	StripQuotes bool
	Open        Opener
	Glob        Globber
}

func (p *ParseArgs) init() {
	if p.Open == nil {
		p.Open = osOpen()
	}
	if p.Glob == nil {
		p.Glob = osGlob()
	}
}

// Payload represents a parsed nginx config(s)
// It is the parent struct for parsed data
type Payload struct {
	Status string       `json:"status"`
	Errors []ParseError `json:"errors"`
	Config []*Config    `json:"config"`
}

// because "parser" is overloaded already
type runner struct {
	p        *Payload
	dir      string
	included map[string]int
	includes []string
}

// ConfigError represents a config error
type ConfigError struct {
	Error string `json:"error"`
	Line  int    `json:"line"`
}

// Config represents a config file
type Config struct {
	File   string        `json:"file"`
	Status string        `json:"status"`
	Errors []ConfigError `json:"errors"`
	Parsed []*Directive  `json:"parsed"`
}

// Directive -
type Directive struct {
	File      string       `json:"file,omitempty"`
	Directive string       `json:"directive"`
	Line      int          `json:"line"`
	Args      []string     `json:"args"`
	Includes  []int        `json:"includes,omitempty"`
	Block     []*Directive `json:"block,omitempty"`
	Comment   string       `json:"comment,omitempty"`
}

// IsBlock returns true if this is a block directive.
func (d *Directive) IsBlock() bool {
	return d.Block != nil
}

// IsComment returns true when the directive is a comment directive
func (d *Directive) IsComment() bool {
	return d.Directive == "#"
}

// IsIf returns true when the directive is an if directive
func (d *Directive) IsIf() bool {
	return d.Directive == "if"
}

// String makes this a Stringer
func (d *Directive) String() string {
	return fmt.Sprintf("%s:%s", d.Directive, strings.Join(d.Args, ","))
}

// ParseError provides context for a parse error
type ParseError struct {
	File   string `json:"file,omitempty"` // python parity
	Fail   string `json:"error"`
	Line   int    `json:"line"`
	Column int    `json:"-"` // disable for python validation
}

// Error makes this a proper Error, eh?
func (p ParseError) Error() string {
	return fmt.Sprintf("file:%s line:%d, col:%d error:%v", p.File, p.Line, p.Column, p.Fail)
}

func pythonNotFound(file string) error {
	return fmt.Errorf("[Errno 2] No such file or directory: '%s'", file)
}

// LoadPayload creates a Payload from a json file
func LoadPayload(file string) (*Payload, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var payload Payload
	return &payload, json.NewDecoder(f).Decode(&payload)
}

// Parse - ingests an nginx config file and returns json payload
func Parse(args ParseArgs) (*Payload, error) {
	(&args).init()

	// agent expects absolute paths everywhere, so if the
	// file to be parsed is an absolute path, then signal
	// to save all paths as absolute
	if filepath.IsAbs(args.FileName) && args.ConfigDir == "" {
		args.ConfigDir = filepath.Dir(args.FileName)
	}

	payload := &Payload{
		Status: "ok",
		Errors: []ParseError{},
	}

	run := &runner{
		p:        payload,
		dir:      filepath.Dir(args.FileName),
		included: map[string]int{args.FileName: 0},
		includes: []string{args.FileName},
	}

	for offset := 0; offset < len(run.includes); offset++ {
		args.FileName = run.includes[offset]
		conf, err := run.readFile(args)
		if err != nil {
			return nil, err
		}
		run.p.Config = append(run.p.Config, conf)
	}

	if len(payload.Errors) > 0 {
		payload.Status = "failed"
	}
	return payload, nil
}

func osOpen() Opener {
	return func(filename string) (io.Reader, error) {
		return os.Open(filename)
	}
}

func osGlob() Globber {
	return func(filename string) ([]string, error) {
		return filepath.Glob(filename)
	}
}

// parse a config file
func (run *runner) readFile(a ParseArgs) (*Config, error) {
	r, err := a.Open(a.FileName)
	if err != nil {
		return nil, pythonNotFound(a.FileName)
	}
	if c, ok := r.(io.Closer); ok {
		defer func() {
			c.Close()
		}()
	}
	return run.configReader(r, a)
}

// parse a config via io.Reader
func (run *runner) configReader(r io.Reader, a ParseArgs) (*Config, error) {
	c := &Config{
		Status: "ok",
		File:   a.FileName,
		Errors: []ConfigError{},
	}

	var err error
	tokens := lexer.LexScanReader(r, a.StripQuotes)
	c.Parsed, err = run.parse(c, tokens, a, false)
	if len(c.Errors) > 0 {
		c.Status = "failed"
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ParseString - Parses an nginx config string and returns payload
func ParseString(s string, args ParseArgs) (*Payload, error) {
	(&args).init()
	p := &Payload{Status: "ok", Errors: []ParseError{}}
	r := strings.NewReader(s)
	run := &runner{p: p}
	conf, err := run.configReader(r, args)
	if err != nil {
		return nil, err
	}
	p.Config = append(p.Config, conf)
	return p, nil
}

func (run *runner) parse(config *Config, tokens <-chan lexer.LexicalItem, args ParseArgs, consume bool) ([]*Directive, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	var out []*Directive
	var e error
	// TODO: verify/remove this if we don't return quotes
	isQuoted := false

	for token := range tokens {
		if token.Item == "}" {
			break
		}

		var commentsInArgs []string

		block := &Directive{
			Directive: strings.Trim(token.Item, `"'`),
			Line:      token.LineNum,
			Args:      []string{},
		}

		// closure over token, block, & config
		trap := func(err error) error {
			if !args.CatchErrors {
				return err
			}
			out = append(out, block)
			if err != nil {
				parseErr := ParseError{
					Fail: err.Error(),
					Line: token.LineNum,
					//Column: token.Column, // python doesn't do this!?
					File: args.FileName,
				}
				run.p.Errors = append(run.p.Errors, parseErr)
				confErr := ConfigError{
					Error: err.Error(),
					Line:  token.LineNum,
				}
				config.Errors = append(config.Errors, confErr)
			}
			return nil
		}

		if consume {
			if token.Item == "{" {
				if block.Block, e = run.parse(config, tokens, args, true); e != nil {
					return nil, e
				}
				out = append(out, block)
			}
			continue
		}

		if strings.HasPrefix(token.Item, "#") && !isQuoted {
			if args.Comments {
				block.Directive = "#"
				block.Comment = token.Item[1:]
				out = append(out, block)
			}
			continue
		}

		// args for directives
		token, ok := <-tokens
		if !ok {
			break
		}

		// consume directive args
		for !(isQuoted || among(token.Item, ";", "{", "}")) {
			// TODO: not sure if lexer is even returning these as tokens (thus no need for isQuoted)
			if token.Item == "\"" || token.Item == "'" {
				isQuoted = !isQuoted
			}
			if strings.HasPrefix(token.Item, "#") && !isQuoted {
				commentsInArgs = append(commentsInArgs, token.Item[1:])
			} else {
				// strip quotes for python compat
				// TODO: probably want to keep them for retaining formatting
				if args.StripQuotes {
					q := token.Item[0]
					if q == token.Item[len(token.Item)-1] && strings.ContainsAny(token.Item[:1], `"'`) {
						token.Item = strings.Trim(token.Item, string(token.Item[:1]))
					}
				}
				block.Args = append(block.Args, token.Item)
			}
			token, ok = <-tokens
			if !ok {
				break
			}
		}

		if block.Directive == "if" {
			block.Args = removeBrackets(block.Args)
		}

		if !args.Single && block.Directive == "include" {
			if len(block.Args) == 0 {
				return nil, errors.New("no parameter for include")
			}
			// skip excluded files
			if among(block.Args[0], args.Ignore...) {
				continue
			}
			// should not be nil for include directives (python compat)
			block.Includes = []int{}

			pattern := block.Args[0]
			if filepath.IsAbs(pattern) {
				pattern = strings.TrimPrefix(pattern, args.PrefixPath)
			} else if args.ConfigDir != "" {
				pattern = path.Join(args.ConfigDir, pattern)
			} else {
				pattern = path.Join(run.dir, pattern)
			}

			fnames, err := args.Glob(pattern)
			if err != nil {
				if err = trap(err); err != nil {
					return nil, err
				}
				continue
			}

			if len(fnames) == 0 {
				if err := trap(pythonNotFound(pattern)); err != nil {
					return nil, err
				}
				block.Includes = []int{}
				continue
			}
			// for python compatability until validated
			sort.Slice(fnames, func(i, j int) bool {
				return filepath.Base(fnames[i]) < filepath.Base(fnames[j])
			})

			// NOTE: a file might be included mutliple times
			for _, fname := range fnames {
				idx, ok := run.included[fname]
				if !ok {
					idx = len(run.included)
					run.included[fname] = idx
					run.includes = append(run.includes, fname)
				}
				block.Includes = append(block.Includes, idx)
			}
		}

		if token.Item == "{" && !isQuoted {
			blocks, err := run.parse(config, tokens, args, false)
			if err != nil {
				if err = trap(err); err != nil {
					return out, err
				}
				continue
			}
			if len(blocks) > 0 {
				block.Block = blocks
			}
			if block.Block == nil {
				//ensure Block is not nil to force braces being printed
				block.Block = []*Directive{}
			}
		}

		out = append(out, block)
		for _, comment := range commentsInArgs {
			out = append(out, &Directive{
				Directive: "#",
				Line:      block.Line,
				Args:      []string{},
				Comment:   comment,
				Includes:  []int{},
			})
		}
	}
	return out, nil
}

func removeBrackets(brackets []string) []string {
	clean := make([]string, 0, len(brackets))
	for _, s := range brackets {
		if s = strings.Trim(s, "()"); s != "" {
			clean = append(clean, s)
		}
	}
	return clean
}

// replaces blocks that include other configs by their respective blocks
// used to consolidate a multi-file config
func (p *Payload) blockIncludes(filename string, blocks []*Directive) []*Directive {
	var updated []*Directive
	for _, block := range blocks {
		block.File = filename
		if len(block.Includes) == 0 {
			block.File = filename
			block.Block = p.blockIncludes(filename, block.Block)
			updated = append(updated, block)
			continue
		}
		for _, include := range block.Includes {
			conf := p.Config[include]
			included := conf.Parsed
			updated = append(updated, p.blockIncludes(conf.File, included)...)
		}
	}
	return updated
}

// Unify converts a payload with multiple imports
// into a single config structure
func (p *Payload) Unify() (*Payload, error) {
	if p.Config == nil {
		return nil, errors.New("input payload config is nil")
	}

	filename := p.Config[0].File
	newConfig := &Config{
		File:   filename,
		Parsed: p.blockIncludes(filename, p.Config[0].Parsed),
		Status: p.Config[0].Status,
		Errors: p.Config[0].Errors,
	}

	singular := &Payload{
		Status: "ok",
		Config: []*Config{newConfig},
		Errors: []ParseError{},
	}

	for _, conf := range p.Config[1:] {
		for _, cerr := range conf.Errors {
			newConfig.Errors = append(newConfig.Errors, cerr)
			perr := ParseError{
				File: conf.File,
				Fail: cerr.Error,
			}
			singular.Errors = append(singular.Errors, perr)
		}
	}
	return singular, nil
}

func among(item string, list ...string) bool {
	for _, x := range list {
		if item == x {
			return true
		}
	}
	return false
}
