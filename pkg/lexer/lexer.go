package lexer

import (
	"bufio"
	"io"
	"strings"
)

// UnbalancedBracesError -
type UnbalancedBracesError string

// LexicalItem -
type LexicalItem struct {
	Item    string
	LineNum int
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

func consumeWord(data []byte) (int, []byte, error) {
	var accum []byte
	var o int
	var skip bool
	for i, b := range data {
		// TODO make this more robust
		if (b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ';' || b == '{') && data[i-1] != '\\' && data[i-1] != '$' {
			return i, accum, nil
		}
		accum = append(accum, b)
	}

	return 0, nil, nil
}

func consumeNum(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if '0' <= b && b <= '9' || b == '.' || b == ':' {
			accum = append(accum, b)
		} else {
			return i, accum, nil
		}
	}
	return len(accum), accum, nil
}

func consumeString(data []byte) (int, []byte, error) {
	delim := data[0]
	var otherStringDelim byte
	if delim == '"' {
		otherStringDelim = '\''
	} else {
		otherStringDelim = '"'
	}
	skip := false
	var accum []byte
	for i, b := range data[1:] {
		if b == delim && !skip {

			return i + 2, accum, nil
		}
		skip = false
		if b == '\\' && data[i+2] == delim {
			skip = true
			continue
		} else if b == '\\' || b == otherStringDelim {
			accum = append(accum, '\\')
		}
		accum = append(accum, b)
	}
	return 0, nil, nil
}

func consumeComment(data []byte) (int, []byte, error) {
	var accum []byte

	for i, b := range data {
		if b != '\n' && i < len(data) {
			accum = append(accum, b)
		} else {
			return i, accum, nil
		}
	}
	return 0, nil, nil

}

// Reader -
type Reader struct {
	*bufio.Scanner
	lastL, lastCol int
	l, col         int
}

// LexScanner -
func LexScanner(input string) <-chan LexicalItem {
	s := NewLexer(strings.NewReader(input))
	chnl := make(chan LexicalItem)
	go func() {
		for s.Scan() {
			tok := s.Bytes()
			if string(tok) != " " && string(tok) != "\t" && string(tok) != "\n" {
				chnl <- LexicalItem{string(tok), s.l}

			}
		}
		close(chnl)
	}()

	return chnl
}

// NewLexer -
func NewLexer(r io.Reader) *Reader {
	s := bufio.NewScanner(r)
	rdr := &Reader{
		Scanner: s,
	}
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if rdr.l == 0 {
			rdr.l = 1
			rdr.col = 1
			rdr.lastL = 1
			rdr.lastCol = 1
		}
		if atEOF {
			return
		}
		switch data[0] {

		case '{', '}', ';':
			advance, token, err = 1, data[:1], nil
		case '"', '\'':
			advance, token, err = consumeString(data)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			advance, token, err = consumeNum(data)
		case ' ', '\n', '\r', '\t':
			advance, token, err = 1, data[:1], nil
		case '#':
			advance, token, err = consumeComment(data)
		default:

			advance, token, err = consumeWord(data)
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
		return
	}
	s.Split(split)
	return rdr
}
