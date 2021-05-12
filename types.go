package crossplane

type Payload struct {
	Status string         `json:"status"`
	Errors []PayloadError `json:"errors"`
	Config []Config       `json:"config"`
}

type PayloadError struct {
	File     string      `json:"file"`
	Line     *int        `json:"line"`
	Error    error       `json:"error"`
	Callback interface{} `json:"callback,omitempty"`
}

type Config struct {
	File   string        `json:"file"`
	Status string        `json:"status"`
	Errors []ConfigError `json:"errors"`
	Parsed Directives    `json:"parsed"`
}

type ConfigError struct {
	Line  *int  `json:"line"`
	Error error `json:"error"`
}

type Directive struct {
	Directive string     `json:"directive"`
	Line      int        `json:"line"`
	Args      []string   `json:"args"`
	Includes  []int      `json:"includes,omitempty"`
	Block     Directives `json:"block,omitempty"`
	Comment   *string    `json:"comment,omitempty"`
}
type Directives []*Directive

// IsBlock returns true if this is a block directive.
func (d Directive) IsBlock() bool {
	return d.Block != nil
}

// IsInclude returns true if this is an include directive.
func (d Directive) IsInclude() bool {
	return d.Directive == "include" && d.Includes != nil
}

// IsComment returns true iff the directive represents a comment.
func (d Directive) IsComment() bool {
	return d.Directive == "#" && d.Comment != nil
}

// Combined returns a new Payload that is the same except that the inluding
// logic is performed on its configs. This means that the resulting Payload
// will always have 0 or 1 configs in its Config field.
func (p *Payload) Combined() (*Payload, error) {
	return combineConfigs(p)
}
