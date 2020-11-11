package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// UnbalancedBracesError -
type UnbalancedBracesError string

// LexicalItem -
type LexicalItem struct {
	Item    string
	LineNum int
	Column  int
}

// for identifyig Lua directives
var byLua = []byte(`_by_lua_`)

// String satisfies the Stringer interface
func (l *LexicalItem) String() string {
	return fmt.Sprintf("%d) %s", l.LineNum, l.Item)
}

// Repr returns a representation of a lex item.
// uses return type interface{} as it could either be LexicalItem or
// a bare string
func (l *LexicalItem) Repr(lineNumbers bool) interface{} {
	if lineNumbers {
		return l
	}
	return l.Item
}

// BalanceBraces found in a lexical item array.
// returns an error if lexemes right and left braces are not balanced
func BalanceBraces(lexicalItems []LexicalItem) UnbalancedBracesError {
	balance := 0
	for _, lexicalItem := range lexicalItems {
		switch lexicalItem.Item {
		case "{":
			balance = balance + 1
		case "}":
			balance = balance - 1
		}
	}
	if balance != 0 {
		return UnbalancedBracesError("UnbalancedBracesError: braces are not balanced")
	}
	return UnbalancedBracesError("")
}

func consumeWord(data []byte, isLua bool) (int, []byte, bool, error) {
	var accum []byte

	if isLua {
		return consumeLuaDirective(data)
	}
	for i, b := range data {
		// TODO make this more robust
		if (b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ';' || b == '{') && data[i-1] != '\\' && data[i-1] != '$' {
			isLua = bytes.Contains(accum, byLua)
			return i, accum, isLua, nil
		}
		accum = append(accum, b)
	}
	return 0, nil, isLua, nil
}

func consumeNum(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ';' || b == '{' {
			return i, accum, nil
		}
		accum = append(accum, b)
	}
	return len(accum), accum, nil
}

// consume a quoted string
func consumeString(data []byte, isLua, keepQuotes bool) (int, []byte, bool, error) {
	if isLua {
		return consumeLuaDirective(data)
	}
	delim := data[0]
	skip := false
	var accum []byte

	if keepQuotes {
		accum = append(accum, delim)
	}

	for i, b := range data[1:] {
		if b == delim {
			if !skip {
				isLua = bytes.Contains(accum, byLua)
				if keepQuotes {
					accum = append(accum, delim)
				}
				return i + 2, accum, isLua, nil
			}
		}
		skip = false
		if b == '\\' && i < len(data)-2 && data[i+2] == delim {
			skip = true
			continue
		}
		accum = append(accum, b)
	}
	return 0, nil, isLua, nil
}

// read to end of line
func consumeComment(data []byte) (int, []byte, error) {
	index := bytes.IndexByte(data, '\n')
	if index < 0 {
		return 0, nil, nil

	}
	return index, data[:index], nil
}

// consume every character between the braces of the directive
func consumeLuaDirective(data []byte) (int, []byte, bool, error) {
	var accum []byte

	// we came here after an opening brace
	// TODO: clean up the quoted logic
	count := 1
	var quote byte
	quoted := 0
	for i, b := range data {
		if b == '\'' || b == '"' {
			if quoted == 0 {
				quoted++
				quote = b
			} else if b == quote && data[i-1] != '\\' {
				quoted--
			}
		}

		if quoted == 0 {
			if b == '}' {
				count--
				if count <= 0 {
					// don't include the closing brace, as that's
					// needed as a token to signal the parser we're done here
					return i - 1, accum, false, nil
				}
			}
			if b == '{' {
				count++
			}
		}

		accum = append(accum, b)
	}
	return 0, nil, false, nil
}

// Reader -
type Reader struct {
	*bufio.Scanner
	lastL, lastCol int
	l, col         int
}

// LexScanReader lexes the given reader
func LexScanReader(r io.Reader, unquoted bool) <-chan LexicalItem {
	s := NewLexer(r, unquoted)
	chnl := make(chan LexicalItem, 1)
	go func() {
		for s.Scan() {
			tok := string(s.Bytes())
			// TODO: tok can equal " " in some config torture tests
			// (it should be clarified if it is a valid concern)
			if tok != "" && tok != " " && tok != "\t" && tok != "\n" {
				chnl <- LexicalItem{tok, s.l, s.col}
			}
		}
		close(chnl)
	}()
	return chnl
}

// LexScanner -
func LexScanner(input string, unquoted bool) <-chan LexicalItem {
	return LexScanReader(strings.NewReader(input), unquoted)
}

// NewLexer -
func NewLexer(r io.Reader, unquoted bool) *Reader {
	s := bufio.NewScanner(r)
	rdr := &Reader{
		Scanner: s,
	}
	consume := false
	isLua := false
	level := 0
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if rdr.l == 0 {
			rdr.l = 1
			rdr.col = 1
			rdr.lastL = 1
			rdr.lastCol = 1
		}
		// TODO: or just check data len?
		if atEOF && len(data) == 0 {
			return
		}

		if isLua && consume {
			advance, token, isLua, err = consumeLuaDirective(data)
			consume = false
		} else {
			switch data[0] {
			case '{', '}', ';':
				if data[0] == '}' {
					level--
					isLua = false
					consume = false
				} else if data[0] == '{' {
					level++
					consume = isLua
				}
				advance, token, err = 1, data[:1], nil
				if isLua && data[0] == '{' {
					// parser normally takes open brace as start of a new block,
					// but with Lua we're treating the contents as args,
					// so make the brace a NOP to skip it
					token = data[:0]
				}
			case '"', '\'':
				advance, token, isLua, err = consumeString(data, isLua, !unquoted)
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				advance, token, err = consumeNum(data)
			case ' ', '\n', '\r', '\t':
				advance, token, err = 1, data[:1], nil
			case '#':
				advance, token, err = consumeComment(data)
			default:
				advance, token, isLua, err = consumeWord(data, isLua)
			}
		}

		if advance > 0 {
			rdr.lastCol = rdr.col
			rdr.lastL = rdr.l
			for _, b := range data[:advance] {
				if b == '\n' || atEOF {
					rdr.l++
					rdr.col = 1
				}
				rdr.col++
			}
		}
		// fiddly bit adjustment for python compatability
		// trim trailing backslash (if it's not backslashed itself)
		n := len(token)
		if n > 1 {
			prior, last := token[n-2], token[n-1]
			if last == '\\' && prior != '\\' {
				token = token[:n-1]
			}
		}
		return
	}
	s.Split(split)
	return rdr
}
