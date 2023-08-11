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
func Lex(reader io.Reader) chan NgxToken {
	tc := make(chan NgxToken, tokChanCap)
	go tokenize(reader, tc)
	return tc
}

//nolint:gocyclo,gocognit,cyclop,funlen
func tokenize(reader io.Reader, tokenCh chan NgxToken) {
	token := strings.Builder{}
	tokenLine := 1
	tokenStartLine := 1

	lexState := skipSpace
	newToken := false
	dupSpecialChar := false
	readNext := true
	esc := false
	depth := 0
	var inputText, quote string

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

			inputText = scanner.Text()
			if isEOL(inputText) {
				tokenLine++
			}
		} else {
			readNext = true
		}

		// skip CRs
		if inputText == "\r" || inputText == "\\\r" {
			continue
		}

		if inputText == "\\" && !esc {
			esc = true
			continue
		}
		if esc {
			esc = false
			inputText = "\\" + inputText
		}

		switch lexState {
		case skipSpace:
			if !isSpace(inputText) {
				lexState = inWord
				newToken = true
				readNext = false // re-eval
				tokenStartLine = tokenLine
			}
			continue
		case inWord:
			if newToken {
				newToken = false
				if inputText == "#" {
					token.WriteString(inputText)
					lexState = inComment
					tokenStartLine = tokenLine
					continue
				}
			}

			if isSpace(inputText) {
				emit(tokenStartLine, false, nil)
				continue
			}

			// handle parameter expansion syntax (ex: "${var[@]}")
			if token.Len() > 0 && strings.HasSuffix(token.String(), "$") && inputText == "{" {
				token.WriteString(inputText)
				lexState = inVar
				dupSpecialChar = false
				continue
			}

			// if a quote is found, add the whole string to the token buffer
			if inputText == `"` || inputText == "'" {
				if token.Len() > 0 {
					// if a quote is inside a token, treat it like any other char
					token.WriteString(inputText)
				} else {
					// swallow quote and change state
					quote = inputText
					lexState = inQuote
					tokenStartLine = tokenLine
				}
				dupSpecialChar = false
				continue
			}

			// handle special characters that are treated like full tokens
			if inputText == "{" || inputText == "}" || inputText == ";" {
				// if token complete yield it and reset token buffer
				if token.Len() > 0 {
					emit(tokenStartLine, false, nil)
				}

				// only '}' can be repeated
				if dupSpecialChar && inputText != "}" {
					emit(tokenStartLine, false, &ParseError{
						File: &lexerFile,
						What: fmt.Sprintf(`unexpected "%s"`, inputText),
						Line: &tokenLine,
					})
					close(tokenCh)
					return
				}

				dupSpecialChar = true

				if inputText == "{" {
					depth++
				}
				if inputText == "}" {
					depth--
					// early exit if unbalanced braces
					if depth < 0 {
						emit(tokenStartLine, false, &ParseError{File: &lexerFile, What: `unexpected "}"`, Line: &tokenLine})
						close(tokenCh)
						return
					}
				}

				token.WriteString(inputText)
				// this character is a full token so emit it
				emit(tokenStartLine, false, nil)
				continue
			}

			dupSpecialChar = false
			token.WriteString(inputText)

		case inComment:
			if isEOL(inputText) {
				emit(tokenStartLine, false, nil)
				continue
			}
			token.WriteString(inputText)

		case inVar:
			// this is using the same logic as the exiting lexer, but this is wrong since it does not terminate on token boundary
			token.WriteString(inputText)
			if !strings.HasSuffix(token.String(), "}") && !isSpace(inputText) {
				continue
			}
			lexState = inWord

		case inQuote:
			if inputText == quote {
				emit(tokenStartLine, true, nil)
				continue
			}
			if inputText == "\\"+quote {
				inputText = quote
			}
			token.WriteString(inputText)
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
