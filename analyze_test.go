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
				} else if e, ok := err.(*ParseError); !ok {
					t.Fatalf("error was not a ParseError: %v", err)
				} else if !strings.HasSuffix(e.What, `directive is not allowed here`) {
					t.Fatalf("unexpected error message: %q", e.What)
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
			blockCtx{"http", "location", "limit_except"},
			true,
		},
		"auth_jwt not ok": {
			&Directive{
				Directive: "auth_jwt",
				Args:      []string{"closed site", "token=$cookie_auth_token"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"auth_jwt_key_file": {
			&Directive{
				Directive: "auth_jwt_key_file",
				Args:      []string{"some/weird/file"},
				Line:      5,
			},
			blockCtx{"http", "location", "limit_except"},
			false,
		},
		"auth_jwt_key_request": {
			&Directive{
				Directive: "auth_jwt_key_request",
				Args:      []string{"http://some.weird.uri"},
				Line:      5,
			},
			blockCtx{"http", "location", "limit_except"},
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
}

func TestAnalyze_njs(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"js_import ok": {
			&Directive{
				Directive: "js_import",
				Args:      []string{"http.js"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"js_import not ok": {
			&Directive{
				Directive: "js_import",
				Args:      []string{"http.js"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestAnalyze_stream_resolver(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"resolver ok stream": {
			&Directive{
				Directive: "resolver",
				Args:      []string{"127.0.0.53:53", "valid=100s"},
				Line:      5,
			},
			blockCtx{"stream", "upstream"},
			false,
		},
		"resolver ok http": {
			&Directive{
				Directive: "resolver",
				Args:      []string{"127.0.0.53:53", "valid=100s"},
				Line:      5,
			},
			blockCtx{"http", "upstream"},
			false,
		},
		"resolver_timeout stream": {
			&Directive{
				Directive: "resolver_timeout",
				Args:      []string{"10s"},
				Line:      5,
			},
			blockCtx{"stream", "upstream"},
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

			if !tc.wantErr && err != nil {
				t.Fatal("expected nil, got error")
			}
		})
	}

	stmt := &Directive{
		Directive: "resolver",
		Args:      []string{"127.0.0.53:53", "valid=100s"},
		Line:      5,
	}
	ctx := blockCtx{"stream", "upstream"}
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
			} else if e, ok := err.(*ParseError); !ok {
				t.Fatalf("error was not a ParseError: %v", err)
			} else if !strings.HasSuffix(e.What, `it must be "on" or "off"`) {
				t.Fatalf("unexpected error message: %q", e.What)
			}
		}
	})
}
