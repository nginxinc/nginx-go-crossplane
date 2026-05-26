/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"os"
	"strings"
	"testing"
)

type tokenLine struct {
	value string
	line  int
}

type lexFixture struct {
	name   string
	tokens []tokenLine
}

//nolint:gochecknoglobals
var lexFixtures = []lexFixture{
	{"simple", []tokenLine{
		{"events", 1},
		{"{", 1},
		{"worker_connections", 2},
		{"1024", 2},
		{";", 2},
		{"}", 3},
		{"http", 5},
		{"{", 5},
		{"server", 6},
		{"{", 6},
		{"listen", 7},
		{"127.0.0.1:8080", 7},
		{";", 7},
		{"server_name", 8},
		{"default_server", 8},
		{";", 8},
		{"location", 9},
		{"/", 9},
		{"{", 9},
		{"return", 10},
		{"200", 10},
		{"foo bar baz", 10},
		{";", 10},
		{"}", 11},
		{"}", 12},
		{"}", 13},
	}},
	{"with-comments", []tokenLine{
		{"events", 1},
		{"{", 1},
		{"worker_connections", 2},
		{"1024", 2},
		{";", 2},
		{"}", 3},
		{"#comment", 4},
		{"http", 5},
		{"{", 5},
		{"server", 6},
		{"{", 6},
		{"listen", 7},
		{"127.0.0.1:8080", 7},
		{";", 7},
		{"#listen", 7},
		{"server_name", 8},
		{"default_server", 8},
		{";", 8},
		{"location", 9},
		{"/", 9},
		{"{", 9},
		{"## this is brace", 9},
		{"# location /", 10},
		{"return", 11},
		{"200", 11},
		{"foo bar baz", 11},
		{";", 11},
		{"}", 12},
		{"}", 13},
		{"}", 14},
	}},
	{"messy", []tokenLine{
		{"user", 1},
		{"nobody", 1},
		{";", 1},
		{"# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n ", 2},
		{"events", 3},
		{"{", 3},
		{"worker_connections", 3},
		{"2048", 3},
		{";", 3},
		{"}", 3},
		{"http", 5},
		{"{", 5},
		{"#forteen", 5},
		{"# this is a comment", 6},
		{"access_log", 7},
		{"off", 7},
		{";", 7},
		{"default_type", 7},
		{"text/plain", 7},
		{";", 7},
		{"error_log", 7},
		{"off", 7},
		{";", 7},
		{"server", 8},
		{"{", 8},
		{"listen", 9},
		{"8083", 9},
		{";", 9},
		{"return", 10},
		{"200", 10},
		{"Ser\" ' ' ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n", 10},
		{";", 10},
		{"}", 11},
		{"server", 12},
		{"{", 12},
		{"listen", 12},
		{"8080", 12},
		{";", 12},
		{"root", 13},
		{"/usr/share/nginx/html", 13},
		{";", 13},
		{"location", 14},
		{"~", 14},
		{"/hello/world;", 14},
		{"{", 14},
		{"return", 14},
		{"301", 14},
		{"/status.html", 14},
		{";", 14},
		{"}", 14},
		{"location", 15},
		{"/foo", 15},
		{"{", 15},
		{"}", 15},
		{"location", 15},
		{"/bar", 15},
		{"{", 15},
		{"}", 15},
		{"location", 16},
		{"/\\{\\;\\}\\ #\\ ab", 16},
		{"{", 16},
		{"}", 16},
		{"# hello", 16},
		{"if", 17},
		{"($request_method", 17},
		{"=", 17},
		{"P\\{O\\)\\###\\;ST", 17},
		{")", 17},
		{"{", 17},
		{"}", 17},
		{"location", 18},
		{"/status.html", 18},
		{"{", 18},
		{"try_files", 19},
		{"/abc/${uri}", 19},
		{"/abc/${uri}.html", 19},
		{"=404", 19},
		{";", 19},
		{"}", 20},
		{"location", 21},
		{"/sta;\n                    tus", 21},
		{"{", 22},
		{"return", 22},
		{"302", 22},
		{"/status.html", 22},
		{";", 22},
		{"}", 22},
		{"location", 23},
		{"/upstream_conf", 23},
		{"{", 23},
		{"return", 23},
		{"200", 23},
		{"/status.html", 23},
		{";", 23},
		{"}", 23},
		{"}", 23},
		{"server", 24},
		{"{", 25},
		{"}", 25},
		{"}", 25},
	}},
	{"quote-behavior", []tokenLine{
		{"outer-quote", 1},
		{"left", 1},
		{"-quote", 1},
		{"right-\"quote\"", 1},
		{"inner\"-\"quote", 1},
		{";", 1},
		{"", 2},
		{"", 2},
		{"left-empty", 2},
		{"right-empty\"\"", 2},
		{"inner\"\"empty", 2},
		{"right-empty-single\"", 2},
		{";", 2},
	}},
	{"quoted-right-brace", []tokenLine{
		{"events", 1},
		{"{", 1},
		{"}", 1},
		{"http", 2},
		{"{", 2},
		{"log_format", 3},
		{"main", 3},
		{"escape=json", 3},
		{"{ \"@timestamp\": \"$time_iso8601\", ", 4},
		{"\"server_name\": \"$server_name\", ", 5},
		{"\"host\": \"$host\", ", 6},
		{"\"status\": \"$status\", ", 7},
		{"\"request\": \"$request\", ", 8},
		{"\"uri\": \"$uri\", ", 9},
		{"\"args\": \"$args\", ", 10},
		{"\"https\": \"$https\", ", 11},
		{"\"request_method\": \"$request_method\", ", 12},
		{"\"referer\": \"$http_referer\", ", 13},
		{"\"agent\": \"$http_user_agent\"", 14},
		{"}", 15},
		{";", 15},
		{"}", 16},
	}},
	{"lua-basic", []tokenLine{
		{"http", 1},
		{"{", 1},
		{"init_by_lua", 2},
		{"\n        print(\"I need no extra escaping here, for example: \\r\\nblah\")\n    ", 2},
		{";", 4},
		{"lua_shared_dict", 5},
		{"dogs", 5},
		{"1m", 5},
		{";", 5},
		{"server", 6},
		{"{", 6},
		{"listen", 7},
		{"8080", 7},
		{";", 7},
		{"location", 8},
		{"/", 8},
		{"{", 8},
		{"set_by_lua", 9},
		{"$res", 9},
		{" return 32 + math.cos(32) ", 9},
		{";", 9},
		{"access_by_lua_file", 10},
		{"/path/to/lua/access.lua", 10},
		{";", 10},
		{"}", 11},
		{"}", 12},
		{"}", 13},
	}},
	{"lua-block-simple", []tokenLine{
		{"http", 1},
		{"{", 1},
		{"init_by_lua_block", 2},
		{"\n        print(\"Lua block code with curly brace str {\")\n    ", 4},
		{";", 4},
		{"init_worker_by_lua_block", 5},
		{"\n        print(\"Work that every worker\")\n    ", 7},
		{";", 7},
		{"body_filter_by_lua_block", 8},
		{"\n        local data, eof = ngx.arg[1], ngx.arg[2]\n    ", 10},
		{";", 10},
		{"header_filter_by_lua_block", 11},
		{"\n        ngx.header[\"content-length\"] = nil\n    ", 13},
		{";", 13},
		{"server", 14},
		{"{", 14},
		{"listen", 15},
		{"127.0.0.1:8080", 15},
		{";", 15},
		{"location", 16},
		{"/", 16},
		{"{", 16},
		{"content_by_lua_block", 17},
		{"\n                ngx.say(\"I need no extra escaping here, for example: \\r\\nblah\")\n            ", 19},
		{";", 19},
		{"return", 20},
		{"200", 20},
		{"foo bar baz", 20},
		{";", 20},
		{"}", 21},
		{"ssl_certificate_by_lua_block", 22},
		{"\n            print(\"About to initiate a new SSL handshake!\")\n        ", 24},
		{";", 24},
		{"log_by_lua_block", 25},
		{"\n            print(\"I need no extra escaping here, for example: \\r\\nblah\")\n        ", 27},
		{";", 27},
		{"location", 28},
		{"/a", 28},
		{"{", 28},
		{"client_max_body_size", 29},
		{"100k", 29},
		{";", 29},
		{"client_body_buffer_size", 30},
		{"100k", 30},
		{";", 30},
		{"}", 31},
		{"}", 32},
		{"upstream", 34},
		{"foo", 34},
		{"{", 34},
		{"server", 35},
		{"127.0.0.1", 35},
		{";", 35},
		{"balancer_by_lua_block", 36},
		{"\n            -- use Lua to do something interesting here\n        ", 38},
		{";", 38},
		{"}", 39},
		{"}", 40},
	}},
	{"lua-block-larger", []tokenLine{
		{"http", 1},
		{"{", 1},
		{"access_by_lua_block", 2},
		{
			"\n        -- check the client IP address is in our black list" +
				"\n        if ngx.var.remote_addr == \"132.5.72.3\" then" +
				"\n            ngx.exit(ngx.HTTP_FORBIDDEN)" +
				"\n        end" +
				"\n" +
				"\n        -- check if the URI contains bad words" +
				"\n        if ngx.var.uri and" +
				"\n               string.match(ngx.var.request_body, \"evil\")" +
				"\n        then" +
				"\n            return ngx.redirect(\"/terms_of_use.html\")" +
				"\n        end" +
				"\n" +
				"\n        -- tests passed" +
				"\n    ", 16,
		},
		{";", 16},
		{"server", 17},
		{"{", 17},
		{"listen", 18},
		{"127.0.0.1:8080", 18},
		{";", 18},
		{"location", 19},
		{"/", 19},
		{"{", 19},
		{"content_by_lua_block", 20},
		{
			"\n                ngx.req.read_body()  -- explicitly read the req body" +
				"\n                local data = ngx.req.get_body_data()" +
				"\n                if data then" +
				"\n                    ngx.say(\"body data:\")" +
				"\n                    ngx.print(data)" +
				"\n                    return" +
				"\n                end" +
				"\n" +
				"\n                -- body may get buffered in a temp file:" +
				"\n                local file = ngx.req.get_body_file()" +
				"\n                if file then" +
				"\n                    ngx.say(\"body is in file \", file)" +
				"\n                else" +
				"\n                    ngx.say(\"no body found\")" +
				"\n                end" +
				"\n            ", 36,
		},
		{";", 36},
		{"return", 37},
		{"200", 37},
		{"foo bar baz", 37},
		{";", 37},
		{"}", 38},
		{"}", 39},
		{"}", 40},
	}},
	{"lua-block-tricky", []tokenLine{
		{"http", 1},
		{"{", 1},
		{"server", 2},
		{"{", 2},
		{"listen", 3},
		{"127.0.0.1:8080", 3},
		{";", 3},
		{"server_name", 4},
		{"content_by_lua_block", 4},
		{";", 4},
		{"# make sure this doesn't trip up lexers", 4},
		{"set_by_lua_block", 5},
		{"$res", 5},
		{
			" -- irregular lua block directive" +
				"\n            local a = 32" +
				"\n            local b = 56" +
				"\n" +
				"\n            ngx.var.diff = a - b;  -- write to $diff directly" +
				"\n            return a + b;          -- return the $sum value normally" +
				"\n        ",
			11,
		},
		{";", 11},
		{"rewrite_by_lua_block", 12},
		{
			" -- have valid braces in Lua code and quotes around directive" +
				"\n            do_something(\"hello, world!\\nhiya\\n\")" +
				"\n            a = { 1, 2, 3 }" +
				"\n            btn = iup.button({title=\"ok\"})" +
				"\n        ",
			16,
		},
		{";", 16},
		{"}", 17},
		{"upstream", 18},
		{"content_by_lua_block", 18},
		{"{", 18},
		{"# stuff", 19},
		{"}", 20},
		{"}", 21},
	}},
	{"comments-between-args", []tokenLine{
		{"http", 1},
		{"{", 1},
		{"#comment 1", 1},
		{"log_format", 2},
		{"#comment 2", 2},
		{"\\#arg\\ 1", 3},
		{"#comment 3", 3},
		{"#arg 2", 4},
		{"#comment 4", 4},
		{"#comment 5", 5},
		{";", 6},
		{"}", 7},
	}},
}

func TestLex(t *testing.T) {
	t.Parallel()
	for _, fixture := range lexFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path := getTestConfigPath(fixture.name, "nginx.conf")
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			options := LexOptions{
				Lexers: []RegisterLexer{lua.RegisterLexer()},
			}
			i := 0

			for token := range LexWithOptions(file, options) {
				expected := fixture.tokens[i]
				if token.Value != expected.value || token.Line != expected.line {
					t.Fatalf("expected (%q,%d) but got (%q,%d)", expected.value, expected.line, token.Value, token.Line)
				}
				i++
			}
		})
	}
}

func benchmarkLex(b *testing.B, path string, options LexOptions) {
	var t NgxToken

	file, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := file.Seek(0, 0); err != nil {
			b.Fatal(err)
		}

		for tok := range LexWithOptions(file, options) {
			t = tok
		}
	}

	_ = t
}

func BenchmarkLex(b *testing.B) {
	for _, bm := range lexFixtures {
		if strings.HasPrefix(bm.name, "lua") {
			continue
		}

		b.Run(bm.name, func(b *testing.B) {
			path := getTestConfigPath(bm.name, "nginx.conf")
			benchmarkLex(b, path, LexOptions{})
		})
	}
}

func BenchmarkLexWithLua(b *testing.B) {
	for _, bm := range lexFixtures {
		if !strings.HasPrefix(bm.name, "lua") {
			continue
		}

		b.Run(bm.name, func(b *testing.B) {
			path := getTestConfigPath(bm.name, "nginx.conf")
			benchmarkLex(b, path, LexOptions{Lexers: []RegisterLexer{lua.RegisterLexer()}})
		})
	}
}

//nolint:gochecknoglobals
var unhappyFixtures = map[string]string{
	"unbalanced open brance":                  `http {{}`,
	"unbalanced closing brace":                `http {}}`,
	"multiple open braces":                    `http {{server {}}`,
	"multiple closing braces after block end": `http {server {}}}`,
	"multiple semicolons":                     `server { listen 80;; }`,
	"semicolon afer closing brace":            `server { listen 80; };`,
	"open brace after semicolon":              `server { listen 80; {}`,
	"braces with no directive":                `http{}{}`,
	"missing final brace":                     `http{`,
}

func TestLex_unhappy(t *testing.T) {
	t.Parallel()

	for name, c := range unhappyFixtures {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var err error
			for t := range Lex(strings.NewReader(c)) {
				if t.Error != nil {
					err = t.Error
					break
				}
			}
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
