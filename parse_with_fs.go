package crossplane

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var (
	memFsOpen = func(memfs fs.FS, path string) (io.ReadCloser, error) { return memfs.Open(path) }
)

var openFunc func(memfs fs.FS, path string) (io.ReadCloser, error)

type MemFsParseOptions struct {
	ParseOptions
	Open func(memfs fs.FS, path string) (io.ReadCloser, error)
}

func ParseWithMemFs(memfs fs.FS, filename string, options *MemFsParseOptions) (*Payload, error) {
	payload := &Payload{
		Status: "ok",
		Errors: []PayloadError{},
		Config: []Config{},
	}
	if options.Glob == nil {
		options.Glob = filepath.Glob
	}

	handleError := func(config *Config, err error) {
		var line *int
		if e, ok := err.(*ParseError); ok {
			line = e.Line
		}
		cerr := ConfigError{Line: line, Error: err}
		perr := PayloadError{Line: line, Error: err, File: config.File}
		if options.ErrorCallback != nil {
			perr.Callback = options.ErrorCallback(err)
		}

		const failedSts = "failed"
		config.Status = failedSts
		config.Errors = append(config.Errors, cerr)

		payload.Status = failedSts
		payload.Errors = append(payload.Errors, perr)
	}

	if options.Open != nil {
		openFunc = options.Open
	}

	p := parser{
		configDir:   filepath.Dir(filename),
		options:     &options.ParseOptions,
		handleError: handleError,
		includes:    []fileCtx{{path: filename, ctx: blockCtx{}}},
		included:    map[string]int{filename: 0},
		// adjacency list where an edge exists between a file and the file it includes
		includeEdges: map[string][]string{},
		// number of times a file is included by another file
		includeInDegree: map[string]int{filename: 0},
	}
	for len(p.includes) > 0 {
		incl := p.includes[0]
		p.includes = p.includes[1:]

		file, err := p.openMemfsFile(memfs, incl.path)
		if err != nil {
			return nil, err
		}

		defer file.Close()

		tokens := LexWithOptions(file, options.LexOptions)
		config := Config{
			File:   incl.path,
			Status: "ok",
			Errors: []ConfigError{},
			Parsed: Directives{},
		}
		parsed, err := p.memfsParse(memfs, &config, tokens, incl.ctx, false)
		if err != nil {
			if options.StopParsingOnError {
				return nil, err
			}
			handleError(&config, err)
		} else {
			config.Parsed = parsed
		}

		payload.Config = append(payload.Config, config)
	}
	if p.isAcyclic() {
		return nil, errors.New("configs contain include cycle")
	}

	if options.CombineConfigs {
		return payload.Combined()
	}

	return payload, nil
}

func (p *parser) openMemfsFile(memfs fs.FS, path string) (io.ReadCloser, error) {
	open := memFsOpen
	if openFunc != nil {
		open = openFunc
	}
	return open(memfs, path)
}

// parse Recursively parses directives from an nginx config context.
//
//nolint:gocyclo,funlen,gocognit,maintidx,nonamedreturns
func (p *parser) memfsParse(memfs fs.FS, parsing *Config, tokens <-chan NgxToken, ctx blockCtx, consume bool) (parsed Directives, err error) {
	var tokenOk bool
	// parse recursively by pulling from a flat stream of tokens
	for t := range tokens {
		if t.Error != nil {
			var perr *ParseError
			if errors.As(t.Error, &perr) {
				perr.File = &parsing.File
				perr.BlockCtx = ctx.getLastBlock()
				return nil, perr
			}
			return nil, &ParseError{
				What:        t.Error.Error(),
				File:        &parsing.File,
				Line:        &t.Line,
				originalErr: t.Error,
				BlockCtx:    ctx.getLastBlock(),
			}
		}

		var commentsInArgs []string

		// we are parsing a block, so break if it's closing
		if t.Value == "}" && !t.IsQuoted {
			break
		}

		// if we are consuming, then just continue until end of context
		if consume {
			// if we find a block inside this context, consume it too
			if t.Value == "{" && !t.IsQuoted {
				_, _ = p.memfsParse(memfs, parsing, tokens, nil, true)
			}
			continue
		}

		var fileName string
		if p.options.CombineConfigs {
			fileName = parsing.File
		}

		// the first token should always be an nginx directive
		stmt := &Directive{
			Directive: t.Value,
			Line:      t.Line,
			Args:      []string{},
			File:      fileName,
		}

		// if token is comment
		if strings.HasPrefix(t.Value, "#") && !t.IsQuoted {
			if p.options.ParseComments {
				comment := t.Value[1:]
				stmt.Directive = "#"
				stmt.Comment = &comment
				parsed = append(parsed, stmt)
			}
			continue
		}

		// parse arguments by reading tokens
		t, tokenOk = <-tokens
		if !tokenOk {
			return nil, &ParseError{
				What:        ErrPrematureLexEnd.Error(),
				File:        &parsing.File,
				Line:        &stmt.Line,
				originalErr: ErrPrematureLexEnd,
				BlockCtx:    ctx.getLastBlock(),
			}
		}
		for t.IsQuoted || (t.Value != "{" && t.Value != ";" && t.Value != "}") {
			if !strings.HasPrefix(t.Value, "#") || t.IsQuoted {
				stmt.Args = append(stmt.Args, t.Value)
			} else if p.options.ParseComments {
				commentsInArgs = append(commentsInArgs, t.Value[1:])
			}
			t, tokenOk = <-tokens
			if !tokenOk {
				return nil, &ParseError{
					What:        ErrPrematureLexEnd.Error(),
					File:        &parsing.File,
					Line:        &stmt.Line,
					originalErr: ErrPrematureLexEnd,
					BlockCtx:    ctx.getLastBlock(),
				}
			}
		}

		// if inside "map-like" block - add contents to payload, but do not parse further
		if len(ctx) > 0 {
			if _, ok := mapBodies[ctx[len(ctx)-1]]; ok {
				mapErr := analyzeMapBody(parsing.File, stmt, t.Value, ctx[len(ctx)-1])
				if mapErr != nil && p.options.StopParsingOnError {
					return nil, mapErr
				} else if mapErr != nil {
					p.handleError(parsing, mapErr)
					// consume invalid block
					if t.Value == "{" && !t.IsQuoted {
						_, _ = p.memfsParse(memfs, parsing, tokens, nil, true)
					}
					continue
				}
				stmt.IsMapBlockParameter = true
				parsed = append(parsed, stmt)
				continue
			}
		}

		// consume the directive if it is ignored and move on
		if contains(p.options.IgnoreDirectives, stmt.Directive) {
			// if this directive was a block consume it too
			if t.Value == "{" && !t.IsQuoted {
				_, _ = p.memfsParse(memfs, parsing, tokens, nil, true)
			}
			continue
		}

		// raise errors if this statement is invalid
		err = analyze(parsing.File, stmt, t.Value, ctx, p.options)

		if perr, ok := err.(*ParseError); ok && !p.options.StopParsingOnError {
			p.handleError(parsing, perr)
			// if it was a block but shouldn"t have been then consume
			if strings.HasSuffix(perr.What, ` is not terminated by ";"`) {
				if t.Value != "}" && !t.IsQuoted {
					_, _ = p.memfsParse(memfs, parsing, tokens, nil, true)
				} else {
					break
				}
			}
			// keep on parsin'
			continue
		} else if err != nil {
			return nil, err
		}

		// prepare arguments - strip parentheses
		if stmt.Directive == "if" {
			stmt = prepareIfArgs(stmt)
		}

		// add "includes" to the payload if this is an include statement
		if !p.options.SingleFile && stmt.Directive == "include" {
			if len(stmt.Args) == 0 {
				return nil, &ParseError{
					What: fmt.Sprintf(`invalid number of arguments in "%s" directive in %s:%d`,
						stmt.Directive,
						parsing.File,
						stmt.Line,
					),
					File:      &parsing.File,
					Line:      &stmt.Line,
					Statement: stmt.String(),
					BlockCtx:  ctx.getLastBlock(),
				}
			}

			pattern := stmt.Args[0]
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(p.configDir, pattern)
			}

			// get names of all included files
			var fnames []string
			if hasMagic.MatchString(pattern) {
				fnames, err = p.options.Glob(pattern)
				if err != nil {
					return nil, err
				}
				sort.Strings(fnames)
			} else {
				// if the file pattern was explicit, nginx will check
				// that the included file can be opened and read
				if f, err := p.openMemfsFile(memfs, pattern); err != nil {
					perr := &ParseError{
						What:      err.Error(),
						File:      &parsing.File,
						Line:      &stmt.Line,
						Statement: stmt.String(),
						BlockCtx:  ctx.getLastBlock(),
					}
					if !p.options.StopParsingOnError {
						p.handleError(parsing, perr)
					} else {
						return nil, perr
					}
				} else {
					defer f.Close()
					fnames = []string{pattern}
				}
			}

			for _, fname := range fnames {
				// the included set keeps files from being parsed twice
				// TODO: handle files included from multiple contexts
				if _, ok := p.included[fname]; !ok {
					p.included[fname] = len(p.included)
					p.includes = append(p.includes, fileCtx{fname, ctx})
				}
				stmt.Includes = append(stmt.Includes, p.included[fname])
				// add edge between the current file and it's included file and
				// increase the included file's in degree
				p.includeEdges[parsing.File] = append(p.includeEdges[parsing.File], fname)
				p.includeInDegree[fname]++
			}
		}

		// if this statement terminated with "{" then it is a block
		if t.Value == "{" && !t.IsQuoted {
			stmt.Block = make(Directives, 0)
			inner := enterBlockCtx(stmt, ctx) // get context for block
			blocks, err := p.memfsParse(memfs, parsing, tokens, inner, false)
			if err != nil {
				return nil, err
			}
			stmt.Block = append(stmt.Block, blocks...)
		}

		parsed = append(parsed, stmt)

		// add all comments found inside args after stmt is added
		for _, comment := range commentsInArgs {
			comment := comment
			parsed = append(parsed, &Directive{
				Directive: "#",
				Line:      stmt.Line,
				Args:      []string{},
				File:      fileName,
				Comment:   &comment,
			})
		}
	}

	return parsed, nil
}
