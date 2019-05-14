package cmd

import (
	"fmt"
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
				CheckCtx:    true,
				CheckArgs:   true,
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
				CheckCtx:    true,
				CheckArgs:   true,
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
				CheckCtx:    true,
				CheckArgs:   true,
			},
		},
		{
			"includes-globbed",
			parser.ParseArgs{
				FileName:    "",
				CatchErrors: true,
				Ignore:      []string{},
				Single:      true,
				Comments:    false,
				Strict:      false,
				Combine:     true,
				CheckCtx:    true,
				CheckArgs:   true,
			},
		}, /*
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
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
					CheckCtx:    true,
					CheckArgs:   true,
				},
			},*/
	}

	for _, t := range tests {
		t.args.FileName = "configs/" + t.name + "/nginx.conf"
		parsed, err := parser.Parse(t.args)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("PARSED : ", parsed)
		fmt.Println()
		// build the file back up

	}

}

func TestExecute(t *testing.T) {

}
