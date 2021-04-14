package crossplane

import (
	"fmt"
)

type ParseError struct {
	what string
	file *string
	line *int
}

func (e ParseError) Error() string {
	file := "(nofile)"
	if e.file != nil {
		file = *e.file
	}
	if e.line != nil {
		return fmt.Sprintf("%s in %s:%d", e.what, file, *e.line)
	}
	return fmt.Sprintf("%s in %s", e.what, file)
}
