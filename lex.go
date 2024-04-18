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

type LexScanner interface {
	Scan() bool
	Err() error
	Text() string
	Line() int
}

type ExtLexer interface {
	Register(LexScanner) []string
	Lex(matchedToken string) <-chan NgxToken
}

type LexOptions struct {
	ExternalLexers []ExtLexer
}

func LexWithOptions(r io.Reader, options LexOptions) chan NgxToken {
	tc := make(chan NgxToken, tokChanCap)
	go tokenize(r, tc, options)
	return tc
}

func Lex(reader io.Reader) chan NgxToken {
	return LexWithOptions(reader, LexOptions{})
}

type extScanner struct {
	scanner   *bufio.Scanner
	tokenLine int
}

func (e *extScanner) Scan() bool {
	if !e.scanner.Scan() {
		return false
	}
	if t := e.scanner.Text(); isEOL(t) {
		e.tokenLine++
	}
	return true
}

func (e *extScanner) Err() error   { return e.scanner.Err() }
func (e *extScanner) Text() string { return e.scanner.Text() }
func (e *extScanner) Line() int    { return e.tokenLine }

//nolint:gocyclo,funlen,gocognit
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

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanRunes)

	emit := func(line int, quoted bool, err error) {
		tokenCh <- NgxToken{Value: token.String(), Line: line, IsQuoted: quoted, Error: err}
		token.Reset()
		lexState = skipSpace
	}

	var externalLexers map[string]ExtLexer
	var externalScanner *extScanner
	for _, ext := range options.ExternalLexers {
		if externalLexers == nil {
			externalLexers = make(map[string]ExtLexer)
		}

		if externalScanner == nil {
			externalScanner = &extScanner{scanner: scanner, tokenLine: tokenLine}
		}

		for _, d := range ext.Register(externalScanner) {
			externalLexers[d] = ext
		}
	}

	for {
		if readNext {
			if !scanner.Scan() {
				break // done
			}

			la = scanner.Text()
			if isEOL(la) {
				tokenLine++
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
			if ext, ok := externalLexers[tokenStr]; ok {
				emit(tokenStartLine, false, nil)
				// for {
				// 	tok, err := ext.Lex()

				// 	if errors.Is(err, StopExtLexer{}) {
				// 		break
				// 	}

				// 	if err != nil {
				// 		emit(tokenStartLine, false, &ParseError{
				// 			File: &lexerFile,
				// 			What: err.Error(),
				// 			Line: &tokenLine, // TODO: I See this used but never updated
				// 		})
				// 		continue
				// 	}
				// 	tokenCh <- tok
				// }
				externalScanner.tokenLine = tokenLine
				extTokenCh := ext.Lex(tokenStr)
				for tok := range extTokenCh {
					tokenCh <- tok
				}

				lexState = skipSpace
				tokenLine = externalScanner.tokenLine
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
					lexState = inComment
					tokenStartLine = tokenLine
					continue
				}
			}

			if isSpace(la) {
				emit(tokenStartLine, false, nil)
				continue
			}

			// handle parameter expansion syntax (ex: "${var[@]}")
			if token.Len() > 0 && strings.HasSuffix(token.String(), "$") && la == "{" {
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

type LuaLexer struct {
	s LexScanner
}

func (ll *LuaLexer) Register(s LexScanner) []string {
	ll.s = s
	return []string{
		"init_by_lua_block",
		"init_worker_by_lua_block",
		"exit_worker_by_lua_block",
		"set_by_lua_block",
		"content_by_lua_block",
		"server_rewrite_by_lua_block",
		"rewrite_by_lua_block",
		"access_by_lua_block",
		"header_filter_by_lua_block",
		"body_filter_by_lua_block",
		"log_by_lua_block",
		"balancer_by_lua_block",
		"ssl_client_hello_by_lua_block",
		"ssl_certificate_by_lua_block",
		"ssl_session_fetch_by_lua_block",
		"ssl_session_store_by_lua_block",
	}
}

func (ll *LuaLexer) Lex(matchedToken string) <-chan NgxToken {
	tokenCh := make(chan NgxToken)

	tokenDepth := 0
	// var leadingWhitespace strings.Builder

	go func() {
		defer close(tokenCh)
		var tok strings.Builder
		var inQuotes bool
		var startedText bool

		if matchedToken == "set_by_lua_block" {
			arg := ""
			for {
				if !ll.s.Scan() {
					return
				}

				next := ll.s.Text()
				if isSpace(next) {
					if arg != "" {
						tokenCh <- NgxToken{Value: arg, Line: ll.s.Line(), IsQuoted: false}
						break
					}

					for isSpace(next) {
						if !ll.s.Scan() {
							return
						}
						next = ll.s.Text()
					}
				}
				arg += next
			}
		}

		for {
			if !ll.s.Scan() {
				return
			}
			if next := ll.s.Text(); !isSpace(next) {
				break
			}
		}

		if !ll.s.Scan() {
			// TODO: err?
			return
		}

		if next := ll.s.Text(); next != "{" {
			// TODO: Return error, need { to open blockj
		}

		tokenDepth += 1

		// TODO: check for opening brace?

		for {
			if !ll.s.Scan() {
				return // shrug emoji
			}

			if err := ll.s.Err(); err != nil {
				lineno := ll.s.Line()
				tokenCh <- NgxToken{Error: &ParseError{File: &lexerFile, What: err.Error(), Line: &lineno}}

			}

			next := ll.s.Text()

			// Handle leading whitespace
			if !startedText && isSpace(next) && tokenDepth == 0 && !inQuotes {
				// leadingWhitespace.WriteString(next)
				continue
			}

			// As soon as we hit a nonspace, we consider text to have started.
			startedText = true

			switch {
			case next == "{" && !inQuotes:
				tokenDepth++

			case next == "}" && !inQuotes:
				tokenDepth--
				if tokenDepth < 0 {
					lineno := ll.s.Line()
					tokenCh <- NgxToken{Error: &ParseError{File: &lexerFile, What: `unexpected "}"`, Line: &lineno}}
					return
				}

				if tokenDepth == 0 {
					// tokenCh <- NgxToken{Value: next, Line: 0}
					// Before finishing the block, prepend any leading whitespace.
					// finalTokenValue := leadingWhitespace.String() + tok.String()
					tokenCh <- NgxToken{Value: tok.String(), Line: ll.s.Line(), IsQuoted: true}
					// tokenCh <- NgxToken{Value: finalTokenValue, Line: ll.s.Line(), IsQuoted: true}
					// tok.Reset()
					// leadingWhitespace.Reset()
					tokenCh <- NgxToken{Value: ";", Line: ll.s.Line(), IsQuoted: false} // For an end to the Lua string, seems hacky.
					// See: https://github.com/nginxinc/crossplane/blob/master/crossplane/ext/lua.py#L122C25-L122C41
					return
				}
			case next == `"` || next == "'":
				if !inQuotes {
					inQuotes = true
				} else {
					inQuotes = false
				}
				tok.WriteString(next)

			// case isSpace(next):
			// 	if tok.Len() == 0 {
			// 		continue
			// 	}
			// 	tok.WriteString(next)
			default:
				// Expected first token encountered to be a "{" to open a Lua block. Handle any other non-whitespace
				// character to mean we are not about to tokenize Lua.
				if tokenDepth == 0 {
					tokenCh <- NgxToken{Value: next, Line: ll.s.Line(), IsQuoted: false}
					return
				}
				tok.WriteString(next)
			}

			// if tokenDepth == 0 {
			// 	return
			// }
			if tokenDepth == 0 && !inQuotes {
				startedText = false
				// leadingWhitespace.Reset()
			}
		}
	}()

	return tokenCh
}
