package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/builder"
	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/lexer"
	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

var rootCmd = &cobra.Command{
	Use:   "crossplane",
	Short: "Crossplane is a quick and reliable way to convert NGINX configurations into JSON and back.",
	Long: `A quick and reliable way to convert NGINX configurations into JSON and back.

built with ❤ by nginxinc and gophers who live in Cork and are from Cork
Complete documentation is available at: https://gitswarm.f5net.com/indigo/poc/crossplane-go
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
func Execute() (err error) {
	var (
		indent  uint
		outFile string
		combine bool
		comment bool
		single  bool
		noCatch bool
		quotes  bool
		debug   bool
		ignore  []string
		prefix  string
	)
	parseCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "path prefix of the config file on the host")
	parseCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	parseCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output to a file. If not specified, it will output to STDOUT")
	parseCmd.Flags().BoolVar(&noCatch, "no-catch", false, "Parse entire config and return all errors")
	parseCmd.Flags().BoolVar(&combine, "combine", false, "Inline includes to create single config object")
	parseCmd.Flags().BoolVar(&single, "single", false, "Skip includes")
	parseCmd.Flags().BoolVar(&quotes, "quotes", false, "Strip quotes in config")
	parseCmd.Flags().BoolVar(&comment, "include-comments", false, "Include comments in json")
	parseCmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "List of ignored directives")
	parseCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]
		arg := parser.ParseArgs{FileName: filename, Ignore: ignore, CatchErrors: !noCatch, Comments: comment, PrefixPath: prefix, StripQuotes: quotes}
		payload, err := parser.Parse(arg)
		if err != nil {
			log.Fatalf("Error parsing file %s: %v", filename, err)
		}
		if combine {
			payload, err = payload.Unify()
			if err != nil {
				log.Fatal(err)
			}
		}
		s := make([]string, indent)
		b, err := json.MarshalIndent(payload, "", strings.Join(s, " "))
		if err != nil {
			log.Fatalf("Error marshalling data: %v", err)
		}
		if outFile != "" {
			if err = ioutil.WriteFile(outFile, b, 0644); err != nil {
				log.Fatalf("Error writing data file %s: %v", outFile, err)
			}
		} else {
			os.Stdout.Write(b)
			os.Stdout.Write([]byte("\n")) // compatibility: always return newline at end
		}
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
		filename := args[0]
		input, err := parser.LoadPayload(filename)
		if err != nil {
			log.Fatalf("Error translating payload file %s: %v", filename, err)
		}
		opts := &builder.Options{
			Dirname: buildPath,
			Indent:  int(indent),
			Tabs:    tabs,
		}
		_, err = builder.BuildFiles(input, opts)
		if err != nil {
			log.Fatalf("Error: cannot build file %s: %v", filename, err)
		}
		// it used to print out the results of building,
		// but the goods are in the files, so unnecessary
	}

	var (
		lineNumbers bool
	)
	lexCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	lexCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output to a file. If not specified, it will output to STDOUT")
	lexCmd.Flags().BoolVarP(&lineNumbers, "line-numbers", "n", false, "Include line numbers in output")
	lexCmd.Flags().BoolVar(&quotes, "quotes", false, "Strip quotes in config")
	lexCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]
		_, err := os.Stat(filename)
		if err != nil {
			log.Fatalf("Error: cannot access file %s", filename)
		}
		f, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error: cannot read file %s: %v", filename, err)
		}
		input := string(f)
		tokenStream := lexer.LexScanner(input, quotes)
		li := []interface{}{}
		for token := range tokenStream {
			li = append(li, token.Repr(lineNumbers))
		}
		b, err := json.Marshal(li)
		if err != nil {
			log.Fatalf("Error marshalling token stream data from lexer: %v", err)
		}
		os.Stdout.Write(b)
		os.Stdout.Write([]byte("\n")) // compatibility: always return newline at end
	}

	rootCmd.AddCommand(parseCmd, buildCmd, lexCmd)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debugging output")
	return rootCmd.Execute()
}
