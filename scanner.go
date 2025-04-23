package crossplane

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type scannerOptions struct {
	extensions map[string]ScannerExt
}

type ScannerOption interface {
	applyScannerOptions(options *scannerOptions)
}

// Token is a lexical token of the NGINX configuration syntax.
type Token struct {
	// Text is the string corresponding to the token. It could be a directive or symbol. The value is the actual token
	// sequence in order to support defining directives in modules other than the core NGINX module set.
	Text string
	// Line is the source starting line number the token within a file.
	Line int
	// IsQuoted signifies if the token is wrapped by quotes (", '). Quotes are not usually necessary in an NGINX
	// configuration and mostly serve to help make the config less ambiguous.
	IsQuoted bool
}

func (t Token) String() string { return fmt.Sprintf("{%d, %s, %t}", t.Line, t.Text, t.IsQuoted) }

type scannerError struct {
	msg  string
	line int
}

func (e *scannerError) Error() string { return e.msg }
func (e *scannerError) Line() int     { return e.line }

func newScannerErrf(line int, format string, a ...any) *scannerError {
	return &scannerError{line: line, msg: fmt.Sprintf(format, a...)}
}

// LineNumber reports the line on which the error occurred by finding the first error in
// the errs chain that returns a line number. Otherwise, it returns 0, false.
//
// An error type should provide a Line() int method to return a line number.
func LineNumber(err error) (int, bool) {
	var e interface{ Line() int }
	if !errors.As(err, &e) {
		return 0, false
	}

	return e.Line(), true
}

// Scanner provides an interface for tokenizing an NGINX configuration. Successive calls to the Scane method will step
// through the 'tokens; of an NGINX configuration.
//
// Scanning stops unrecoverably at EOF, the first I/O error, or an unexpected token.
//
// Use NewScanner to construct a Scanner.
type Scanner struct {
	scanner              *bufio.Scanner
	lineno               int
	tokenStartLine       int
	tokenDepth           int
	repeateSpecialChar   bool //  only '}' can be repeated
	nextTokenIsDirective bool
	prev                 string
	err                  error
	options              *scannerOptions
	ext                  Tokenizer
}

// NewScanner returns a new Scanner to read from r.
func NewScanner(r io.Reader, options ...ScannerOption) *Scanner {
	opts := &scannerOptions{}
	for _, opt := range options {
		opt.applyScannerOptions(opts)
	}

	s := &Scanner{
		scanner:              bufio.NewScanner(r),
		lineno:               1,
		tokenStartLine:       1,
		tokenDepth:           0,
		repeateSpecialChar:   false,
		nextTokenIsDirective: true,
		options:              opts,
	}

	s.scanner.Split(bufio.ScanRunes)

	return s
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *Scanner) setErr(err error) {
	if s.err == nil || s.err != io.EOF {
		s.err = err
	}
}

// Scan reads the next token from source and returns it.. It returns io.EOF at the end of the source. Scanner errors are
// returned when encountered.
func (s *Scanner) Scan() (Token, error) { //nolint: funlen, gocognit, gocyclo, maintidx // sorry
	if s.ext != nil {
		t, err := s.ext.Next()
		if err != nil {
			if !errors.Is(err, ErrTokenizerDone) {
				s.setErr(err)
				return Token{}, s.err
			}

			s.ext = nil
		} else {
			return t, nil
		}
	}

	var tok strings.Builder

	lexState := skipSpace
	newToken := false
	readNext := true
	esc := false

	var r, quote string

	for {
		if s.err != nil {
			return Token{}, s.err
		}

		switch {
		case s.prev != "":
			r, s.prev = s.prev, ""
		case readNext:
			if !s.scanner.Scan() {
				if tok.Len() > 0 {
					return Token{Text: tok.String(), Line: s.tokenStartLine, IsQuoted: lexState == inQuote}, nil
				}

				if s.tokenDepth > 0 {
					s.setErr(&scannerError{line: s.tokenStartLine, msg: "unexpected end of file, expecting }"})
					return Token{}, s.err
				}

				s.setErr(io.EOF)
				return Token{}, s.err
			}

			nextRune := s.scanner.Text()
			r = nextRune
			if isEOL(r) {
				s.lineno++
				s.nextTokenIsDirective = true
			}
		default:
			readNext = true
		}

		// skip CRs
		if r == "\r" || r == "\\\r" {
			continue
		}

		if r == "\\" && !esc {
			esc = true
			continue
		}

		if esc {
			esc = false
			r = "\\" + r
		}

		if tok.Len() > 0 {
			t := tok.String()
			if s.nextTokenIsDirective {
				if ext, ok := s.options.extensions[t]; ok {
					s.ext = ext.Tokenizer(&SubScanner{parent: s, tokenLine: s.tokenStartLine}, t)
					return Token{Text: t, Line: s.tokenStartLine}, nil
				}
			}
		}

		switch lexState {
		case skipSpace:
			if !isSpace(r) {
				lexState = inWord
				newToken = true
				readNext = false // re-eval
				s.tokenStartLine = s.lineno
			}
			continue

		case inWord:
			if newToken {
				newToken = false
				if r == "#" {
					tok.WriteString(r)
					lexState = inComment
					s.tokenStartLine = s.lineno
					s.nextTokenIsDirective = false
					continue
				}
			}

			if isSpace(r) {
				s.nextTokenIsDirective = false
				return Token{Text: tok.String(), Line: s.tokenStartLine}, nil
			}

			// parameter expansion syntax (ex: "${var[@]}")
			if tok.Len() > 0 && strings.HasSuffix(tok.String(), "$") && r == "{" {
				tok.WriteString(r)
				lexState = inVar
				s.repeateSpecialChar = false
				s.nextTokenIsDirective = false
				continue
			}

			// add entire quoted string to the token buffer
			if r == `"` || r == "'" {
				if tok.Len() > 0 {
					// if a quote is inside a token, treat it like any other char
					tok.WriteString(r)
				} else {
					quote = r
					lexState = inQuote
					s.tokenStartLine = s.lineno
				}
				s.repeateSpecialChar = false
				continue
			}

			// special characters treated as full tokens
			if isSpecialChar(r) {
				if tok.Len() > 0 {
					s.prev = r
					return Token{Text: tok.String(), Line: s.tokenStartLine}, nil
				}

				// only } can be repeated
				if s.repeateSpecialChar && r != "}" {
					s.setErr(newScannerErrf(s.tokenStartLine, "unxpected %q", r))
					return Token{}, s.err
				}

				s.repeateSpecialChar = true
				if r == "{" {
					s.tokenDepth++
				}

				if r == "}" {
					s.tokenDepth--
					if s.tokenDepth < 0 {
						s.setErr(&scannerError{line: s.tokenStartLine, msg: `unexpected "}"`})
						return Token{}, s.err
					}
				}

				tok.WriteString(r)
				s.nextTokenIsDirective = true
				return Token{Text: tok.String(), Line: s.tokenStartLine}, nil
			}

			s.repeateSpecialChar = false
			tok.WriteString(r)
		case inComment:
			if isEOL(r) {
				return Token{Text: tok.String(), Line: s.tokenStartLine}, nil
			}
			tok.WriteString(r)
		case inVar:
			tok.WriteString(r)
			if r != "}" && !isSpace(r) {
				continue
			}
			lexState = inWord
		case inQuote:
			if r == quote {
				return Token{Text: tok.String(), Line: s.tokenStartLine, IsQuoted: true}, nil
			}
			if r == "\\"+quote {
				r = quote
			}
			tok.WriteString(r)
		}
	}
}

// ScannerExt is the interface that describes an extension for the [Scanner]. Scanner extensions enable scanning of
// configurations that contain syntaxes that do not follow the usual grammar.
type ScannerExt interface {
	Tokenizer(s *SubScanner, matchedToken string) Tokenizer
}

// ErrTokenizerDone is returned by [Tokenizer] when tokenization is complete.
var ErrTokenizerDone = errors.New("done")

// Tokenizer is the interface that wraps the Next method.
//
// Next returns the next token scanned from the NGINX configuration or an error if the configuration cannot be
// tokenized. Return the special error, [ErrTokenizerDone] when finished tokenizing.
type Tokenizer interface {
	Next() (Token, error)
}

// LexerScanner is a compatibility layer between Lexers and Scanner.
type LexerScanner struct {
	lexer        Lexer
	scanner      *SubScanner
	matchedToken string
	ch           <-chan NgxToken
}

func (s *LexerScanner) Tokenizer(scanner *SubScanner, matchedtoken string) Tokenizer {
	s.scanner = scanner
	s.matchedToken = matchedtoken
	return s
}

func (s *LexerScanner) Next() (Token, error) {
	if s.ch == nil {
		s.ch = s.lexer.Lex(s.scanner, s.matchedToken)
	}

	ngxTok, ok := <-s.ch
	if !ok {
		return Token{}, ErrTokenizerDone
	}

	if ngxTok.Error != nil {
		return Token{}, newScannerErrf(ngxTok.Line, ngxTok.Error.Error())
	}

	return Token{Text: ngxTok.Value, Line: ngxTok.Line, IsQuoted: ngxTok.IsQuoted}, nil
}
