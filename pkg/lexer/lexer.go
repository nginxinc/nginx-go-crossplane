package lexer

// UnbalancedBracesError -
type UnbalancedBracesError string

// LexicalItem -
type LexicalItem struct {
	item    string
	lineNum int
}

// balance braces found in a lexical item array.
// returns an error if lexemes right and left braces are not balanced
func balanceBraces([]LexicalItem) UnbalancedBracesError {
	return ""
}

// Lex - Generates tokens from an nginx config file
func Lex(filename string) ([]LexicalItem, error) {
	return []LexicalItem{}, nil
}
