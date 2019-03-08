package parser

import (
	"testing"
)

func TestIncludes(t *testing.T) {
	var includePayload = []struct {
		payload []Config
	}{
		{
			payload: []Config{
				{
					Title:  "regular",
					Status: "ok",
					Parsed: []Block{
						Block{
							Directive: "include",
							Line:      1,
							Args:      []string{" "},
						},
					},
				},
			},
		},
	}
}
