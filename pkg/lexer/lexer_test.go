package lexer

import (
	"reflect"
	"testing"
)

func TestBalanceBraces(t *testing.T) {
	var testCases = []struct {
		title    string
		input    []LexicalItem
		expected UnbalancedBracesError
	}{
		{
			"balanced: super-simple",
			[]LexicalItem{
				{"{", 1}, {"}", 1},
			},
			"",
		},
		{
			"balanced: simple long",
			[]LexicalItem{
				{"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1},
				{"}", 1}, {"}", 1}, {"}", 1}, {"}", 1}, {"}", 1}, {"}", 1}, {"}", 1}, {"}", 1},
			},
			"",
		},
		{
			"balanced: with directives",
			[]LexicalItem{
				{"http", 1}, {"{", 1}, {"user", 1}, {"nginx", 1}, {";", 1}, {"}", 1},
			},
			"",
		},
		{
			"unbalanced: needle in a haystack",
			[]LexicalItem{
				{"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1}, {"{", 1},
				{"}", 2}, {"}", 2}, {"}", 2}, {"}", 2}, {"{", 2}, {"}", 2}, {"}", 2}, {"}", 2},
			},
			"UnbalancedBracesError: braces are not balanced",
		},
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		err := balanceBraces(tt.input)
		if err != tt.expected {
			t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected, err)
		}
	}
}

func TestLexBasicStuff(t *testing.T) {
	var lexBasicTests = []struct {
		title    string
		input    string
		expected []LexicalItem
	}{
		{
			"basic: one-line lexical analysis",
			"http { user nginx; }",
			[]LexicalItem{
				{"http", 1}, {"{", 1}, {"user", 1}, {"nginx", 1}, {";", 1}, {"}", 1},
			},
		},
		{
			"basic: multi-line lexical analysis",
			`http {
				user nginx;
			}`,
			[]LexicalItem{
				{"http", 1}, {"{", 1},
				{"user", 2}, {"nginx", 2}, {";", 2},
				{"}", 3},
			},
		},
		{
			"basic: comments",
			`http { # this is a comment
				user nginx;
				# this is another comment
			}`,
			[]LexicalItem{
				{"http", 1}, {"{", 1}, {"# this is a comment", 1},
				{"user", 2}, {"nginx", 2}, {";", 2},
				{"# this is another comment", 3},
				{"}", 4},
			},
		},
		{
			"basic: messy",
			`user nobody;
			# hello\n\\n\\\n worlddd  \#\\#\\\# dfsf\n \\n \\\n \
			"events" { "worker_connections" "2048"; }
			"http" {#forteen
				# this is a comment
				"access_log" off;default_type "text/plain"; error_log "off";
				server {
					"listen" "8083"            ;
					"return" 200 "Ser\" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n";
				}
				"server" {"listen" 8080;
					'root' /usr/share/nginx/html;
					location ~ "/hello/world;"{"return" 301 /status.html;}
					location /foo{}location /bar{}
					location /\{\;\}\ #\ ab {}# hello
					if ($request_method = P\{O\)\###\;ST  ){}
					location "/status.html" {
						try_files /abc/${uri} /abc/${uri}.html =404 ;
					}
					"location" "/sta;
								tus" {"return" 302 /status.html;}
					"location" /upstream_conf { "return" 200 /status.html; }}
				server
							{}}
			`,
			[]LexicalItem{
				{"user", 1}, {"nobody", 1}, {";", 1},
				{"# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n ", 2},
				{"events", 3}, {"{", 3}, {"worker_connections", 3}, {"2048", 3},
				{";", 3}, {"}", 3}, {"http", 5}, {"{", 5}, {"#forteen", 5},
				{"# this is a comment", 6}, {"access_log", 7}, {"off", 7}, {";", 7},
				{"default_type", 7}, {"text/plain", 7}, {";", 7}, {"error_log", 7},
				{"off", 7}, {";", 7}, {"server", 8}, {"{", 8}, {"listen", 9},
				{"8083", 9}, {";", 9}, {"return", 10}, {"200", 10},
				{"Ser\" \" \" ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n", 10},
				{";", 10}, {"}", 11}, {"server", 12}, {"{", 12}, {"listen", 12},
				{"8080", 12}, {";", 12}, {"root", 13}, {"/usr/share/nginx/html", 13},
				{";", 13}, {"location", 14}, {"~", 14}, {"/hello/world;", 14},
				{"{", 14}, {"return", 14}, {"301", 14}, {"/status.html", 14},
				{";", 14}, {"}", 14}, {"location", 15}, {"/foo", 15},
				{"{", 15}, {"}", 15}, {"location", 15}, {"/bar", 15},
				{"{", 15}, {"}", 15}, {"location", 16}, {"/\\{\\;\\}\\ #\\ ab", 16},
				{"{", 16}, {"}", 16}, {"# hello", 16}, {"if", 17},
				{"{$request_method", 17}, {"=", 17}, {"P\\{O\\}\\###\\;ST", 17},
				{"}", 17}, {"{", 17}, {"}", 17}, {"location", 18}, {"/status.html", 18},
				{"{", 18}, {"try_files", 19}, {"/abc/${uri} /abc/${uri}.html", 19},
				{"=404", 19}, {";", 19}, {"}", 20}, {"location", 21},
				{"/sta;\n                    tus", 21}, {"{", 22}, {"return", 22},
				{"302", 22}, {"/status.html", 22}, {";", 22}, {"}", 22},
				{"location", 23}, {"/upstream_conf", 23}, {"{", 23},
				{"return", 23}, {"200", 23}, {"/status.html", 23}, {";", 23},
				{"}", 23}, {"}", 23}, {"server", 24}, {"{", 25}, {"}", 25},
				{"}", 25},
			},
		},
	}
	for _, tt := range lexBasicTests {
		t.Log(tt.title)
		actual, err := Lex(tt.input)
		if err != nil {
			t.Errorf("Test failed due to: %v", err)
		}
		result := reflect.DeepEqual(tt.expected, actual)
		if !result {
			t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected, actual)
		}
	}
}
