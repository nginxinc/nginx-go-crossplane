package parser

// ParseArgs -
type ParseArgs struct {
	FileName    string
	OnError     bool
	CatchErrors bool
	Ignore      string
	Single      bool
	Comments    bool
	Strict      bool
	Combine     bool
	CheckCtx    bool
	CheckArgs   bool
}

// Config -
type Config struct {
	Title  string
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
	Title string
	Line  int
	Error string
}

// Parse -
func Parse(args ParseArgs) (string, error) {

	return "parse", nil
}
