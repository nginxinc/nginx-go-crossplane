package parser

import "encoding/json"

// ParseArgs -
type ParseArgs struct {
	FileName string
	// onerror     bool
	CatchErrors bool
	Ignore      string
	Single      bool
	Comments    bool
	Strict      bool
	Combine     bool
	// checkCtx    bool
	// checkArgs bool
}

// ParsingError -
type ParsingError string

// Config -
type Config struct {
	File   string
	Status string
	Errors []Errors
	Parsed []Block
}

// Block -
type Block struct {
	Directive string
	Line      int
	Args      []string
	Includes  []int
	Block     []Block
}

//Errors -
type Errors struct {
	File  string
	Line  int
	Error string
}

// Parse -
func Parse(args ParseArgs) (string, error) {

	return "parse", nil
}

// Parsing -
func Parsing(config []Config) ParsingError {
	err := json.Marshal(tests)
	if err != nil {
		return "ParsingError"
	}
	return ""
}
