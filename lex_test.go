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
			i := 0
			for token := range Lex(file) {
				expected := fixture.tokens[i]
				if token.Value != expected.value || token.Line != expected.line {
					t.Fatalf("expected (%q,%d) but got (%q,%d)", expected.value, expected.line, token.Value, token.Line)
				}
				i++
			}
		})
	}
}

var lexToken NgxToken //nolint: gochecknoglobals // trying to avoid return value being optimzed away

func BenchmarkLex(b *testing.B) {
	var t NgxToken

	for _, bm := range lexFixtures {
		b.Run(bm.name, func(b *testing.B) {
			path := getTestConfigPath(bm.name, "nginx.conf")
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

				for tok := range Lex(file) {
					t = tok
				}
			}
		})
	}

	lexToken = t
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
