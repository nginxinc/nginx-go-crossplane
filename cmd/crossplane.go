package cmd

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/structs"
	"github.com/nginxinc/crossplane-go/pkg/builder"
	"github.com/nginxinc/crossplane-go/pkg/lexer"
	"github.com/nginxinc/crossplane-go/pkg/parser"
	"github.com/spf13/cobra"
)

// mock external funcs for now

// Build -
func Build(...interface{}) interface{} {
	return "build_Foo"
}

// Parse -
func Parse(...interface{}) interface{} {
	return "parse_Foo"
}

// Lex -
func Lex(...interface{}) interface{} {
	return "lex_Foo"
}

var rootCmd = &cobra.Command{
	Use:   "crossplane",
	Short: "Crossplane is a quick and reliable way to convert NGINX configurations into JSON and back.",
	Long: `A quick and reliable way to convert NGINX configurations into JSON and back.

built with ❤ by nginxinc and gophers who live in Cork and are from Cork
Complete documentation is available at: https://github.com/nginxinc/crossplane-go
	`,
}

var parseCmd = &cobra.Command{
	Use:   "parse [/path/to/nginx.conf]",
	Short: "Parses an NGINX config for a JSON format",
	Args:  cobra.ExactArgs(1),
}

var buildCmd = &cobra.Command{
	Use:   "build [/path/to/payload.json]",
	Short: "Build an NGINX config using a JSON format",
	Args:  cobra.ExactArgs(1),
}

var lexCmd = &cobra.Command{
	Use:   "lex [/path/to/tokens-file.txt]",
	Short: "Lexes tokens from an NGINX config file",
	Args:  cobra.ExactArgs(1),
}

// Execute - cmd entrypoint
func Execute() {

	// TODO: strict mode	    	(BoolVarP)
	// TODO: ignore directives  	(StringArrayVarP)

	var (
		indent                                                             uint
		outFile                                                            string
		catchErrors, combine, comment, single, strict, checkctx, checkargs bool
		ignore                                                             []string
	)
	parseCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	parseCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output to a file. If not specified, it will output to STDOUT")
	parseCmd.Flags().BoolVar(&catchErrors, "catch-errors", false, "Stop parse after first error")
	parseCmd.Flags().BoolVar(&combine, "combine", false, "Inline includes to create single config object")
	parseCmd.Flags().BoolVar(&single, "single", false, "Skip includes")
	parseCmd.Flags().BoolVar(&strict, "strict", false, "Strict mode: error on unrecognized directives")
	parseCmd.Flags().BoolVar(&checkctx, "check-ctx", false, "Run context analysis on directives")
	parseCmd.Flags().BoolVar(&checkargs, "check-args", false, "Run arg count analysis on directives")
	parseCmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "List of ignored directives")
	parseCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[1]
		_, err := os.Stat(filename)
		if err != nil {
			log.Fatalf("Error: cannot access file %s", filename)
		}
		payload, err := parser.Parse(filename, catchErrors, ignore, single, comment, strict, combine, true, checkctx, checkargs)
		if err != nil {
			log.Fatalf("Error parsing file %s: %v", filename, err)
		}
		pl := structs.Map(payload)
		log.Printf("%v", pl)
	}

	var (
		buildPath   string
		force, tabs bool
	)
	buildCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	buildCmd.Flags().StringVarP(&buildPath, "path", "d", "", "Output to a directory. If not specified, it will output to STDOUT")
	buildCmd.Flags().BoolVarP(&tabs, "tabs", "t", false, "Use tabs instead of spaces on built files")
	buildCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing files")
	buildCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[1]
		_, err := os.Stat(filename)
		if err != nil {
			log.Fatalf("Error: cannot access file %s", filename)
		}

		f, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error: cannot read file %s: %v", filename, err)
		}
		input := string(f)
		output, err := builder.Build(input, int(indent), tabs, false)
		if err != nil {
			log.Fatalf("Error: cannot build file %s: %v", filename, err)
		}
		log.Printf(output)
	}

	var (
		lineNumbers bool
	)
	lexCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	lexCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output to a file. If not specified, it will output to STDOUT")
	lexCmd.Flags().BoolVarP(&lineNumbers, "line-numbers", "n", false, "Include line numbers in output")
	lexCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[1]
		_, err := os.Stat(filename)
		if err != nil {
			log.Fatalf("Error: cannot access file %s", filename)
		}
		f, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error: cannot read file %s: %v", filename, err)
		}
		input := string(f)
		tokenStream := lexer.LexScanner(input)
		for token := range tokenStream {
			log.Print(token)
		}
	}

	rootCmd.AddCommand(parseCmd, buildCmd, lexCmd)
	rootCmd.Execute()
}
