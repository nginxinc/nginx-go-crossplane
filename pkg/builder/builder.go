package builder

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

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

// FileString is a string representation of a file
type FileString struct {
	Name     string
	Contents string
	w        strings.Builder
}

// Write makes this a writer
func (fs *FileString) Write(b []byte) (int, error) {
	return fs.w.Write(b)
}

// Close makes this a an io.Closer
func (fs *FileString) Close() error {
	fs.Contents = fs.w.String()
	return nil
}

// StringsCreator is an option for rendering config files to strings(s)
type StringsCreator struct {
	Dir   string
	Files []FileString
}

// Create makes this a Creator
func (sc *StringsCreator) Create(file string) (io.WriteCloser, error) {
	idx := len(sc.Files)
	sc.Files = append(sc.Files, FileString{Name: file})
	return &(sc.Files[idx]), nil
}

// Options define how the config is rendered
type Options struct {
	Dirname string
	Indent  int
	Tabs    bool
	Header  bool
	Creator Creator
	Writer  io.WriteCloser
}

// Build takes a parsed NGINX configuration and builds an NGINX configuration
func Build(parsed []*parser.Directive, opts *Options) string {
	if len(parsed) == 0 {
		log.Println("warning: no directives to build with")
		return ""
	}

	if opts == nil {
		opts = &Options{Indent: 2}
	}

	if opts.Writer == nil {
		opts.Writer = &writeCloser{ioutil.Discard}
	}

	var b strings.Builder

	// payloads without line numbers get their own renderer
	if parsed[0].Line == 0 {
		if err := buildBlockNumberless(&b, parsed, 0, opts); err != nil {
			log.Fatal(err)
		}
		return b.String()

	}
	i, err := buildBlock(&b, parsed, 0, 0, opts)
	if err != nil {
		log.Printf("build size: %d -- error: %v\n", i, err)
	}
	s := b.String()
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

// buildBlock recursively builds NGINX configuration blocks
func buildBlock(
	w io.Writer,
	block []*parser.Directive,
	depth,
	lastLine int,
	opts *Options,
) (int, error) {
	if lastLine == 0 {
		lastLine = 1
	}

	var err error
	defer func() {
		if r := recover(); r != nil {
			if perr, ok := r.(error); ok {
				err = perr
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
			log.Println("PANIC'd with:", err)
		}
	}()

	writeString := func(strs ...string) {
		for _, s := range strs {
			if _, err := w.Write([]byte(s)); err != nil {
				panic(fmt.Errorf("writeString for %q -- error: %w", s, err))
			}
		}
	}

	indented := indent(depth, opts)

	for _, stmt := range block {
		same := true
		thisLine := stmt.Line

		for prev := lastLine; prev < thisLine; prev++ {
			writeString("\n")
			same = false
		}
		if !same {
			writeString(indented)
		}
		if stmt.IsComment() && same {
			writeString(" #", stmt.Comment)
			continue
		} else if stmt.IsComment() {
			writeString("#", stmt.Comment)
		} else {
			if stmt.IsIf() {
				writeString("if (", strings.Join(stmt.Args, " "), ")")
			} else if len(stmt.Args) > 0 {
				writeString(stmt.Directive, " ", strings.Join(stmt.Args, " "))
			} else {
				writeString(stmt.Directive)
			}

			if stmt.Block == nil {
				writeString(";")
			} else {
				writeString(" {")
				thisLine, err = buildBlock(w, stmt.Block, depth+1, stmt.Line, opts)
				if err != nil {
					return thisLine, err
				}
				if thisLine > stmt.Line {
					writeString("\n", indented)
				}
				writeString("}\n")
				thisLine += 2 // move past the brace
			}
		}
		lastLine = thisLine
	}
	return lastLine, err
}

// when using a payload that has no line numbers
func buildBlockNumberless(
	w io.Writer,
	block []*parser.Directive,
	depth int,
	opts *Options,
) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if perr, ok := r.(error); ok {
				err = perr
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
			log.Println("PANIC'd with:", err)
		}
	}()

	writeString := func(strs ...string) {
		for _, s := range strs {
			if _, err := w.Write([]byte(s)); err != nil {
				panic(fmt.Errorf("writeString for %q -- error: %w", s, err))
			}
		}
	}

	indented := indent(depth, opts)

	for _, stmt := range block {
		writeString(indented)
		if stmt.IsComment() {
			writeString(" #", stmt.Comment)
			continue
		}
		if stmt.IsIf() {
			writeString("if (", strings.Join(stmt.Args, " "), ")")
		} else if len(stmt.Args) > 0 {
			writeString(stmt.Directive, " ", strings.Join(stmt.Args, " "))
		} else {
			writeString(stmt.Directive)
		}

		if stmt.Block == nil {
			writeString(";\n")
		} else {
			writeString(" {\n")
			err = buildBlockNumberless(w, stmt.Block, depth+1, opts)
			if err != nil {
				return err
			}
			writeString("}\n")
		}
	}
	return err
}

func indent(depth int, opts *Options) string {
	if opts.Tabs {
		return strings.Repeat("\t", depth)
	}
	return strings.Repeat(" ", opts.Indent*depth)
}

// BuildFiles renders the crossplane payload back to native nginx configs
// TODO: remove returning string -- just return error
//        if string is desired, then use a StringCreator
func BuildFiles(payload *parser.Payload, opts *Options) (string, error) {
	var built strings.Builder
	var output string

	if opts == nil {
		opts = &Options{}
	}
	// sensible defaults
	if opts.Indent == 0 {
		opts.Indent = 2
	}

	mkdir := false
	creator := opts.Creator
	if creator == nil {
		creator = osCreator{}
		mkdir = true
	}

	for _, config := range payload.Config {
		path := config.File
		path = filepath.Join(opts.Dirname, path)
		if mkdir {
			if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				return "", err
			}
		}
		w, err := creator.Create(path)
		if err != nil {
			return "", err
		}
		opts.Writer = w

		parsed := config.Parsed
		output = Build(parsed, opts)
		output = strings.TrimLeft(output, "\n")
		if _, err = w.Write([]byte(output)); err != nil {
			return "", err
		}
		if err = w.Close(); err != nil {
			return "", err
		}

		if _, err = built.WriteString(output); err != nil {
			return "", err
		}
		w.Write([]byte("\n"))
	}
	return built.String(), nil
}
