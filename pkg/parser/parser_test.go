package parser

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParsingSimple(t *testing.T) {

	dir, err := filepath.Abs("configs/simple/nginx.conf")
	if err != nil {
		t.Error(err)
	}
	// making the test data
	b := Block{directive: "worker_connections", line: 2, args: []string{"1024"}}
	b1 := Block{directive: "return", line: 10, args: []string{"200", "foo bar baz"}}
	b2 := Block{directive: "location", line: 9, args: []string{"/"}, block: []Block{b1}}
	b3 := Block{directive: "server", line: 6, args: []string{}, block: []Block{b2}}
	b4 := Block{directive: "http", line: 5, args: []string{}, block: []Block{b3}}
	b5 := Block{directive: "events", line: 1, args: []string{}, block: []Block{b}}
	c := Config{file: dir, status: "ok", errors: []Errors{}, parsed: []Block{b4, b5}}
	e := Payload{status: "ok", errors: []string{}, conf: c}

	var d1, d2 interface{}
	payload, err := Parsing(dir)
	if err != nil {
		t.Error(err)
	}
	expectedData, err := json.Marshal(e)
	if err := json.Unmarshal(payload, &d1); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(expectedData, &d2); err != nil {
		t.Error(err)
	}

	if res := reflect.DeepEqual(d2, d1); !res {
		// they're not equal
		t.Errorf("expected data does not match test json generated")
	}

	// not equal

	/*var tests = []struct {
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
	}*/

}
