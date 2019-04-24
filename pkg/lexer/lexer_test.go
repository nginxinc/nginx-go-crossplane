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
	}
	for _, tt := range testCases {
		t.Log(tt.title)
		actual, err := LexScanner(tt.input)
		fmt.Println(actual)
		if err != nil {
			t.Errorf("Test failed due to: %v", err)
		}
		result := reflect.DeepEqual(tt.expected, actual)
		if !result {
			t.Errorf("Test assertion failed: \t\nexpected: %v, \t\nactual: %v", tt.expected, actual)
		}
	}
}
