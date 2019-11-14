# crossplane-go

A quick and reliable way to convert NGINX configurations into JSON and back.

built with ❤ by nginxinc and gophers who live in Cork and are from Cork

```
Usage:
  crossplane [command]

Available Commands:
  build       Build an NGINX config using a JSON format
  help        Help about any command
  lex         Lexes tokens from an NGINX config file
  parse       Parses an NGINX config for a JSON format

Flags:
  -h, --help   help for crossplane

Use "crossplane [command] --help" for more information about a command.
```

## crossplane build

```
Build an NGINX config using a JSON format

Usage:
  crossplane build [/path/to/payload.json] [flags]

Flags:
  -f, --force         Force overwrite existing files
  -h, --help          help for build
  -i, --indent uint   Set spaces for indentation (default 4)
  -d, --path string   Output to a directory. If not specified, it will output to STDOUT
  -t, --tabs          Use tabs instead of spaces on built files
```

## crossplane parse

```
Parses an NGINX config for a JSON format

Usage:
  crossplane parse [/path/to/nginx.conf] [flags]

Flags:
      --catch-errors         Stop parse after first error
      --check-args           Run arg count analysis on directives
      --check-ctx            Run context analysis on directives
      --combine              Inline includes to create single config object
  -h, --help                 help for parse
      --ignore stringArray   List of ignored directives
  -i, --indent uint          Set spaces for indentation (default 4)
  -o, --out string           Output to a file. If not specified, it will output to STDOUT
      --single               Skip includes
      --strict               Strict mode: error on unrecognized directives
```

## crossplane lex

```
Lexes tokens from an NGINX config file

Usage:
  crossplane lex [/path/to/tokens-file.txt] [flags]

Flags:
  -h, --help           help for lex
  -i, --indent uint    Set spaces for indentation (default 4)
  -n, --line-numbers   Include line numbers in output
  -o, --out string     Output to a file. If not specified, it will output to STDOUT
```
