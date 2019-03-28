package parser

import (
	"testing"
)

/*
func TestIncludes(t *testing.T) {
	var includePayload = []struct {
		payload []Config
	}{
		{
			payload: []Config{
				{
					File:   "nginx.conf",
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
*/

func TestParsing(t *testing.T) {
	var tests = []struct {
		title    string
		payload  []Config
		expected ParsingError
	}{
		{
			"simple",
			[]Config{
				{
					File:   "nginx.conf",
					Status: "ok",
					Parsed: []Block{
						Block{
							Directive: "events",
							Line:      1,
							Block: []Block{
								Block{
									Directive: "worker_connections",
									Line:      2,
									Args:      []string{"1024"},
								},
							},
						},
					},
				},
			},
			"",
		},
	}

	for _, test := range tests {
		t.Log(test.title)
		err := Parsing(test.payload)
		if err != test.expected {
			t.Errorf("Error: \t\nexpected: %v, \t\nactual: %v", test.expected, err)
		}
	}
}
