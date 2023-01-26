/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

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
			false,
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

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestAnalyze_auth_jwt_require(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"auth_jwt_require ok": {
			&Directive{
				Directive: "auth_jwt_require",
				Args:      []string{"$value1", "$value2"},
				Line:      5,
			},
			blockCtx{"http", "location", "limit_except"},
			false,
		},
		"auth_jwt_require with error code ok": {
			&Directive{
				Directive: "auth_jwt_require",
				Args:      []string{"$value1", "$value2", "error=403"},
				Line:      5,
			},
			blockCtx{"http", "location", "limit_except"},
			false,
		},
		"auth_jwt_require not ok": {
			&Directive{
				Directive: "auth_jwt_require",
				Args:      []string{"$value"},
				Line:      5,
			},
			blockCtx{"stream"},
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

func TestAnalyze_nap_app_protect_enable(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_enable ok http": {
			&Directive{
				Directive: "app_protect_enable",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_enable not ok stream": {
			&Directive{
				Directive: "app_protect_enable",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_enable not ok true": {
			&Directive{
				Directive: "app_protect_enable",
				Args:      []string{"true"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_enable not ok extra parameters": {
			&Directive{
				Directive: "app_protect_enable",
				Args:      []string{"on", "off"},
				Line:      5,
			},
			blockCtx{"http", "server"},
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

func TestAnalyze_nap_app_protect_security_log_enable(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_security_log_enable ok http": {
			&Directive{
				Directive: "app_protect_security_log_enable",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_security_log_enable not ok stream": {
			&Directive{
				Directive: "app_protect_security_log_enable",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_security_log_enable not ok false": {
			&Directive{
				Directive: "app_protect_security_log_enable",
				Args:      []string{"false"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_security_log_enable not ok extra parameters": {
			&Directive{
				Directive: "app_protect_security_log_enable",
				Args:      []string{"on", "off"},
				Line:      5,
			},
			blockCtx{"http", "server"},
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

func TestAnalyze_nap_app_protect_security_log(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_security_log ok http": {
			&Directive{
				Directive: "app_protect_security_log",
				Args:      []string{"/etc/app_protect/nap_log_format.json", "syslog:localhost:522"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_security_log not ok stream": {
			&Directive{
				Directive: "app_protect_security_log",
				Args:      []string{"/etc/app_protect/nap_log_format.json", "syslog:localhost:522"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_security_log not ok extra parameters": {
			&Directive{
				Directive: "app_protect_security_log",
				Args:      []string{"/etc/app_protect/nap_log_format.json", "syslog:localhost:522", "true"},
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

func TestAnalyze_nap_app_protect_policy_file(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_policy_file ok http": {
			&Directive{
				Directive: "app_protect_policy_file",
				Args:      []string{"/etc/app_protect/nap_policy.json"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_policy_file not ok stream": {
			&Directive{
				Directive: "app_protect_policy_file",
				Args:      []string{"/etc/app_protect/nap_policy.json"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_policy_file not ok extra parameters": {
			&Directive{
				Directive: "app_protect_policy_file",
				Args:      []string{"/etc/app_protect/nap_policy.json", "/etc/app_protect/nap_policy2.json"},
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

func TestAnalyze_nap_app_protect_physical_memory_util_thresholds(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_physical_memory_util_thresholds ok http": {
			&Directive{
				Directive: "app_protect_physical_memory_util_thresholds",
				Args:      []string{"high=100", "low=10"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_physical_memory_util_thresholds not ok stream": {
			&Directive{
				Directive: "app_protect_physical_memory_util_thresholds",
				Args:      []string{"high=100", "low=10"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_physical_memory_util_thresholds not ok http location": {
			&Directive{
				Directive: "app_protect_physical_memory_util_thresholds",
				Args:      []string{"high=100", "low=10"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_physical_memory_util_thresholds not ok extra parameters": {
			&Directive{
				Directive: "app_protect_physical_memory_util_thresholds",
				Args:      []string{"high=100", "low=10", "true"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_cpu_thresholds(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_cpu_thresholds ok http": {
			&Directive{
				Directive: "app_protect_cpu_thresholds",
				Args:      []string{"high=100", "low=10"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_cpu_thresholds not ok stream": {
			&Directive{
				Directive: "app_protect_cpu_thresholds",
				Args:      []string{"high=100", "low=10"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_cpu_thresholds not ok http server": {
			&Directive{
				Directive: "app_protect_cpu_thresholds",
				Args:      []string{"high=100", "low=10"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"app_protect_cpu_thresholds not ok extra parameters": {
			&Directive{
				Directive: "app_protect_cpu_thresholds",
				Args:      []string{"high=100", "low=10", "true"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_failure_mode_action(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_failure_mode_action ok http": {
			&Directive{
				Directive: "app_protect_failure_mode_action",
				Args:      []string{"pass"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_failure_mode_action not ok stream": {
			&Directive{
				Directive: "app_protect_failure_mode_action",
				Args:      []string{"drop"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_failure_mode_action not ok http server": {
			&Directive{
				Directive: "app_protect_failure_mode_action",
				Args:      []string{"pass"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"app_protect_failure_mode_action not ok extra parameters": {
			&Directive{
				Directive: "app_protect_failure_mode_action",
				Args:      []string{"pass", "on"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_cookie_seed(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_cookie_seed ok http": {
			&Directive{
				Directive: "app_protect_cookie_seed",
				Args:      []string{"jkldsf90upiokasdj120"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_cookie_seed not ok stream": {
			&Directive{
				Directive: "app_protect_cookie_seed",
				Args:      []string{"jkldsf90upiokasdj120"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_cookie_seed not ok http location": {
			&Directive{
				Directive: "app_protect_cookie_seed",
				Args:      []string{"jkldsf90upiokasdj120"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_cookie_seed not ok extra parameters": {
			&Directive{
				Directive: "app_protect_cookie_seed",
				Args:      []string{"jkldsf90upiokasdj120", "on"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_compressed_requests_action(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_compressed_requests_action ok http": {
			&Directive{
				Directive: "app_protect_compressed_requests_action",
				Args:      []string{"drop"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_compressed_requests_action not ok stream": {
			&Directive{
				Directive: "app_protect_compressed_requests_action",
				Args:      []string{"pass"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_compressed_requests_action not ok http location": {
			&Directive{
				Directive: "app_protect_compressed_requests_action",
				Args:      []string{"pass"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_compressed_requests_action not ok extra parameters": {
			&Directive{
				Directive: "app_protect_compressed_requests_action",
				Args:      []string{"pass", "on"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_request_buffer_overflow_action(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_request_buffer_overflow_action ok http": {
			&Directive{
				Directive: "app_protect_request_buffer_overflow_action",
				Args:      []string{"pass"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_request_buffer_overflow_action not ok stream": {
			&Directive{
				Directive: "app_protect_request_buffer_overflow_action",
				Args:      []string{"drop"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_request_buffer_overflow_action not ok http server": {
			&Directive{
				Directive: "app_protect_request_buffer_overflow_action",
				Args:      []string{"drop"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"app_protect_request_buffer_overflow_action not ok extra parameters": {
			&Directive{
				Directive: "app_protect_request_buffer_overflow_action",
				Args:      []string{"drop", "on"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_user_defined_signatures(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_user_defined_signatures ok http": {
			&Directive{
				Directive: "app_protect_user_defined_signatures",
				Args:      []string{"app_protect_user_defined_signature_def_01"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_user_defined_signatures not ok stream": {
			&Directive{
				Directive: "app_protect_user_defined_signatures",
				Args:      []string{"app_protect_user_defined_signature_def_01"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_user_defined_signatures not ok http location": {
			&Directive{
				Directive: "app_protect_user_defined_signatures",
				Args:      []string{"app_protect_user_defined_signature_def_01"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_user_defined_signatures not ok extra parameters": {
			&Directive{
				Directive: "app_protect_user_defined_signatures",
				Args:      []string{"app_protect_user_defined_signature_def_01", "on"},
				Line:      5,
			},
			blockCtx{"http"},
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

func TestAnalyze_nap_app_protect_reconnect_period_seconds(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"app_protect_reconnect_period_seconds ok http": {
			&Directive{
				Directive: "app_protect_reconnect_period_seconds",
				Args:      []string{"10"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_reconnect_period_seconds not ok stream": {
			&Directive{
				Directive: "app_protect_reconnect_period_seconds",
				Args:      []string{"10"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_reconnect_period_seconds not ok http server": {
			&Directive{
				Directive: "app_protect_reconnect_period_seconds",
				Args:      []string{"10"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"app_protect_reconnect_period_seconds not ok extra parameters": {
			&Directive{
				Directive: "app_protect_reconnect_period_seconds",
				Args:      []string{"10", "20"},
				Line:      5,
			},
			blockCtx{"http"},
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
