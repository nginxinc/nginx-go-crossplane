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

//nolint:exhaustruct,funlen
func TestAnalyze_njs(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"js_import in http location context ok": {
			&Directive{
				Directive: "js_import",
				Args:      []string{"http.js"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			false,
		},
		"js_import in stream server context ok": {
			&Directive{
				Directive: "js_import",
				Args:      []string{"http.js"},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			false,
		},
		"js_import not ok": {
			&Directive{
				Directive: "js_import",
				Args:      []string{"http.js"},
				Line:      5,
			},
			blockCtx{"http", "location", "if"},
			true,
		},
		"js_content in location if context ok": {
			&Directive{
				Directive: "js_content",
				Args:      []string{"function"},
				Line:      5,
			},
			blockCtx{"http", "location", "if"},
			false,
		},
		"js_content not ok": {
			&Directive{
				Directive: "js_content",
				Args:      []string{"function"},
				Line:      5,
			},
			blockCtx{"http", "server", "if"},
			true,
		},
		"js_periodic ok in http": {
			&Directive{
				Directive: "js_periodic",
				Args:      []string{"function"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			false,
		},
		"js_periodic not ok in http if": {
			&Directive{
				Directive: "js_periodic",
				Args:      []string{"function"},
				Line:      5,
			},
			blockCtx{"http", "location", "if"},
			true,
		},
		"js_periodic ok in stream": {
			&Directive{
				Directive: "js_periodic",
				Args:      []string{"function"},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			false,
		},
		"js_periodic not ok in stream": {
			&Directive{
				Directive: "js_periodic",
				Args:      []string{"function"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"js_shared_dict_zone in http context ok": {
			&Directive{
				Directive: "js_shared_dict_zone",
				Args:      []string{"zone=foo:1M"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"js_shared_dict_zone in stream context ok": {
			&Directive{
				Directive: "js_shared_dict_zone",
				Args:      []string{"zone=foo:1M"},
				Line:      5,
			},
			blockCtx{"stream"},
			false,
		},
		"js_shared_dict_zone not ok": {
			&Directive{
				Directive: "js_shared_dict_zone",
				Args:      []string{"zone=foo:1M"},
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

//nolint:funlen
func TestAnalyze_zone_sync(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"zone_sync in stream server context ok": {
			&Directive{
				Directive: "zone_sync",
				Args:      []string{},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			false,
		},
		"zone_sync invalid context": {
			&Directive{
				Directive: "zone_sync",
				Args:      []string{},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"zone_sync invalid args": {
			&Directive{
				Directive: "zone_sync",
				Args:      []string{"invalid"},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			true,
		},
		"zone_sync_ssl in stream context ok": {
			&Directive{
				Directive: "zone_sync_ssl",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"stream"},
			false,
		},
		"zone_sync_ssl in stream server context ok": {
			&Directive{
				Directive: "zone_sync_ssl",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			false,
		},
		"zone_sync_ssl invalid context": {
			&Directive{
				Directive: "zone_sync_ssl",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"http"},
			true,
		},
		"zone_sync_ssl invalid args": {
			&Directive{
				Directive: "zone_sync_ssl",
				Args:      []string{"invalid"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"zone_sync_ssl_conf_command in stream context ok": {
			&Directive{
				Directive: "zone_sync_ssl_conf_command",
				Args:      []string{"somename", "somevalue"},
				Line:      5,
			},
			blockCtx{"stream"},
			false,
		},
		"zone_sync_ssl_conf_command in stream server context ok": {
			&Directive{
				Directive: "zone_sync_ssl_conf_command",
				Args:      []string{"somename", "somevalue"},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			false,
		},
		"zone_sync_ssl_conf_command invalid context": {
			&Directive{
				Directive: "zone_sync_ssl_conf_command",
				Args:      []string{"somename", "somevalue"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"zone_sync_ssl_conf_command missing one arg": {
			&Directive{
				Directive: "zone_sync_ssl_conf_command",
				Args:      []string{"somename"},
				Line:      5,
			},
			blockCtx{"stream", "server"},
			true,
		},
		"zone_sync_ssl_conf_command missing both args": {
			&Directive{
				Directive: "zone_sync_ssl_conf_command",
				Args:      []string{},
				Line:      5,
			},
			blockCtx{"stream", "server"},
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

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv4},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
func TestAnalyze_http3(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"http3 ok": {
			&Directive{
				Directive: "http3",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"http3 not ok": {
			&Directive{
				Directive: "http3",
				Args:      []string{"somevalue"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"http3_hq ok": {
			&Directive{
				Directive: "http3_hq",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"http3_hq not ok": {
			&Directive{
				Directive: "http3_hq",
				Args:      []string{"somevalue"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"http3_max_concurrent_streams ok": {
			&Directive{
				Directive: "http3_max_concurrent_streams",
				Args:      []string{"10"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"http3_max_concurrent_streams not ok": {
			&Directive{
				Directive: "http3_max_concurrent_streams",
				Args:      []string{"10"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"http3_stream_buffer_size ok": {
			&Directive{
				Directive: "http3_stream_buffer_size",
				Args:      []string{"128k"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"http3_stream_buffer_size not ok": {
			&Directive{
				Directive: "http3_stream_buffer_size",
				Args:      []string{"128k"},
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

//nolint:funlen
func TestAnalyze_quic(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"quic_active_connection_id_limit ok": {
			&Directive{
				Directive: "quic_active_connection_id_limit",
				Args:      []string{"2"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"quic_active_connection_id_limit not ok": {
			&Directive{
				Directive: "quic_active_connection_id_limit",
				Args:      []string{"2"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"quic_bpf ok": {
			&Directive{
				Directive: "quic_bpf",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"main"},
			false,
		},
		"quic_bpf not ok": {
			&Directive{
				Directive: "quic_bpf",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"quic_gso ok": {
			&Directive{
				Directive: "quic_gso",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"quic_gso not ok": {
			&Directive{
				Directive: "quic_gso",
				Args:      []string{"somevalue"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"quic_host_key ok": {
			&Directive{
				Directive: "http3_max_concurrent_streams",
				Args:      []string{"somefile"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"quic_retry ok": {
			&Directive{
				Directive: "quic_retry",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			false,
		},
		"quic_retry not ok": {
			&Directive{
				Directive: "quic_retry",
				Args:      []string{"somevalue"},
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

//nolint:funlen,maintidx
func TestAnalyze_nap_app_protect_waf_v5(t *testing.T) {
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
		"app_protect_enforcer_address ok http": {
			&Directive{
				Directive: "app_protect_enforcer_address",
				Args:      []string{"127.0.0.1:50000"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_enforcer_address not ok stream": {
			&Directive{
				Directive: "app_protect_enforcer_address",
				Args:      []string{"127.0.0.1:50000"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_enforcer_address not ok http server": {
			&Directive{
				Directive: "app_protect_enforcer_address",
				Args:      []string{"127.0.0.1:50000"},
				Line:      5,
			},
			blockCtx{"http", "server"},
			true,
		},
		"app_protect_enforcer_address not ok extra parameters": {
			&Directive{
				Directive: "app_protect_enforcer_address",
				Args:      []string{"127.0.0.1:50000", "foo"},
				Line:      5,
			},
			blockCtx{"http"},
			true,
		},
		"app_protect_custom_log_attribute ok http": {
			&Directive{
				Directive: "app_protect_custom_log_attribute",
				Args:      []string{"environment", "env1"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"app_protect_custom_log_attribute ok http location": {
			&Directive{
				Directive: "app_protect_custom_log_attribute",
				Args:      []string{"environment", "env1"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			false,
		},
		"app_protect_custom_log_attribute not ok stream": {
			&Directive{
				Directive: "app_protect_custom_log_attribute",
				Args:      []string{"environment", "env1"},
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"app_protect_security_log_enable not ok too few paramseters": {
			&Directive{
				Directive: "app_protect_custom_log_attribute",
				Args:      []string{"environment"},
				Line:      5,
			},
			blockCtx{"http", "location"},
			true,
		},
		"app_protect_custom_log_attribute not ok extra parameters": {
			&Directive{
				Directive: "app_protect_custom_log_attribute",
				Args:      []string{"environment", "env1", "env2"},
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchAppProtectWAFv5},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
func TestAnalyze_lua(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"content_by_lua_file ok": {
			&Directive{
				Directive: "content_by_lua_file",
				Args:      []string{"/path/to/lua/app/root/$path.lua"},
				Line:      5,
			},
			blockCtx{"http", "location", "location if"},
			false,
		},
		"content_by_lua_file relative path ok": {
			&Directive{
				Directive: "content_by_lua_file",
				Args:      []string{"foo/bar.lua"},
				Line:      5,
			},
			blockCtx{"http", "location", "location if"},
			false,
		},
		"content_by_lua_file nor ok": {
			&Directive{
				Directive: "content_by_lua_file",
				Args:      []string{"foo/bar.lua"},
				Line:      5,
			},
			blockCtx{"server"},
			false,
		},
		"lua_shared_dict ok": {
			&Directive{
				Directive: "lua_shared_dict",
				Args:      []string{"dogs", "10m"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"lua_shared_dict not ok": {
			&Directive{
				Directive: "lua_shared_dict",
				Args:      []string{"10m"},
				Line:      5,
			},
			blockCtx{"http"},
			true,
		},
		"lua_sa_restart ok": {
			&Directive{
				Directive: "lua_sa_restart",
				Args:      []string{"off"},
				Line:      5,
			},
			blockCtx{"http"},
			false,
		},
		"lua_sa_restart not ok": {
			&Directive{
				Directive: "lua_sa_restart",
				Args:      []string{"something"},
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
			err := analyze("nginx.conf", tc.stmt, ";", tc.ctx, &ParseOptions{
				MatchFuncs: []MatchFunc{MatchLua},
			})

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

//nolint:funlen
func TestAnalyze_mgmt(t *testing.T) {
	t.Parallel()
	testcases := map[string]struct {
		stmt    *Directive
		ctx     blockCtx
		wantErr bool
	}{
		"connect_timeout in mgmt context ok": {
			&Directive{
				Directive: "connect_timeout",
				Args:      []string{"15s"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"mgmt in main context ok": {
			&Directive{
				Directive: "mgmt",
				Line:      5,
			},
			blockCtx{"main"},
			false,
		},
		"read_timeout in mgmt context ok": {
			&Directive{
				Directive: "read_timeout",
				Args:      []string{"60s"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"resolver in mgmt context ok": {
			&Directive{
				Directive: "resolver",
				Args:      []string{"127.0.0.53:53", "valid=100s"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"resolver_timeout in mgmt context ok": {
			&Directive{
				Directive: "resolver_timeout",
				Args:      []string{"30s"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"send_timeout in mgmt context ok": {
			&Directive{
				Directive: "send_timeout",
				Args:      []string{"60s"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl in mgmt context ok": {
			&Directive{
				Directive: "ssl",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_certificate in mgmt context ok": {
			&Directive{
				Directive: "ssl_certificate",
				Args:      []string{"/etc/nginx/foo.pem"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_certificate_key in mgmt context ok": {
			&Directive{
				Directive: "ssl_certificate_key",
				Args:      []string{"/etc/nginx/foo.pem"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_ciphers in mgmt context ok": {
			&Directive{
				Directive: "ssl_ciphers",
				Args:      []string{"DEFAULT"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_crl in mgmt context ok": {
			&Directive{
				Directive: "ssl_crl",
				Args:      []string{"/etc/nginx/foo.pem"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_name in mgmt context ok": {
			&Directive{
				Directive: "ssl_name",
				Args:      []string{"15s"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_password_file in mgmt context ok": {
			&Directive{
				Directive: "ssl_password_file",
				Args:      []string{"/etc/nginx/foo.pem"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_protocols in mgmt context ok": {
			&Directive{
				Directive: "ssl_protocols",
				Args:      []string{"TLSv1 TLSv1.1 TLSv1.2 TLSv1.3"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_server_name in mgmt context ok": {
			&Directive{
				Directive: "ssl_server_name",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_trusted_certificate in mgmt context ok": {
			&Directive{
				Directive: "ssl_trusted_certificate",
				Args:      []string{"/etc/nginx/foo.pem"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_verify in mgmt context ok": {
			&Directive{
				Directive: "ssl_verify",
				Args:      []string{"on"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"ssl_verify_depth in mgmt context ok": {
			&Directive{
				Directive: "ssl_verify_depth",
				Args:      []string{"1"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"usage_report in mgmt context ok": {
			&Directive{
				Directive: "usage_report",
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"uuid_file in mgmt context ok": {
			&Directive{
				Directive: "uuid_file",
				Args:      []string{"logs/uuid"},
				Line:      5,
			},
			blockCtx{"mgmt"},
			false,
		},
		"usage_report not in mgmt context not ok": {
			&Directive{
				Directive: "usage_report",
				Line:      5,
			},
			blockCtx{"stream"},
			true,
		},
		"ssl_protocols in mgmt context ok but not enough arguments": {
			&Directive{
				Directive: "ssl_protocols",
				Line:      5,
			},
			blockCtx{"mgmt"},
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
