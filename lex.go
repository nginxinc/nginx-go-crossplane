/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type NgxToken struct {
	Value    string
	Line     int
	IsQuoted bool
	Error    error
}

type state int

const (
	skipSpace state = iota
	inWord
	inComment
	inVar
	inQuote
)

const TokenChanCap = 2048

//nolint:gochecknoglobals
var lexerFile = "lexer" // pseudo file name for use by parse errors

//nolint:gochecknoglobals
var tokChanCap = TokenChanCap // capacity of lexer token channel

// note: this is only used during tests, should not be changed
func SetTokenChanCap(size int) {
	tokChanCap = size
}

// Lexer is an interface for implementing lexers that handle external NGINX tokens during the lexical analysis phase.
type Lexer interface {
	// Lex processes a matched token and returns a channel of NgxToken objects.
	// This method performs lexical analysis on the matched token and produces a stream of tokens for the parser to consume.
	// The external lexer should close the channel once it has completed lexing the input to signal the end of tokens.
	// Failure to close the channel will cause the receiver to wait indefinitely.
	Lex(s *SubScanner, matchedToken string) <-chan NgxToken
}

// LexOptions allows customization of the lexing process by specifying external lexers
// that can handle specific directives. By registering interest in particular directives,
// external lexers can ensure that these directives are processed separately
// from the general lexical analysis logic.
type LexOptions struct {
	Lexers    []RegisterLexer
	extLexers map[string]Lexer
}

// RegisterLexer is an option that cna be used to add a lexer to tokenize external NGINX tokens.
type RegisterLexer interface {
	applyLexOptions(options *LexOptions)
}

type registerLexer struct {
	l            Lexer
	stringTokens []string
}

func (rl registerLexer) applyLexOptions(o *LexOptions) {
	if o.extLexers == nil {
		o.extLexers = make(map[string]Lexer)
	}

	for _, s := range rl.stringTokens {
		o.extLexers[s] = rl.l
	}
}

// LexWithLexer registers a Lexer that implements tokenization of an NGINX configuration after one of the given
// stringTokens is encountered by Lex.
func LexWithLexer(l Lexer, stringTokens ...string) RegisterLexer {
	return registerLexer{l: l, stringTokens: stringTokens}
}

func LexWithOptions(r io.Reader, options LexOptions) chan NgxToken {
	for _, o := range options.Lexers {
		o.applyLexOptions(&options)
	}

	tc := make(chan NgxToken, tokChanCap)
	go tokenize(r, tc, options)
	return tc
}

func Lex(reader io.Reader) chan NgxToken {
	return LexWithOptions(reader, LexOptions{})
}

type SubScanner struct {
	scanner   *bufio.Scanner
	tokenLine int
}

func (e *SubScanner) Scan() bool {
	if !e.scanner.Scan() {
		return false
	}
	if t := e.scanner.Text(); isEOL(t) {
		e.tokenLine++
	}
	return true
}

func (e *SubScanner) Err() error   { return e.scanner.Err() }
func (e *SubScanner) Text() string { return e.scanner.Text() }
func (e *SubScanner) Line() int    { return e.tokenLine }

//nolint:gocyclo,funlen,gocognit,maintidx
func tokenize(reader io.Reader, tokenCh chan NgxToken, options LexOptions) {
	token := strings.Builder{}
	tokenLine := 1
	tokenStartLine := 1

	lexState := skipSpace
	newToken := false
	dupSpecialChar := false
	readNext := true
	esc := false
	depth := 0
	var la, quote string
	nextTokenIsDirective := true

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanRunes)

	emit := func(line int, quoted bool, err error) {
		tokenCh <- NgxToken{Value: token.String(), Line: line, IsQuoted: quoted, Error: err}
		token.Reset()
		lexState = skipSpace
	}

	for {
		if readNext {
			if !scanner.Scan() {
				break // done
			}

			la = scanner.Text()
			if isEOL(la) {
				tokenLine++
				nextTokenIsDirective = true
			}
		} else {
			readNext = true
		}

		// skip CRs
		if la == "\r" || la == "\\\r" {
			continue
		}

		if la == "\\" && !esc {
			esc = true
			continue
		}
		if esc {
			esc = false
			la = "\\" + la
		}

		if token.Len() > 0 {
			tokenStr := token.String()
			if nextTokenIsDirective {
				if ext, ok := options.extLexers[tokenStr]; ok {
					// saving lex state before emitting tokenStr to know if we encountered start quote
					lastLexState := lexState
					emit(tokenStartLine, lexState == inQuote, nil)

					externalScanner := &SubScanner{scanner: scanner, tokenLine: tokenLine}
					extTokenCh := ext.Lex(externalScanner, tokenStr)
					for tok := range extTokenCh {
						tokenCh <- tok
					}
					tokenLine = externalScanner.tokenLine

					// if we detected a start quote and current char after external lexer processing is end quote we skip it
					if lastLexState == inQuote && la == quote {
						continue
					}
				}
			}
		}

		switch lexState {
		case skipSpace:
			if !isSpace(la) {
				lexState = inWord
				newToken = true
				readNext = false // re-eval
				tokenStartLine = tokenLine
			}
			continue
		case inWord:
			if newToken {
				newToken = false
				if la == "#" {
					token.WriteString(la)
					nextTokenIsDirective = false
					lexState = inComment
					tokenStartLine = tokenLine
					continue
				}
			}

			if isSpace(la) {
				emit(tokenStartLine, false, nil)
				nextTokenIsDirective = false
				continue
			}

			// handle parameter expansion syntax (ex: "${var[@]}")
			if token.Len() > 0 && strings.HasSuffix(token.String(), "$") && la == "{" {
				nextTokenIsDirective = false
				token.WriteString(la)
				lexState = inVar
				dupSpecialChar = false
				continue
			}

			// if a quote is found, add the whole string to the token buffer
			if la == `"` || la == "'" {
				if token.Len() > 0 {
					// if a quote is inside a token, treat it like any other char
					token.WriteString(la)
				} else {
					// swallow quote and change state
					quote = la
					lexState = inQuote
					tokenStartLine = tokenLine
				}
				dupSpecialChar = false
				continue
			}

			// handle special characters that are treated like full tokens
			if la == "{" || la == "}" || la == ";" {
				// if token complete yield it and reset token buffer
				if token.Len() > 0 {
					emit(tokenStartLine, false, nil)
				}

				// only '}' can be repeated
				if dupSpecialChar && la != "}" {
					emit(tokenStartLine, false, &ParseError{
						File: &lexerFile,
						What: fmt.Sprintf(`unexpected "%s"`, la),
						Line: &tokenLine,
					})
					close(tokenCh)
					return
				}

				dupSpecialChar = true

				if la == "{" {
					depth++
				}
				if la == "}" {
					depth--
					// early exit if unbalanced braces
					if depth < 0 {
						emit(tokenStartLine, false, &ParseError{File: &lexerFile, What: `unexpected "}"`, Line: &tokenLine})
						close(tokenCh)
						return
					}
				}

				token.WriteString(la)
				// this character is a full token so emit it
				emit(tokenStartLine, false, nil)
				nextTokenIsDirective = true
				continue
			}

			dupSpecialChar = false
			token.WriteString(la)

		case inComment:
			if isEOL(la) {
				emit(tokenStartLine, false, nil)
				continue
			}
			token.WriteString(la)

		case inVar:
			token.WriteString(la)
			// this is using the same logic as the exiting lexer, but this is wrong since it does not terminate on token boundary
			if !strings.HasSuffix(token.String(), "}") && !isSpace(la) {
				continue
			}
			lexState = inWord

		case inQuote:
			if la == quote {
				emit(tokenStartLine, true, nil)
				continue
			}
			if la == "\\"+quote {
				la = quote
			}
			token.WriteString(la)
		}
	}

	if token.Len() > 0 {
		emit(tokenStartLine, lexState == inQuote, nil)
	}
	if depth > 0 {
		emit(tokenStartLine, false, &ParseError{File: &lexerFile, What: `unexpected end of file, expecting "}"`, Line: &tokenLine})
	}

	close(tokenCh)
}
