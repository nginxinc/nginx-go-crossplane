package lexer

import (
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
				{"{", 1, 0}, {"}", 1, 0},
			},
			"",
		},
		{
			"balanced: simple long",
			[]LexicalItem{
				{"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0},
				{"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0}, {"}", 1, 0},
			},
			"",
		},
		{
			"balanced: with directives",
			[]LexicalItem{
				{"http", 1, 0}, {"{", 1, 0}, {"user", 1, 0}, {"nginx", 1, 0}, {";", 1, 0}, {"}", 1, 0},
			},
			"",
		},
		{
			"unbalanced: needle in a haystack",
			[]LexicalItem{
				{"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0}, {"{", 1, 0},
				{"}", 2, 0}, {"}", 2, 0}, {"}", 2, 0}, {"}", 2, 0}, {"{", 2, 0}, {"}", 2, 0}, {"}", 2, 0}, {"}", 2, 0},
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
				{"http", 1, 0}, {"{", 1, 0}, {"user", 1, 0}, {"nginx", 1, 0}, {";", 1, 0}, {"}", 1, 0}, {"# a comment", 1, 0},
			},
		},
		{
			"Messy: multiline file ",
			`# hello\n\\n\\\n worlddd  \#\\#\\\# dfsf\n \\n \\\n \ 
			http {#forteen
    			access_log off;default_type text/plain; error_log off;
				"return" 200 "Ser\" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n";
    		}`,
			[]LexicalItem{
				{"# hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n \\ ", 1, 0},
				{"http", 2, 0}, {"{", 2, 0}, {"#forteen", 2, 0}, {"access_log", 3, 0}, {"off", 3, 0},
				{";", 3, 0}, {"default_type", 3, 0}, {"text/plain", 3, 0}, {";", 3, 0},
				{"error_log", 3, 0}, {"off", 3, 0}, {";", 3, 0}, {"\"return\"", 4, 0}, {"200", 4, 0},
				{`"Ser" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n"`, 4, 0}, {";", 4, 0}, {"}", 5, 0},
			},
		},
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		actual := LexScanner(tt.input)
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
