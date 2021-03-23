package builder

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"gitlab.com/f5/nginx/crossplane-go/pkg/parser"
)

const DefaultIndent = 2

// Creator abstracts file creation (to write configs to something other than files)
type Creator interface {
	Create(string) (io.WriteCloser, error)
}

type osCreator struct{}

// Create satisfies the Creator interface
func (c osCreator) Create(filename string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0777); err != nil {
		return nil, err
	}
	return os.Create(filename)
}

type ioCreator struct {
	w io.Writer
}

// NewIOCreator returns a creator that renders to a writer
// All config files will be rendered to the given writer
func NewIOCreator(w io.Writer) Creator {
	if w == nil {
		w = os.Stdout
	}
	return &ioCreator{w}
}

type writeCloser struct {
	io.Writer
}

// Close is a NOP Close handler
func (*writeCloser) Close() error { return nil }

// Create satisifies the Creator interface
func (c *ioCreator) Create(file string) (io.WriteCloser, error) {
	return &writeCloser{c.w}, nil
}

// StringFile is a string representation of a file
type StringFile struct {
	Name     string
	Contents string
}

// StringWriter is a string representation of a file
type StringWriter struct {
	Name     string
	Contents string
	w        strings.Builder
}

// Write makes this a writer
func (fs *StringWriter) Write(b []byte) (int, error) {
	return fs.w.Write(b)
}

// Close makes this a an io.Closer
func (fs *StringWriter) Close() error {
	fs.Contents = fs.w.String()
	return nil
}

// StringsCreator is an option for rendering config files to strings(s)
type StringsCreator struct {
	Dir   string
	Files []StringWriter
}

// Create makes this a Creator
func (sc *StringsCreator) Create(file string) (io.WriteCloser, error) {
	idx := len(sc.Files)
	sc.Files = append(sc.Files, StringWriter{Name: file})
	return &(sc.Files[idx]), nil
}

// String returns the first file as a string
func (sc *StringsCreator) String() string {
	return sc.Files[0].Contents
}

// Options define how the config is rendered
type Options struct {
	Dirname  string
	Indent   int
	Tabs     bool
	Header   bool
	Block    bool
	Spacer   bool
	Numbered bool // honor line numbers if config data has 'em
	Enquote  bool
	Creator  Creator
	Writer   io.WriteCloser
}

// Build takes a parsed NGINX configuration and builds an NGINX configuration
func Build(parsed []*parser.Directive, opts *Options) error {
	if len(parsed) == 0 {
		return fmt.Errorf("no directives to build with")
	}

	if opts == nil {
		opts = &Options{Indent: DefaultIndent}
	}

	if opts.Writer == nil {
		opts.Writer = &writeCloser{ioutil.Discard}
	}

	// TODO: we should use "normal" line numbers, not zero offset
	r := &renderer{w: opts.Writer, opts: opts}
	if opts.Numbered {
		for _, d := range parsed {
			if d.Line > 0 {
				r.numbered = true
				break
			}
		}
	}
	_, err := r.buildBlock(parsed, 0, 0, opts.Spacer)
	return err
}

func indent(depth int, opts *Options) string {
	if opts.Indent <= 0 {
		return ""
	}
	if opts.Tabs {
		return strings.Repeat("\t", depth)
	}
	return strings.Repeat(" ", opts.Indent*depth)
}

// BuildStrings renders the payload config(s) as strings rather than on the filesystem
func BuildStrings(payload *parser.Payload, opts *Options) ([]StringFile, error) {
	if opts == nil {
		opts = &Options{}
	}
	// sensible defaults
	if opts.Indent == 0 {
		opts.Indent = DefaultIndent
	}
	if opts.Creator != nil {
		return nil, fmt.Errorf("unexpected Creator option -- should be nil")
	}

	sc := &StringsCreator{}
	opts.Creator = sc
	if err := BuildFiles(payload, opts); err != nil {
		return nil, err
	}
	files := make([]StringFile, len(sc.Files))
	for i, file := range sc.Files {
		files[i].Name = file.Name
		files[i].Contents = file.Contents
	}
	return files, nil
}

// BuildFiles renders the crossplane payload back to native nginx configs
func BuildFiles(payload *parser.Payload, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	// sensible defaults
	if opts.Indent == 0 {
		opts.Indent = DefaultIndent
	}

	mkdir := false
	creator := opts.Creator
	if creator == nil {
		creator = osCreator{}
		mkdir = true
	}

	for _, config := range payload.Config {
		path := filepath.Join(opts.Dirname, config.File)
		if mkdir {
			if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				return err
			}
		}
		w, err := creator.Create(path)
		if err != nil {
			return err
		}
		opts.Writer = w

		parsed := config.Parsed
		if err := Build(parsed, opts); err != nil {
			return err
		}

		if err = w.Close(); err != nil {
			return err
		}
	}
	return nil
}

type renderer struct {
	w        io.Writer
	err      error
	opts     *Options
	numbered bool
}

func (r *renderer) writeString(strs ...string) {
	if r.err != nil {
		return
	}
	for _, s := range strs {
		if _, err := fmt.Fprint(r.w, s); err != nil {
			r.err = err
			return
		}
	}
}

func (r *renderer) buildBlock(block []*parser.Directive, lastline, depth int, spacer bool) (int, error) {
	indented := indent(depth, r.opts)
	trailing := false
	// write sameline on following loop to allow for comments on the same line
	var sameline bool
	until := len(block) - 1 // pre-calc the last block
	for count, stmt := range block {
		if trailing && count > 0 {
			r.writeString("\n")
			trailing = false
			sameline = false
			lastline++
		}
		if r.numbered {
			for prev := lastline; prev < stmt.Line; prev++ {
				r.writeString("\n")
				trailing = false
				sameline = false
			}
			lastline = stmt.Line
		}
		r.writeString(indented)

		if stmt.IsComment() {
			if sameline {
				r.writeString(" ")
			}
			r.writeString("#", stmt.Comment)
			if spacer {
				r.writeString("\n")
				lastline++
			}
			trailing = !spacer
			sameline = false
			continue
		}

		if stmt.IsIf() {
			r.writeString("if (")
			r.writeArgs(stmt.Args)
			r.writeString(")")
		} else if len(stmt.Args) > 0 {
			r.writeString(stmt.Directive, " ")
			r.writeArgs(stmt.Args)
		} else {
			r.writeString(stmt.Directive)
		}
		sameline = true

		if stmt.Block == nil {
			r.writeString(";")
			trailing = true
		} else {
			r.writeString(" {")
			// special handling for `events {}` to keep that formatting
			if stmt.Directive == "events" && len(stmt.Block) == 0 {
				r.writeString("}\n")
				lastline++
				sameline = false
				trailing = true
				continue
			}
			r.writeString("\n")
			lastline++
			sameline = false
			var err error
			lastline, err = r.buildBlock(stmt.Block, lastline, depth+1, spacer)
			if err != nil {
				return lastline, err
			}
			r.writeString(indented, "}\n")
			if spacer && count < until {
				r.writeString("\n")
			}
			sameline = false
			trailing = false
			lastline++
		}
		if r.err != nil {
			return lastline, r.err
		}
	}
	if trailing {
		r.writeString("\n")
		lastline++
	}
	return lastline, nil
}

func (r *renderer) writeArgs(args []string) {
	if !r.opts.Enquote {
		r.writeString(strings.Join(args, " "))
		return
	}
	for i, arg := range args {
		if i > 0 {
			r.writeString(" ")
		}
		r.writeString(enquote(arg))
	}
}

func repr(s string) string {
	q := fmt.Sprintf("%q", s)
	for _, char := range q {
		if char == '"' {
			q = strings.ReplaceAll(q, `\"`, `"`)
			q = strings.ReplaceAll(q, `'`, `\'`)
			q = `'` + q[1:len(q)-1] + `'`
			return q
		}
	}
	return q
}

func enquote(arg string) string {
	if !needsQuote(arg) {
		return arg
	}
	return strings.ReplaceAll(repr(arg), `\\`, `\`)
}

func needsQuote(s string) bool {
	if s == "" {
		return true
	}

	// lexer should throw an error when variable expansion syntax
	// is messed up, but just wrap it in quotes for now I guess
	var char rune
	chars := escape(s)

	if len(chars) == 0 {
		return true
	}

	// get first rune
	char, off := utf8.DecodeRune([]byte(chars))

	// arguments can't start with variable expansion syntax
	if unicode.IsSpace(char) || strings.ContainsRune("{};\"'", char) || strings.HasPrefix(chars, "${") {
		return true
	}

	chars = chars[off:]

	expanding := false
	var prev rune = 0
	for _, c := range chars {
		char = c

		if prev == '\\' {
			prev = 0
			continue
		}
		if unicode.IsSpace(char) || strings.ContainsRune("{;\"'", char) {
			return true
		}

		if (expanding && (prev == '$' && char == '{')) || (!expanding && char == '}') {
			return true
		}

		if (expanding && char == '}') || (!expanding && (prev == '$' && char == '{')) {
			expanding = !expanding
		}

		prev = char
	}

	return expanding || char == '\\' || char == '$'
}

func escape(s string) string {
	if !strings.ContainsAny(s, "{}$;\\") {
		return s
	}

	sb := strings.Builder{}
	var pc, cc rune

	for _, r := range s {
		cc = r
		if pc == '\\' || (pc == '$' && cc == '{') {
			sb.WriteRune(pc)
			sb.WriteRune(cc)
			pc = 0
			continue
		}

		if pc == '$' {
			sb.WriteRune(pc)
		}
		if cc != '\\' && cc != '$' {
			sb.WriteRune(cc)
		}
		pc = cc
	}

	if cc == '\\' || cc == '$' {
		sb.WriteRune(cc)
	}

	return sb.String()
}
