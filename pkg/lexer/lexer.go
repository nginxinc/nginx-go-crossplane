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
			case "}":
				balance = balance - 1
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

func iterescape(data string) []byte {
	var b []byte
	d := []byte(data)
	for i, v := range d {
		if v == '\\' && i+1 < len(d) {
			if d[i+1] != '"' && d[i+1] != '\'' {
				b = append(b, '\\')
			}
		} else if v == '\'' {
			b = append(b, '\\')
		}
		b = append(b, v)
	}
	if len(b) < len(d) {
		b = append(b, data[len(b):]...)
	}
	if len(b) < 1 {
		return d
	}
	return b
}

func consumeWord(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ';' {
			return i, accum, nil
		} else if b == '\'' {
			accum = append(accum, '\\')
		}
		accum = append(accum, b)
	}
	return 0, nil, nil
}

func consumeNum(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if '0' <= b && b <= '9' || b == '.' {
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
	accum := []byte{}
	for i, b := range data[1:] {
		if b == delim && !skip {
			return i + 2, accum, nil
		}
		skip = false
		if b == '\\' {
			skip = true
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
func LexScanner(input string) ([]LexicalItem, error) {
	s := NewLexer(strings.NewReader(string(iterescape(input))))
	res := []LexicalItem{}
	for s.Scan() {
		tok := s.Bytes()
		if string(tok) != " " && string(tok) != "\t" && string(tok) != "\n" {
			res = append(res, LexicalItem{string(tok), s.l})
		}
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
		case '"':
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
