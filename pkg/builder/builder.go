package builder

import (
	"encoding/json"
	"fmt"
)

// Block -
type Block struct {
	Directive string
	Line      int
	Args      []string
	Comment   string
	Block     []Block
}

// ConfFiles -
type ConfFiles struct {
	File   string
	Status string
	Errors string
	Config []ConfFiles
	Parsed []Block
}

// Build takes a string representing NGINX configuration
// builds it into conf format and returns that as a string
func Build(payload string, indent int, tabs, header bool) (string, error) {
	data := Block{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling payload: %v", err)
	}

	return "built", nil
}

// BuildFiles -
func BuildFiles(payload string, dirname string, indent int, tabs, header bool) (string, error) {
	data := ConfFiles{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling payload: %v", err)
	}

	return "built", nil
}
