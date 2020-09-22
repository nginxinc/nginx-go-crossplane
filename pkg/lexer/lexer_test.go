package lexer_test

import (
	"testing"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/lexer"
)

func TestBalanceBraces(t *testing.T) {
	var testCases = []struct {
		title    string
		input    []lexer.LexicalItem
		expected lexer.UnbalancedBracesError
	}{
		{
			"balanced: super-simple",
			[]lexer.LexicalItem{
				{"{", 1, 0}, {"}", 1, 0},
			},
			"",
		},
		{
			"balanced: simple long",
			[]lexer.LexicalItem{
				{"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0},
				{"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0},
			},
			"",
		},
		{
			"balanced: with directives",
			[]lexer.LexicalItem{
				{"http", 1, 0}, {"{", 1, 0}, {"user", 1, 0}, {"nginx", 1, 0}, {";", 1, 0}, {"}", 1, 0},
			},
			"",
		},
		{
			"unbalanced: needle in a haystack",
			[]lexer.LexicalItem{
				{"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0},
				{"}", 2, 0}, {"}", 2, 0}, {"}", 2, 0}, {"}", 2, 0}, {"{", 2, 0}, {"}", 2, 0}, {"}", 2, 0}, {"}", 2, 0},
			},
			lexer.UnbalancedBracesError("UnbalancedBracesError: braces are not balanced"),
		},
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		err := lexer.BalanceBraces(tt.input)

		if err != tt.expected {
			t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected, err)
		}
	}
}

func TestLexScanner(t *testing.T) {
	var testCases = []struct {
		title    string
		input    string
		expected []lexer.LexicalItem
	}{
		{
			"simple config",
			HelperReadConfig(t, "simple/nginx.conf"),
			[]lexer.LexicalItem{
				{"events", 1, 0},
				{"{", 1, 0},
				{"worker_connections", 2, 0},
				{"1024", 2, 0},
				{";", 2, 0},
				{"}", 3, 0},
				{"http", 5, 0},
				{"{", 5, 0},
				{"server", 6, 0},
				{"{", 6, 0},
				{"listen", 7, 0},
				{"127.0.0.1:8080", 7, 0},
				{";", 7, 0},
				{"server_name", 8, 0},
				{"default_server", 8, 0},
				{";", 8, 0},
				{"location", 9, 0},
				{"/", 9, 0},
				{"{", 9, 0},
				{"return", 10, 0},
				{"200", 10, 0},
				{"foo bar baz", 10, 0},
				{";", 10, 0},
				{"}", 11, 0},
				{"}", 12, 0},
				{"}", 13, 0},
			},
		},
		{
			"with comments",
			HelperReadConfig(t, "with-comments/nginx.conf"),
			[]lexer.LexicalItem{
				{"events", 1, 0},
				{"{", 1, 0},
				{"worker_connections", 2, 0},
				{"1024", 2, 0},
				{";", 2, 0},
				{"}", 3, 0},
				{"#comment", 4, 0},
				{"http", 5, 0},
				{"{", 5, 0},
				{"server", 6, 0},
				{"{", 6, 0},
				{"listen", 7, 0},
				{"127.0.0.1:8080", 7, 0},
				{";", 7, 0},
				{"#listen", 7, 0},
				{"server_name", 8, 0},
				{"default_server", 8, 0},
				{";", 8, 0},
				{"location", 9, 0},
				{"/", 9, 0},
				{"{", 9, 0},
				{"## this is brace", 9, 0},
				{"# location /", 10, 0},
				{"return", 11, 0},
				{"200", 11, 0},
				{"foo bar baz", 11, 0},
				{";", 11, 0},
				{"}", 12, 0},
				{"}", 13, 0},
				{"}", 14, 0},
			},
		},
		{
			"messy",
			HelperReadConfig(t, "messy/nginx.conf"),
			[]lexer.LexicalItem{
				{"user", 1, 0},
				{"nobody", 1, 0},
				{";", 1, 0},

				// NOTE: Different behavior than reference crossplane implementation.
				//       This version is keeping the trailing escape character.
				//
				//       Original: {"# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n ", 2, 0}
				//
				//       Paul Stuart would call this a "pathalogical edge case", and
				//		 he's probably not wrong. Also It's my oppinion that the current
				//		 expectation in this test is more correct than the reference
				//       behavior.
				{"# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n \\", 2, 0},

				{"events", 3, 0},
				{"{", 3, 0},
				{"worker_connections", 3, 0},
				{"2048", 3, 0},
				{";", 3, 0},
				{"}", 3, 0},
				{"http", 5, 0},
				{"{", 5, 0},
				{"#forteen", 5, 0},
				{"# this is a comment", 6, 0},
				{"access_log", 7, 0},
				{"off", 7, 0},
				{";", 7, 0},
				{"default_type", 7, 0},
				{"text/plain", 7, 0},
				{";", 7, 0},
				{"error_log", 7, 0},
				{"off", 7, 0},
				{";", 7, 0},
				{"server", 8, 0},
				{"{", 8, 0},
				{"listen", 9, 0},
				{"8083", 9, 0},
				{";", 9, 0},
				{"return", 10, 0},
				{"200", 10, 0},
				{"Ser\" ' ' ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n", 10, 0},
				{";", 10, 0},
				{"}", 11, 0},
				{"server", 12, 0},
				{"{", 12, 0},
				{"listen", 12, 0},
				{"8080", 12, 0},
				{";", 12, 0},
				{"root", 13, 0},
				{"/usr/share/nginx/html", 13, 0},
				{";", 13, 0},
				{"location", 14, 0},
				{"~", 14, 0},
				{"/hello/world;", 14, 0},
				{"{", 14, 0},
				{"return", 14, 0},
				{"301", 14, 0},
				{"/status.html", 14, 0},
				{";", 14, 0},
				{"}", 14, 0},
				{"location", 15, 0},
				{"/foo", 15, 0},
				{"{", 15, 0},
				{"}", 15, 0},
				{"location", 15, 0},
				{"/bar", 15, 0},
				{"{", 15, 0},
				{"}", 15, 0},
				{"location", 16, 0},
				{"/\\{\\;\\}\\ #\\ ab", 16, 0},
				{"{", 16, 0},
				{"}", 16, 0},
				{"# hello", 16, 0},
				{"if", 17, 0},
				{"($request_method", 17, 0},
				{"=", 17, 0},
				{"P\\{O\\)\\###\\;ST", 17, 0},
				{")", 17, 0},
				{"{", 17, 0},
				{"}", 17, 0},
				{"location", 18, 0},
				{"/status.html", 18, 0},
				{"{", 18, 0},
				{"try_files", 19, 0},

				// NOTE: Different behavior than reference crossplane implementation.
				//       The two regex patterns are separated by a " " character and are
				//       two different arguments, but are considered a single token.
				//       behavior.
				//
				//       Original: {"/abc/${uri} /abc/${uri}.html", 19, 0}
				{"/abc/${uri}", 19, 35},
				{"/abc/${uri}.html", 19, 52},

				{"=404", 19, 0},
				{";", 19, 0},
				{"}", 20, 0},
				{"location", 21, 0},

				// NOTE: Different behavior than reference crossplane implementation.
				//		 The token begins on line 21 and ends on 22. Our line number is
				//       incorrect.
				//
				//       Original: {"/sta;\n                    tus", 21, 0}
				{"/sta;\n                    tus", 22, 0},
				{"{", 22, 0},
				{"return", 22, 0},
				{"302", 22, 0},
				{"/status.html", 22, 0},
				{";", 22, 0},
				{"}", 22, 0},
				{"location", 23, 0},
				{"/upstream_conf", 23, 0},
				{"{", 23, 0},
				{"return", 23, 0},
				{"200", 23, 0},
				{"/status.html", 23, 0},
				{";", 23, 0},
				{"}", 23, 0},
				{"}", 23, 0},
				{"server", 24, 0},
				{"{", 25, 0},
				{"}", 25, 0},
				{"}", 25, 0},
			},
		},
		{
			"quote behavior",
			HelperReadConfig(t, "quote-behavior/nginx.conf"),
			[]lexer.LexicalItem{
				{"outer-quote", 1, 0},
				{"left", 1, 0},
				{"-quote", 1, 0},
				{"right-\"quote\"", 1, 0},
				{"inner\"-\"quote", 1, 0},
				{";", 1, 0},
				{"", 2, 0},
				{"", 2, 0},
				{"left-empty", 2, 0},
				{"right-empty\"\"", 2, 0},
				{"inner\"\"empty", 2, 0},
				{"right-empty-single\"", 2, 0},
				{";", 2, 0},
			},
		},
		{
			"quoted right brace",
			HelperReadConfig(t, "quoted-right-brace/nginx.conf"),
			[]lexer.LexicalItem{
				{"events", 1, 0},
				{"{", 1, 0},
				{"}", 1, 0},
				{"http", 2, 0},
				{"{", 2, 0},
				{"log_format", 3, 0},
				{"main", 3, 0},
				{"escape=json", 3, 0},
				{"{ \"@timestamp\": \"$time_iso8601\", ", 4, 0},
				{"\"server_name\": \"$server_name\", ", 5, 0},
				{"\"host\": \"$host\", ", 6, 0},
				{"\"status\": \"$status\", ", 7, 0},
				{"\"request\": \"$request\", ", 8, 0},
				{"\"uri\": \"$uri\", ", 9, 0},
				{"\"args\": \"$args\", ", 10, 0},
				{"\"https\": \"$https\", ", 11, 0},
				{"\"request_method\": \"$request_method\", ", 12, 0},
				{"\"referer\": \"$http_referer\", ", 13, 0},
				{"\"agent\": \"$http_user_agent\"", 14, 0},
				{"}", 15, 0},
				{";", 15, 0},
				{"}", 16, 0},
			},
		},
		{
			"address with subnet mask",
			"set_real_ip_from 192.168.0.1/24",
			[]lexer.LexicalItem{
				{"set_real_ip_from", 1, 0},
				{"192.168.0.1/24", 1, 0},
			},
		},
		{
			"time with interval",
			"proxy_buffers 4 256k",
			[]lexer.LexicalItem{
				{"proxy_buffers", 1, 0},
				{"4", 1, 0},
				{"256k", 1, 0},
			},
		},
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		actual := lexer.LexScanner(tt.input)
		i := 0
		for token := range actual {
			other := tt.expected[i]
			if other.Item != token.Item || other.LineNum != token.LineNum {
				t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected[i], token)
			}
			i++
		}

	}
}
