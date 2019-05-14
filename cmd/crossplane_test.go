package cmd

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/nginxinc/crossplane-go/pkg/parser"
)

func TestParseAndBuild(t *testing.T) {
	var tests = []struct {
		name string
		args parser.ParseArgs
	}{
		{
			"bad-args",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"directive-with-space",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"empty-value-map",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"includes-globbed",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"includes-regular",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"lua-block-larger",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"lua-block-simple",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"lua-block-tricky",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"messy",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    true,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"missing-semicolon",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"quote-behavior",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"quoted-right-brace",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"russian-text",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"simple",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"spelling-mistake",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    false,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
		{
			"with-comments",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      false,
				Comments:    true,
				Strict:      false,
				Combine:     false,
				checkCtx:    true,
				checkArgs:   true,
			},
		},
	}

	for _, t := range tests {
		file, err := ioutil.ReadFile("configs/" + t.name + "/nginx.conf")
		if err != nil {
			log.Fatal(err)
		}
		t.args.FileName = file

	}

}

func TestExecute(t *testing.T) {

}
