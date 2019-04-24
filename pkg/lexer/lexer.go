package lexer

import (
	"bufio"
	"io"
	"strings"
	"sync"
)

// UnbalancedBracesError -
type UnbalancedBracesError string

// LexicalItem -
type LexicalItem struct {
	item    string
	lineNum int
}

// balance braces found in a lexical item array.
// returns an error if lexemes right and left braces are not balanced
func balanceBraces(lexicalItems []LexicalItem) UnbalancedBracesError {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		balance int
	)

	for _, lexicalItem := range lexicalItems {
		wg.Add(1)
		go func(i LexicalItem) {
			defer wg.Done()
			mu.Lock()
			switch i.item {
			case "{":
				balance = balance + 1
				break
			case "}":
				balance = balance - 1
				break
			}
			mu.Unlock()
		}(lexicalItem)
	}
	wg.Wait()
	if balance != 0 {
		return "UnbalancedBracesError: braces are not balanced"
	}
	return ""

}

func consumeWord(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			return i, accum, nil
		} else {
			accum = append(accum, b)
		}
	}
	return 0, nil, nil
}

func consumeWhitespace(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			accum = append(accum, b)
		} else {
			return i, accum, nil
			// return i, data[len(accum):len(data)], nil
		}
	}
	return 0, nil, nil
}

func consumeNum(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if '0' <= b && b <= '9' {
			accum = append(accum, b)
		} else {
			return i, accum, nil
		}
	}
	return len(accum), accum, nil
}

func consumeString(data []byte) (int, []byte, error) {
	delim := data[0]
	skip := false
	accum := []byte{data[0]}
	for i, b := range data[1:] {
		if b == delim && !skip {
			return i + 2, accum, nil
		}
		skip = false
		if b == '\\' {
			skip = true
			continue
		}
		accum = append(accum, b)
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
func LexScanner(input string) ([]LexicalItem, error) {
	s := NewLexer(strings.NewReader(input))
	res := []LexicalItem{}
	for s.Scan() {
		tok := s.Bytes()
		res = append(res, LexicalItem{string(tok), 1})
	}
	return res, nil
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
		case '"', '\'': // TODO(jwall): Rune data?
			advance, token, err = consumeString(data)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			advance, token, err = consumeNum(data)
		case ' ', '\n', '\r', '\t':
			advance, token, err = 1, data[:1], nil
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
