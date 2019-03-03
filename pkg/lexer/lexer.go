package lexer

import "sync"

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

// Lex - Generates tokens from an nginx config file
func Lex(filename string) ([]LexicalItem, error) {
	return []LexicalItem{}, nil
}
