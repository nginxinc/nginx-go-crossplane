// replacing these tests with python "golden master configs"
// +build golden

package parser_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

func init() {
	// evaluate configs from project root
	os.Chdir("../..")
}

// HelperFilepath returns a relative filepath to a file in the testdata directory
func HelperFilepath(paths ...string) string {
	paths = append([]string{"testdata/configs"}, paths...)
	return filepath.Join(paths...)
}

var tests = []struct {
	title  string
	arg    parser.ParseArgs
	file   string
	config []*parser.Directive
}{
	{
		"basic : test Parse ",
		parser.ParseArgs{
			FileName:    "config/simple.conf",
			CatchErrors: true,
		},
		"config/simple.conf",
		[]*parser.Directive{
			{
				Directive: "events",
				Line:      1,
				Block: []*parser.Directive{
					{
						Directive: "worker_connections",
						Line:      2,
						Args:      []string{"1024"},
					},
				},
			},
			{
				Directive: "http",
				Line:      5,
				Block: []*parser.Directive{
					{
						Directive: "server",
						Line:      6,
						Block: []*parser.Directive{
							{
								Directive: "listen",
								Args:      []string{"127.0.0.1:8080"},
								Line:      7,
							},
							{
								Directive: "server_name",
								Args:      []string{"default_server"},
								Line:      8,
							},
							{
								Directive: "location",
								Args:      []string{"/"},
								Line:      9,
								Block: []*parser.Directive{
									{
										Directive: "return",
										Args:      []string{"200", "foo bar baz"},
										Line:      10,
									},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		"Test : with Comments",
		parser.ParseArgs{
			FileName:    "config/withComments.conf",
			CatchErrors: true,
			Comments:    true,
		},
		"config/withComments.conf",
		[]*parser.Directive{
			{
				Directive: "http",
				Line:      1,
				Block: []*parser.Directive{
					{
						Directive: "server",
						Line:      2,
						Block: []*parser.Directive{
							{
								Directive: "listen",
								Args:      []string{"127.0.0.1:8080"},
								Line:      3,
							},
							{
								Directive: "#",
								Line:      3,
								Comment:   "listening",
							},
						},
					},
				},
			},
		},
	},
	{
		"basic : messy test",
		parser.ParseArgs{
			FileName:    "config/messy.conf",
			CatchErrors: true,
			Comments:    true,
		},
		"config/messy.conf",
		[]*parser.Directive{
			{
				Directive: "user",
				Args:      []string{"nobody"},
				Line:      1,
			},
			{
				Directive: "#",
				Line:      2,
				Comment:   " hello\\n\\\\n\\\\\\n worlddd  \\#\\\\#\\\\\\# dfsf\\n \\\\n \\\\\\n ", // removed last 2 "\\"
			},
			{
				Directive: "events",
				Line:      3,
				Block: []*parser.Directive{
					{
						Directive: "worker_connections",
						Args:      []string{"2048"},
						Line:      3,
					},
				},
			},
			{
				Directive: "http",
				Line:      5,
				Block: []*parser.Directive{
					{
						Directive: "#",
						Line:      5,
						Comment:   "forteen",
					},
					{
						Directive: "#",
						Line:      6,
						Comment:   " this is a comment",
					},
					{
						Directive: "access_log",
						Args:      []string{"off"},
						Line:      7,
					},
					{
						Directive: "default_type",
						Args:      []string{"text/plain"},
						Line:      7,
					},
					{
						Directive: "error_log",
						Args:      []string{"off"},
						Line:      7,
					},
					{
						Directive: "server",
						Line:      8,
						Block: []*parser.Directive{
							{
								Directive: "listen",
								Args:      []string{"8083"},
								Line:      9,
							},
							{
								Directive: "return",
								Args:      []string{"200", `Ser" ' ' ver\\ \ $server_addr:\$server_port\n\nTime: $time_local\n\n`},
								Line:      10,
							},
						},
					},
					{
						Directive: "server",
						Line:      12,
						Block: []*parser.Directive{
							{
								Directive: "listen",
								Args:      []string{"8080"},
								Line:      12,
							},
							{
								Directive: "root",
								Args:      []string{"/usr/share/nginx/html"},
								Line:      13,
							},
							{
								Directive: "location",
								Args:      []string{"~", "/hello/world;"},
								Line:      14,
								Block: []*parser.Directive{
									{
										Directive: "return",
										Args:      []string{"301", "/status.html"},
										Line:      14,
									},
								},
							},
							{
								Directive: "location",
								Args:      []string{"/foo"},
								Line:      15,
							},
							{
								Directive: "location",
								Args:      []string{"/bar"},
								Line:      15,
							},
							{
								Directive: "location",
								Args:      []string{"/\\{\\;\\}\\ #\\ ab"},
								Line:      16,
							},
							{
								Directive: "#",
								Line:      16,
								Comment:   " hello",
							},
							{
								Directive: "if",
								Args:      []string{"$request_method", "=", "P\\{O\\)\\###\\;ST"},
								Line:      17,
							},
							{
								Directive: "location",
								Args:      []string{"/status.html"},
								Line:      18,
								Block: []*parser.Directive{
									{
										Directive: "try_files",
										Args:      []string{"/abc/${uri}", "/abc/${uri}.html", "=404"},
										Line:      19,
									},
								},
							},

							{
								Directive: "location",
								Args:      []string{"/sta;\n                    tus"},
								Line:      21,
								Block: []*parser.Directive{
									{
										Directive: "return",
										Args:      []string{"302", "/status.html"},
										Line:      22,
									},
								},
							},

							{
								Directive: "location",
								Args:      []string{"/upstream_conf"},
								Line:      23,
								Block: []*parser.Directive{
									{
										Directive: "return",
										Args:      []string{"200", "/status.html"},
										Line:      23,
									},
								},
							},
						},
					},
					{
						Directive: "server",
						Line:      24,
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {
	//t.Skip("switching to GM")
	parser.Debugging = testing.Verbose()
	tests := map[string]struct {
		args parser.ParseArgs
		want *parser.Payload
	}{
		"includes regular": {
			parser.ParseArgs{
				FileName:    HelperFilepath("includes-regular/nginx.conf"),
				CatchErrors: true,
				Comments:    false,
			},
			&parser.Payload{
				Status: "failed",
				Errors: []parser.ParseError{
					{
						File:   "testdata/configs/includes-regular/conf.d/server.conf",
						Line:   5,
						Column: 27,
						Fail:   `[Errno 2] No such file or directory: 'testdata/configs/includes-regular/bar.conf'`,
					},
				},

				Config: []*parser.Config{
					{
						Status: "ok",
						File:   "testdata/configs/includes-regular/nginx.conf",
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
							},
							{
								Directive: "http",
								Line:      5,
								Block: []*parser.Directive{
									{
										Directive: "include",
										Args:      []string{"conf.d/server.conf"},
										Line:      5,
									},
								},
							},
						},
					},
					{
						File:   "testdata/configs/includes-regular/foo.conf",
						Status: "ok",
						Parsed: []*parser.Directive{
							{
								Directive: "location",
								Args:      []string{"/foo"},
								Line:      4,
								Block: []*parser.Directive{
									{
										Directive: "return",
										Args:      []string{"200", "foo"},
										Line:      5,
									},
								},
							},
						},
					},
					{
						File:   "testdata/includes-regular/conf.d/server.conf",
						Status: "ok",
						Errors: []parser.ConfigError{
							{
								Line: 5,
							},
						},

						Parsed: []*parser.Directive{
							{
								Directive: "server",
								Line:      1,
								Block: []*parser.Directive{
									{
										Directive: "listen",
										Args:      []string{"127.0.0.1:8080"},
										Line:      2,
									},
									{
										Directive: "server_name",
										Args:      []string{"default_server"},
										Line:      3,
									},
									{
										Directive: "include",
										Args:      []string{"foo.conf"},
										Line:      4,
										Includes:  []int{1},
									},
									{
										Directive: "include",
										Args:      []string{"bar.conf"},
										Line:      5,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parser.Parse(tt.args)

			if err != nil {
				t.Fatalf("Failed to parse file, %s", err)
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("\nTest assertion failed (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	t.Skip("redo this")
	for _, tes := range tests {
		file, err := ioutil.ReadFile(tes.arg.FileName)
		if err != nil {
			t.Error(err)
			continue
		}
		t.Logf("parsing: %s\n", tes.arg.FileName)
		parsed, _ := parser.ParseString(string(file), tes.arg)
		if parsed.Errors != nil {
			t.Errorf("Errors encountered: %v", parsed.Errors)
		}
		if len(parsed.Config) < 1 {
			t.Errorf("No configurations parsed for %s", tes.arg.FileName)
		}
		par := parsed.Config[0].Parsed
		for p := 0; p < len(par); p++ {
			compareDirectives(t, par[p], tes.config[p])
		}
	}
}

func absConf(t *testing.T) error {
	t.Helper()
	args := parser.ParseArgs{FileName: "/etc/nginx/nginx.conf", ConfigDir: "/etc/nginx"}
	_, err := parser.Parse(args)
	return err
}

func TestChroot(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("only runs as root")
	}
	fn, err := Chroot("config/absolute")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fn(); err != nil {
			t.Error(err)
		}
	}()
	if err := absConf(t); err != nil {
		t.Fatal(err)
	}
}

func Chroot(path string) (func() error, error) {
	root, err := os.Open("/")
	if err != nil {
		return nil, err
	}

	if err := syscall.Chroot(path); err != nil {
		root.Close()
		return nil, err
	}

	return func() error {
		defer root.Close()
		if err := root.Chdir(); err != nil {
			return err
		}
		return syscall.Chroot(".")
	}, nil
}

func compareDirectives(t *testing.T, gen, config *parser.Directive) {
	t.Helper()
	if gen.Directive != config.Directive {
		t.Errorf("directives want:%q got:%q\n", config.Directive, gen.Directive)
	}
	// loop over and compare
	if len(gen.Args) == len(config.Args) {
		for i, arg := range config.Args {
			if gen.Args[i] != arg {
				t.Errorf("(directive: %s) arg #%d want:%q got:%q\n", gen.Directive, i+1, arg, gen.Args[i])
			}
		}
	} else {
		t.Errorf("mismatched arg count - want(%d):%v got(%d):%v\n",
			len(config.Args), config.Args, len(gen.Args), gen.Args)
	}
	if gen.Line != config.Line {
		t.Errorf("(%s) line want:%d got:%d\n", config.Directive, config.Line, gen.Line)
	}
	/*
		if gen.File != config.File {
			t.Errorf("File want:%q got:%q\n", config.File, gen.File)
		}
	*/
	if gen.Comment != config.Comment {
		t.Errorf("Comment \nwant:%q\n got:%q\n", config.Comment, gen.Comment)
	}

	if len(gen.Block) == len(config.Block) {
		for i := 0; i < len(gen.Block); i++ {
			compareDirectives(t, gen.Block[i], config.Block[i])
		}
	} else {
		t.Errorf("(%s/%d) Mismatched blocks want:%d got:%d\nwant: %v\n got: %v\n",
			config.Directive, config.Line, len(config.Directive), len(gen.Directive), config.Directive, gen.Directive)
	}
}

func TestParseAbs(t *testing.T) {
	t.Skip("redo this too")
	if testing.Verbose() {
		parser.Debugging = true
	}

	args := parser.ParseArgs{FileName: "config/absolute.conf", PrefixPath: "/etc/nginx/"}
	_, err := parser.Parse(args)

	if err != nil {
		t.Fatal(err)
	}

	args.FileName, err = filepath.Abs(args.FileName)
	if err != nil {
		t.Fatal(err)
	}
	_, err = parser.Parse(args)

	if err != nil {
		t.Fatal(err)
	}
}
