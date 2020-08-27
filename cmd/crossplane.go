package cmd

import (
	"encoding/json"
	"fmt"
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

var treeCmd = &cobra.Command{
	Use:   "tree [/path/to/nginx.conf]",
	Short: "Parses an NGINX config and prints tree",
	Args:  cobra.ExactArgs(1),
}

var editCmd = &cobra.Command{
	Use:   "edit [/path/to/nginx.conf] [/path/to/edits.json]",
	Short: "Modifies an NGINX config by changesets",
	Args:  cobra.ExactArgs(3),
}

var getCmd = &cobra.Command{
	Use:   "get [/path/to/nginx.conf] [/config/path/entry]",
	Short: "Prints the value (by path) of an NGINX config",
	Args:  cobra.ExactArgs(2),
}

// Execute - cmd entrypoint
func Execute() (err error) {
	var (
		indent      uint
		outFile     string
		combine     bool
		comment     bool
		single      bool
		catchErrors bool
		debug       bool
		ignore      []string
	)
	parseCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	parseCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output to a file. If not specified, it will output to STDOUT")
	parseCmd.Flags().BoolVar(&catchErrors, "catch-errors", false, "Stop parse after first error")
	parseCmd.Flags().BoolVar(&combine, "combine", false, "Inline includes to create single config object")
	parseCmd.Flags().BoolVar(&single, "single", false, "Skip includes")
	parseCmd.Flags().BoolVar(&comment, "exclude-comments", false, "Exclude comments from json")
	parseCmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "List of ignored directives")
	parseCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]
		payload, err := parser.ParseFile(filename, ignore, catchErrors, single, comment)
		if err != nil {
			log.Fatalf("Error parsing file %s: %v", filename, err)
		}
		if combine {
			fmt.Println("unifying payload")
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
		_, err := os.Stat(filename)
		if err != nil {
			log.Fatalf("Error: cannot access file %s", filename)
		}

		f, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error: cannot read file %s: %v", filename, err)
		}
		input, err := builder.NewPayload(f)
		if err != nil {
			log.Fatalf("Error translating payload file %s: %v", filename, err)
		}
		output, err := builder.BuildFiles(input, buildPath, int(indent), tabs, false)
		if err != nil {
			log.Fatalf("Error: cannot build file %s: %v", filename, err)
		}
		os.Stdout.WriteString(output)
		os.Stdout.Write([]byte("\n")) // compatibility: always return newline at end
	}

	var (
		lineNumbers bool
	)
	lexCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	lexCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output to a file. If not specified, it will output to STDOUT")
	lexCmd.Flags().BoolVarP(&lineNumbers, "line-numbers", "n", false, "Include line numbers in output")
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
		tokenStream := lexer.LexScanner(input)
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

	treeCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	treeCmd.Flags().BoolVar(&combine, "combine", false, "Inline includes to create single config object")
	treeCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]

		// TODO: how to enable debugging from rootCmd to make it global?
		parser.Debugging = debug

		payload, err := parser.ParseFile(filename, ignore, catchErrors, single, comment)
		if err != nil {
			log.Fatalf("Error parsing file %s: %v", filename, err)
		}
		if combine {
			payload, err = payload.Unify()
			if err != nil {
				log.Fatal(err)
			}
		}
		parser.NewTree(payload).ShowTree()
	}

	editCmd.Flags().UintVarP(&indent, "indent", "i", 4, "Set spaces for indentation")
	editCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]
		editname := args[1]
		dirname := args[2]

		parser.Debugging = debug

		s, err := os.Stat(dirname)
		if err != nil {
			log.Fatalf("directory %q error: %v", dirname, err)
		}
		if !s.IsDir() {
			log.Fatalf("%q is not a directory", dirname)
		}

		changed, err := parser.ChangeMe(filename, editname)
		if err != nil {
			log.Fatalf("change failed: %v", err)
		}

		indent = 4
		tabs = false
		header := false
		_, err = builder.BuildFiles(*(changed.Payload), dirname, int(indent), tabs, header)
		if err != nil {
			log.Fatalf("oh crap: %v", err)
		}
		if debug {
			parser.NewTree(changed.Payload).ShowTree()
		}
	}

	getCmd.Flags().BoolVar(&combine, "combine", false, "Inline includes to create single config object")
	getCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]
		key := args[1]
		payload, err := parser.ParseFile(filename, ignore, catchErrors, single, comment)
		if err != nil {
			log.Fatalf("Error parsing file %s: %v", filename, err)
		}
		if combine {
			payload, err = payload.Unify()
			if err != nil {
				log.Fatal(err)
			}
		}
		tm := parser.NewTree(payload)
		fmt.Println("GET:", key)
		val, err := tm.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("GOT (%T): %+v\n", val, val)
	}

	rootCmd.AddCommand(parseCmd, buildCmd, lexCmd, treeCmd, editCmd, getCmd)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debugging output")
	return rootCmd.Execute()
}
