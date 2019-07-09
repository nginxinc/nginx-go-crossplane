package lexer

import (
	"fmt"
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
			UnbalancedBracesError("UnbalancedBracesError: braces are not balanced"),
		},
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		err := BalanceBraces(tt.input)

		if err != tt.expected {
			t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected, err)
		}
	}
}

func TestLexScanner(t *testing.T) {
	var testCases = []struct {
		title    string
		input    string
		expected []LexicalItem
	}{
		{
			"basic: one-line lexical analysis",
			"http { user nginx; } # a comment",
			[]LexicalItem{
				{"http", 1}, {"{", 1}, {"user", 1}, {"nginx", 1}, {";", 1}, {"}", 1}, {"# a comment", 1},
			},
		},
		{
			"Messy: multiline file ",
			`# hello\n\\n\\\n worlddd  \#\\#\\\# dfsf\n \\n \\\n \ 
			http {#forteen
    			access_log off;default_type text/plain; error_log off;
				"return" 200 "Ser\\" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n";
    		}`,
			[]LexicalItem{
				{"# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n \\ ", 1}, {"http", 2}, {"{", 2}, {"#forteen", 2}, {"access_log", 3}, {"off", 3}, {";", 3}, {"default_type", 3}, {"text/plain", 3}, {";", 3}, {"error_log", 3}, {"off", 3}, {";", 3}, {"return", 4}, {"200", 4},
				{"Ser\\\" ' ' ver\\\\ \\ $server_addr:\\$server_port\\n\\nTime: $time_local\\n\\n", 4}, {";", 4}, {"}", 5},
			},
		},
		{
			"Directive-with-space: multiline file with single quote example",
			`events {
			}
			http {
				map $http_user_agent $mobile {
					default 0;
					'~Opera Mini' 1;
				}
			}`,
			[]LexicalItem{
				{"events", 1}, {"{", 1}, {"}", 2}, {"http", 3}, {"{", 3}, {"map", 4}, {"$http_user_agent", 4}, {"$mobile", 4}, {"{", 4},
				{"default", 5}, {"0", 5}, {";", 5}, {"~Opera Mini", 6}, {"1", 6}, {";", 6}, {"}", 7}, {"}", 8},
			},
		},
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		actual := LexScanner(tt.input)
		i := 0
		for token := range actual {
			result := reflect.DeepEqual(tt.expected[i], token)

			s := fmt.Sprintf("%q \n", token)
			fmt.Println(s)

			if !result {
				t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected[i], token)

			}
			i++
		}

	}
}
