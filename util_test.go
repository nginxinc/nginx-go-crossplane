package crossplane_test

import (
	"encoding/json"
	"testing"

	. "gitlab.com/f5/nginx/crossplane-go"
)

//nolint:funlen
func TestPayload(t *testing.T) {
	t.Parallel()
	t.Run("combine", func(t *testing.T) {
		t.Parallel()
		payload := Payload{
			Config: []Config{
				{
					File: "example1.conf",
					Parsed: Directives{
						{
							Directive: "include",
							Args:      []string{"example2.conf"},
							Line:      1,
							Includes:  []int{1},
						},
					},
				},
				{
					File: "example2.conf",
					Parsed: Directives{
						{
							Directive: "events",
							Args:      []string{},
							Line:      1,
						},
						{
							Directive: "http",
							Args:      []string{},
							Line:      2,
						},
					},
				},
			},
		}
		expected := Payload{
			Status: "ok",
			Errors: []PayloadError{},
			Config: []Config{
				{
					File:   "example1.conf",
					Status: "ok",
					Errors: []ConfigError{},
					Parsed: Directives{
						{
							Directive: "events",
							Args:      []string{},
							Line:      1,
						},
						{
							Directive: "http",
							Args:      []string{},
							Line:      2,
						},
					},
				},
			},
		}
		combined, err := payload.Combined()
		if err != nil {
			t.Fatal(err)
		}
		b1, _ := json.Marshal(expected)
		b2, _ := json.Marshal(*combined)
		if string(b1) != string(b2) {
			t.Fatalf("expected: %s\nbut got: %s", b1, b2)
		}
	})
}
