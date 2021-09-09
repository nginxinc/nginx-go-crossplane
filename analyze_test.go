package crossplane

import (
	"strings"
	"testing"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()
	fname := "/path/to/nginx.conf"

	// Checks that the `state` directive should only be in certain contexts.
	t.Run("state-directive", func(t *testing.T) {
		t.Parallel()
		stmt := &Directive{
			Directive: "state",
			Args:      []string{"/path/to/state/file.conf"},
			Line:      5, // this is arbitrary
		}

		// the state directive should not cause errors if it"s in these contexts
		goodCtxs := []blockCtx{
			{"http", "upstream"},
			{"stream", "upstream"},
			{"some_third_party_context"},
		}
		for _, ctx := range goodCtxs {
			if err := analyze(fname, stmt, ";", ctx, &ParseOptions{}); err != nil {
				t.Fatalf("expected err to be nil: %v", err)
			}
		}
		goodMap := map[string]bool{}
		for _, c := range goodCtxs {
			goodMap[c.key()] = true
		}

		for key := range contexts {
			// the state directive should only be in the "good" contexts
			if _, ok := goodMap[key]; !ok {
				actx := blockCtx(strings.Split(key, ">"))
				if err := analyze(fname, stmt, ";", actx, &ParseOptions{}); err == nil {
					t.Fatalf("expected error to not be nil: %v", err)
				} else if e, ok := err.(ParseError); !ok {
					t.Fatalf("error was not a ParseError: %v", err)
				} else if !strings.HasSuffix(e.what, `directive is not allowed here`) {
					t.Fatalf("unexpected error message: %q", e.what)
				}
			}
		}
	})
}

func TestAnalyze_auth_jwt(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"auth_jwt ok": {
			&Directive{
				Directive: "auth_jwt",
				Args:      []string{"closed site", "token=$cookie_auth_token"},
				Line:      5,
			},
			blockCtx{"http", "server", "location", "limit_except"},
			false,
		},
		"auth_jwt_key_file": {
			&Directive{
				Directive: "auth_jwt_key_file",
				Args:      []string{"some/weird/file"},
				Line:      5,
			},
			blockCtx{"http", "server", "location", "limit_except"},
			false,
		},
		"auth_jwt_key_request": {
			&Directive{
				Directive: "auth_jwt_key_request",
				Args:      []string{"http://some.weird.uri"},
				Line:      5,
			},
			blockCtx{"http", "server", "location", "limit_except"},
			false,
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{})

			if tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if !tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}

	stmt := &Directive{
		Directive: "auth_jwt",
		Args:      []string{"closed site", "token=$cookie_auth_token"},
		Line:      5,
	}
	ctx := blockCtx{"http", "server", "location", "limit_except"}
	if err := analyze("nginx.conf", stmt, ";", ctx, &ParseOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeFlagArgs(t *testing.T) {
	t.Parallel()
	fname := "/path/to/nginx.conf"

	// Check which arguments are valid for flag directives.
	t.Run("flag-args", func(t *testing.T) {
		t.Parallel()
		ctx := blockCtx{"events"}
		stmt := &Directive{
			Directive: "accept_mutex",
			Line:      2, // this is arbitrary
		}

		goodArgs := [][]string{{"on"}, {"off"}, {"On"}, {"Off"}, {"ON"}, {"OFF"}}
		for _, args := range goodArgs {
			stmt.Args = args
			if err := analyze(fname, stmt, ";", ctx, &ParseOptions{}); err != nil {
				t.Fatalf("expected err to be nil: %v", err)
			}
		}

		badArgs := [][]string{{"1"}, {"0"}, {"true"}, {"okay"}, {""}}
		for _, args := range badArgs {
			stmt.Args = args
			if err := analyze(fname, stmt, ";", ctx, &ParseOptions{}); err == nil {
				t.Fatalf("expected error to not be nil: %v", err)
			} else if e, ok := err.(ParseError); !ok {
				t.Fatalf("error was not a ParseError: %v", err)
			} else if !strings.HasSuffix(e.what, `it must be "on" or "off"`) {
				t.Fatalf("unexpected error message: %q", e.what)
			}
		}
	})
}
