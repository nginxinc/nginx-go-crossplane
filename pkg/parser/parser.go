package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/lexer"
)

// ParseArgs holds all parameters required to parse a config file
type ParseArgs struct {
	FileName    string
	CatchErrors bool
	Ignore      []string
	Single      bool // TODO: does ignoring includes have value?
	Comments    bool
	Consume     bool
}

// Payload represents a parsed nginx config(s)
// It is the parent struct for parsed data
type Payload struct {
	Config []*Config    `json:"config,omitempty"`
	Errors []ParseError `json:"errors,omitempty"`
	File   string       `json:"file"`
	Dir    string       `json:"dir,omitempty"`
}

// Config represents a config file
type Config struct {
	File    string       `json:"file"`
	Errors  []ParseError `json:"errors,omitempty"`
	Parsed  []*Directive `json:"parsed,omitempty"`
	Context []int        `json:"context,omitempty"`
}

// Directive -
type Directive struct {
	tag       string
	Directive string       `json:"directive"`
	File      string       `json:"file,omitempty"`
	Comment   string       `json:"comment,omitempty"`
	Line      int          `json:"line"`
	Args      []string     `json:"args,omitempty"`
	Includes  []int        `json:"includes,omitempty"`
	Block     []*Directive `json:"block,omitempty"`
}

// IsComment returns true when the directive is a comment directive
func (d *Directive) IsComment() bool {
	return d.Directive == "#"
}

// IsIf returns true when the directive is an if directive
func (d *Directive) IsIf() bool {
	return d.Directive == "if"
}

// ParseError provides context for a parse error
type ParseError struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Fail   error  `json:"error"`
}

// Error makes this a proper Error, eh?
func (p ParseError) Error() string {
	return fmt.Sprintf("file:%s line:%d, col:%d error:%v", p.File, p.Line, p.Column, p.Fail)
}

// ParseFile - ingests an nginx config file and returns json payload
//   :param filename: string containing the name of the config file to parse
//   :param ignore: list or slice of directives to exclude from the payload
//   :param catch_errors: bool; if False, parse stops after first error
//   :param single: bool; if True, including from other files doesn't happen
//   :param comments: bool; if True, including comments to json payload
//   :returns: a payload that describes the parsed nginx config
func ParseFile(file string, ignore []string, catcherr, single, comment bool) (*Payload, error) {
	fpath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(fpath)
	if err := os.Chdir(dir); err != nil {
		return nil, err
	}
	fpath = filepath.Base(fpath)

	a := ParseArgs{
		FileName:    fpath,
		CatchErrors: catcherr,
		Ignore:      ignore,
		Single:      single,
		Comments:    comment,
	}

	payload := &Payload{
		File: fpath,
		Dir:  dir,
	}
	err = payload.readFile(a)

	if err2 := os.Chdir(cwd); err2 != nil && err == nil {
		err = err2
	}
	return payload, err

}

// parse a config file
func (p *Payload) readFile(a ParseArgs) error {
	f, err := os.Open(a.FileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return p.configReader(f, a)
}

// parse a config presented by the Reader
func (p *Payload) configReader(r io.Reader, a ParseArgs) error {
	c := &Config{
		File: a.FileName,
	}

	tokens := lexer.LexScanReader(r)

	// because all configs are in a slice, the first config
	// is effectively the root and needs special accomodations
	// TODO: consider having Payload.Parent blocks and configs are additions
	first := len(p.Config) == 0
	if first {
		p.Config = append(p.Config, c)
	}

	var err error
	c.Parsed, err = p.parse(*c, tokens, a)
	if err != nil {
		return err
	}

	if !first {
		p.Config = append(p.Config, c)
	}
	return nil
}

// ParseString - Parses an nginx config string and returns json payload
//   :param filename: string containing the filename associated with the config
//   :param config: string containing the nginx config contents to parse
//   :param catch_errors: bool; if False, parse stops after first error
//   :param ignore: list or slice of directives to exclude from the payload
//   :param single: bool; if True, including from other files doesn't happen
//   :param comments: bool; if True, exclude comments from json payload
//   :returns: a payload that describes the parsed nginx config
func ParseString(filename, config string, ignore []string, catcherr, single, comment bool) (*Payload, error) {
	a := ParseArgs{
		FileName:    filename,
		CatchErrors: catcherr,
		Ignore:      ignore,
		Single:      single,
		Comments:    comment,
	}

	p := &Payload{
		File: filename,
	}
	r := strings.NewReader(config)
	return p, p.configReader(r, a)
}

func (p *Payload) parse(config Config, tokens <-chan lexer.LexicalItem, args ParseArgs) ([]*Directive, error) {
	var out []*Directive
	var e error
	var tag string
	baseDir := filepath.Dir(config.File)
	for token := range tokens {
		if token.Item == "}" {
			break
		}
		block := &Directive{
			Directive: token.Item,
			Line:      token.LineNum,
			tag:       tag,
		}

		// NOTE: consume is only checked here -- to signal if we are inside a bracket import
		if args.Consume {
			if token.Item == "{" {
				// and our corresponding toggle to signal our next trip goes past
				args.Consume = false
				if block.Block, e = p.parse(config, tokens, args); e != nil {
					return nil, e
				}
				out = append(out, block)
			}
			continue
		}

		if strings.HasPrefix(token.Item, "#") {
			if !args.Comments {
				if strings.HasPrefix(token.Item, "#@") {
					// TODO: maybe directly to "tag"?
					snip := strings.TrimSpace(token.Item[2:])
					if snip != "" {
						debugf("setting tag to: %q\n", snip)
						tag = snip
						block.tag = tag
					}
				}
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

		isQuoted := false
		for !(isQuoted || among(token.Item, ";", "{", "}")) {
			if token.Item == "\"" || token.Item == "'" {
				isQuoted = !isQuoted
			}
			block.Args = append(block.Args, token.Item)
			token, ok = <-tokens
			if !ok {
				break
			}
			// TODO: ugly hack for compatiblity with older crossplane parsing
			//       of comments following args (on the side, rather than above)
			if strings.HasPrefix(token.Item, "#") {
				if args.Comments {
					if strings.HasPrefix(token.Item, "#@") {
						snip := strings.TrimSpace(token.Item[2:])
						if snip != "" {
							debugf("setting tag to: %q\n", snip)
							tag = snip
							block.tag = tag
						}
					}

					block.Block = append(block.Block, &Directive{
						Directive: "#",
						Comment:   token.Item[1:],
					})
				}
				continue
			}
		}

		if block.Directive == "if" {
			block.Args = removeBrackets(block.Args)
		}

		if !args.Single && block.Directive == "include" {
			if len(block.Args) == 0 {
				return nil, errors.New("no parameter for include")
			}
			// skip it if excluded
			if among(block.Args[0], args.Ignore...) {
				continue
			}
			pattern := filepath.Join(baseDir, block.Args[0])
			fnames, err := filepath.Glob(pattern)
			if err != nil {
				if args.CatchErrors {
					parseErr := ParseError{
						Fail:   err,
						Line:   token.LineNum,
						Column: token.Column,
						File:   args.FileName,
					}

					p.Errors = append(p.Errors, parseErr)
					continue
				}
				return nil, err
			}

			// TODO: might a file be included mutliple times (places)?
			// Add test for that, won't matter until we're successfully
			// applying policies to imported files
			for _, fname := range fnames {
				debugf("Import: %s (%30s)\n", fname, pattern)
				args.FileName = fname
				if err = p.readFile(args); err != nil {
					return nil, err
				}
				added := len(p.Config) - 1
				block.Includes = append(block.Includes, added)
			}
		}

		if token.Item == "{" {
			block.Block, e = p.parse(config, tokens, args)
			if e != nil {
				return out, e
			}
		}
		if tag != "" {
			fmt.Printf("\nDIR: %s TAG: %s\n", block.Directive, tag)
		}
		block.tag = tag
		tag = ""
		out = append(out, block)
	}
	return out, nil
}

func removeBrackets(brackets []string) []string {
	for i, s := range brackets {
		brackets[i] = strings.Trim(s, "()")
	}
	return brackets
}

// replaces blocks that include other configs by their respective blocks
// used to consolidate a multi-file config
func (p *Payload) blockIncludes(blocks []*Directive) []*Directive {
	var updated []*Directive
	for _, block := range blocks {
		if len(block.Includes) == 0 {
			block.Block = p.blockIncludes(block.Block)
			updated = append(updated, block)
			continue
		}
		for _, include := range block.Includes {
			included := p.Config[include].Parsed
			updated = append(updated, p.blockIncludes(included)...)
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

	newConfig := &Config{
		File:   p.Config[0].File,
		Parsed: p.blockIncludes(p.Config[0].Parsed),
	}

	singular := &Payload{
		File:   p.File,
		Config: []*Config{newConfig},
	}
	// TODO: merge consolidated errors
	return singular, nil
}

// Dump marshals the payload as json to the writer
func (p *Payload) Dump(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p); err != nil {
		log.Println("failed to dump file:", err)
	}
}

// Render writes out the config back to its original form
// NOTE: the render funcs are effectively copies of the pkg/builder funcs
//       The goal was to refactor as a "universal" writer that will
//       export the config directly to a file(s), tarball, or a string
// TODO: complete this!
func (p *Payload) Render(w io.Writer) error {
	if _, err := io.WriteString(w, "# rendered by CrossPlane\n\n"); err != nil {
		return err
	}
	return RenderDirectives(w, p.Config[0].Parsed)
}

// RenderDirectives writes out a config
// adapted from block.BuildDirective
func RenderDirectives(w io.Writer, blocks []*Directive) error {
	return renderDirectives(w, blocks, 0, 1)
}

func renderDirectives(w io.Writer, blocks []*Directive, depth, lastline int) (err error) {
	// TODO: kinda goofy, rethink this
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	WriteString := func(ss ...string) {
		for _, s := range ss {
			if _, err := io.WriteString(w, s); err != nil {
				panic(err)
			}
		}
	}
	padding := "  "
	spacing := 4 // borrowed from pkg/builder -- rethink this
	margin := strings.Repeat(padding, depth)
	line := lastline
	for _, stmt := range blocks {
		line++
		WriteString(margin)
		if stmt.Directive == "#" {
			WriteString("#"+stmt.Comment, "\n")
			continue
		} else {
			if stmt.Directive == "if" {
				WriteString("if (" + strings.Join(stmt.Args, " ") + ")")
			} else if len(stmt.Args) > 0 {
				WriteString(stmt.Directive + " " + strings.Join(stmt.Args, " "))
			} else {
				WriteString(stmt.Directive)
			}

			if len(stmt.Block) < 1 {
				WriteString(";\n")
			} else {
				WriteString(" {\n")
				_ = renderDirectives(w, stmt.Block, depth+1, line)
				WriteString(margin + "}\n")
				if spacing != 0 {
					spacing -= 4
				}
			}
		}
	}
	return err
}

// TODO: move below to types after merge

// Statement struct used for analysing directives and other information
type Statement struct {
	Directive string
	Args      []string
	Line      int
}

// TODO: move back under directive after merge
func equals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if x != b[i] {
			return false
		}
	}
	return true
}

// Equal returns true if both blocks are functionally equivalent
func (d *Directive) Equal(a *Directive) bool {
	// TODO: add more equality tests!
	return (a.Directive == d.Directive && equals(a.Args, d.Args))
}

// Insert inserts blocks into the blocks child blocks
// TODO: make it private?
func (d *Directive) Insert(index int, children ...*Directive) error {
	if d == nil {
		return errors.New("nil directive")
	}
	if index >= len(d.Directive) {
		return fmt.Errorf("index out of range [%d] with length %d", index, len(d.Directive))
	}
	// courtesey of slicetricks wiki
	blocks := (*d).Block
	(*d).Block = append(blocks[:index], append(children, blocks[index:]...)...)
	return nil
}

// Name returns the string representation of the directive's name
func (d *Directive) Name() string {
	const dirSep = "_"
	switch d.Directive {
	case "location":
		return d.Directive + dirSep + strings.Join(d.Args, pathSep)
	case "server":
		names := subVal("server_name", d.Block)
		if len(names) == 0 {
			names = []string{"unknown"}
		}
		listen := subVal("listen", d.Block)
		if len(listen) == 0 {
			panic("no listener")
		}
		debugf("LISTEN: %q\n", listen[0])
		return fmt.Sprintf("%s%s%s%s%s", d.Directive, dirSep, names[0], dirSep, listen[0])
	case "#":
		return dirSep + d.Comment
	}
	return d.Directive
}

// TODO: gather helpers
func among(item string, list ...string) bool {
	for _, x := range list {
		if item == x {
			return true
		}
	}
	return false
}

// NOTE: This (and it's invocation) will all go once code has crystalized

// Debugging enables debug output
var Debugging bool

func debugf(msg string, args ...interface{}) {
	if Debugging {
		log.Printf(msg, args...)
	}
}
