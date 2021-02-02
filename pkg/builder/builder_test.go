package builder

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

var (
	whitespace = regexp.MustCompile("[\t ]+")
	newlines   = regexp.MustCompile("\n+")
	empty      = regexp.MustCompile(" \n")
)

func walkob(dir string, do func(string) error) error {
	fn := func(path string, f os.FileInfo, err error) error {
		if f == nil || f.IsDir() {
			return nil
		}
		return do(path)
	}
	return filepath.Walk(dir, fn)
}

func glob(dir string, name string) ([]string, error) {
	var files []string
	fn := func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return nil
		}
		if f.Name() == name {
			files = append(files, path)
		}
		return nil
	}
	return files, filepath.Walk(dir, fn)
}

func tempDir() string {
	dir := "/tmp/xptests"
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}
	return dir
}

func init() {
	// work from project root
	os.Chdir("../..")
}

func TestBuilder(t *testing.T) {
	t.Skip("use GMs instead")
	var tests = []struct {
		title    string
		input    []*parser.Directive
		expected string
	}{
		{
			"basic: build with comments",
			[]*parser.Directive{
				{
					Directive: "http",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "server",
							Args:      []string{},
							Line:      2,
							Includes:  []int{},
							Comment:   "",
							Block: []*parser.Directive{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      3,
									Includes:  []int{},
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      3,
									Includes:  []int{},
									Comment:   "listen",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      4,
									Includes:  []int{},
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      5,
									Includes:  []int{},
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "#",
									Args:      []string{},
									Line:      5,
									Includes:  []int{},
									Comment:   "# this is brace",
									Block:     []*parser.Directive{},
								},
							},
						},
					},
				},
			},
			`http {
    server {
        listen 127.0.0.1:8080; #listen
        server_name default_server;
        location /; ## this is brace
    }
}`,
		},
		{
			"basic: build nested and multiple args",
			[]*parser.Directive{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "worker_connections",
							Args:      []string{"1024"},
							Line:      2,
							Includes:  []int{},
							Comment:   "",
							Block:     []*parser.Directive{},
						},
					},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      4,
					Includes:  []int{},
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "server",
							Args:      []string{},
							Line:      5,
							Includes:  []int{},
							Comment:   "",
							Block: []*parser.Directive{
								{
									Directive: "listen",
									Args:      []string{"127.0.0.1:8080"},
									Line:      6,
									Includes:  []int{},
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "server_name",
									Args:      []string{"default_server"},
									Line:      7,
									Includes:  []int{},
									Comment:   "",
									Block:     []*parser.Directive{},
								},
								{
									Directive: "location",
									Args:      []string{"/"},
									Line:      8,
									Includes:  []int{},
									Comment:   "",
									Block: []*parser.Directive{
										{
											Directive: "return",
											Args:      []string{"200", "foo bar baz"},
											Line:      9,
											Includes:  []int{},
											Comment:   "",
											Block:     []*parser.Directive{},
										},
									},
								},
							},
						},
					},
				},
			},
			`events {
	worker_connections 1024;
}
http {
	server {
		listen 127.0.0.1:8080;
		server_name default_server;
		location / {
			return 200 foo bar baz;
		}
	}
}`,
		},
		{
			"basic: build include regular",
			[]*parser.Directive{
				{
					Directive: "events",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					Comment:   "",
					Block:     []*parser.Directive{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "include",
							Args:      []string{"conf.d/server.conf"},
							Line:      3,
							Includes:  []int{1},
							Comment:   "",
							Block:     []*parser.Directive{},
						},
					},
				},
			},
			`events;
http {
	include conf.d/server.conf;
}`,
		},
		{
			"basic: start with comment",
			[]*parser.Directive{
				{
					Directive: "#",
					Args:      []string{},
					Line:      1,
					Includes:  []int{},
					Comment:   "comment",
					Block:     []*parser.Directive{},
				},
				{
					Directive: "http",
					Args:      []string{},
					Line:      2,
					Includes:  []int{},
					Comment:   "",
					Block: []*parser.Directive{
						{
							Directive: "server",
							Args:      []string{},
							Line:      3,
							Includes:  []int{},
							Comment:   "",
							Block:     []*parser.Directive{},
						},
					},
				},
			},
			// WAS: #comment TODO: fix this bug
			` #comment
http {
	server;
}`,
		},
	}

	for _, test := range tests {
		result := Build(test.input, &Options{Indent: 4, Tabs: false})
		expected := strings.TrimLeft(
			strings.ReplaceAll(test.expected, "\t", strings.Repeat(" ", 4)),
			"\n",
		)

		if expected != result {
			t.Errorf("\nexpected:\n%s\ngot:\n%s\n", test.expected, result)
		}
	}
}

func BenchmarkBuild(b *testing.B) {
	args := parser.ParseArgs{FileName: "testdata/nginx-full.conf"}
	input, _ := parser.Parse(args)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Build(input.Config[0].Parsed, &Options{Indent: 4, Tabs: false})
	}
}

func TestBuildFile(t *testing.T) {
	t.Skip("use golden masters instead")
	var tests = []struct {
		title    string
		file     string
		input    *parser.Payload
		expected string
	}{
		{
			"basic: simple build files",
			"testdata/simple.conf",
			&parser.Payload{

				Errors: []parser.ParseError{},
				Config: []*parser.Config{
					{
						File:   "testdata/simple.conf",
						Errors: []parser.ConfigError{},
						Parsed: []*parser.Directive{
							{
								Directive: "events",
								Line:      1,
								Args:      []string{},
								Includes:  []int{},
								Comment:   "",
								Block: []*parser.Directive{
									{
										Directive: "worker_connections",
										Line:      2,
										Args:      []string{"1024"},
										Includes:  []int{},
										Comment:   "",
										Block:     []*parser.Directive{},
									},
								},
							},
							{
								Directive: "http",
								Line:      5,
								Args:      []string{},
								Includes:  []int{},
								Comment:   "",
								Block: []*parser.Directive{
									{
										Directive: "server",
										Line:      6,
										Args:      []string{},
										Includes:  []int{},
										Comment:   "",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      7,
												Includes:  []int{},
												Comment:   "",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "server_name",
												Args:      []string{"default_server"},
												Line:      8,
												Includes:  []int{},
												Comment:   "",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "location",
												Args:      []string{"/"},
												Line:      9,
												Includes:  []int{},
												Comment:   "",
												Block: []*parser.Directive{
													{
														Directive: "return",
														Args:      []string{"200", "foo bar baz"},
														Line:      10,
														Includes:  []int{},
														Comment:   "",
														Block:     []*parser.Directive{},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`events {
	worker_connections 1024;
}

http {
	server {
		listen 127.0.0.1:8080;
		server_name default_server;
		location / {
			return 200 foo bar baz;
		}
	}
}`,
		},
		{
			"basic: with comments build files",
			"testdata/with-comments.conf",
			&parser.Payload{

				Errors: []parser.ParseError{},
				Config: []*parser.Config{
					{
						File:   "testdata/with-comments.conf",
						Errors: []parser.ConfigError{},
						Parsed: []*parser.Directive{
							{
								Directive: "http",
								Args:      []string{},
								Line:      1,
								Includes:  []int{},
								Comment:   "",
								Block: []*parser.Directive{
									{
										Directive: "server",
										Args:      []string{},
										Line:      2,
										Includes:  []int{},
										Comment:   "",
										Block: []*parser.Directive{
											{
												Directive: "listen",
												Args:      []string{"127.0.0.1:8080"},
												Line:      3,
												Includes:  []int{},
												Comment:   "",
												Block:     []*parser.Directive{},
											},
											{
												Directive: "#",
												Args:      []string{},
												Line:      3,
												Includes:  []int{},
												Comment:   "listen",
												Block:     []*parser.Directive{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`http {
	server {
		listen 127.0.0.1:8080; #listen
	}
}`,
		},
	}

	for _, test := range tests {
		opts := &Options{Dirname: t.TempDir(), Indent: 4}
		result, err := BuildFiles(test.input, opts)
		if err != nil {
			t.Error(test.title)
		}

		expected := strings.TrimLeft(
			strings.ReplaceAll(test.expected, "\t", strings.Repeat(" ", 4)),
			"\n",
		)

		if result != expected {
			t.Logf("WANT: %q\n", expected)
			t.Logf(" GOT: %q\n", result)
			t.Error(test.title)
		}
	}
}

func TestRebuild(t *testing.T) {
	t.Skip("still some spacing differences")
	files, err := glob("testdata/configs", "nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	dir := tempDir()
	for i, file := range files {
		t.Logf("%2d/%2d: %s\n", i+1, len(files), file)
		args := parser.ParseArgs{FileName: file, Comments: true, CatchErrors: true}
		payload, err := parser.Parse(args)
		if err != nil {
			t.Error(err)
			continue
		}
		opts := &Options{Dirname: dir}
		_, err = BuildFiles(payload, opts)
		if err != nil {
			t.Error(err)
			continue
		}
		f2 := filepath.Join(dir, file)

		orig, err := cleanFile(file)
		if err != nil {
			t.Error(err)
			continue
		}
		redo, err := cleanFile(f2)
		diff := cmp.Diff(orig, redo, cmpopts.EquateEmpty())
		if diff != "" {
			t.Errorf(diff)
		}
	}

}

type diffInfo struct {
	Want string
	Have string
	Line int
}

func compare(c1, c2 chan string) []diffInfo {
	var diffs []diffInfo
	line := 0
	for {
		line++
		s1, ok1 := <-c1
		s2, ok2 := <-c2
		if s1 != s2 {
			d := diffInfo{
				Want: s1,
				Have: s2,
				Line: line,
			}
			diffs = append(diffs, d)
		}
		if !ok1 || !ok2 {
			break
		}
	}
	return diffs
}

var despaced = regexp.MustCompile(`\s+`)
var unlined = regexp.MustCompile(`\n+`)
var unquote = regexp.MustCompile(`["']`)
var brackets = regexp.MustCompile(`{\s*}`)

func cleanFile(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	s := despaced.ReplaceAllString(string(b), " ")
	s = unlined.ReplaceAllString(s, " ")
	s = unquote.ReplaceAllString(s, "")
	s = brackets.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s, nil
}

func TestBuildStrings(t *testing.T) {
	const testConfig = "testdata/configs/includes-globbed/nginx.conf"

	args := parser.ParseArgs{FileName: testConfig, Comments: true, CatchErrors: true}
	p, err := parser.Parse(args)
	if err != nil {
		t.Fatal(err)
	}
	sc := &StringsCreator{}
	opts := &Options{Creator: sc}
	_, err = BuildFiles(p, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Files) != len(p.Config) {
		t.Errorf("config files count: want %d, got %d", len(p.Config), len(sc.Files))
	}
}

func TestEmptyBraces(t *testing.T) {
	const conf = "testdata/configs/empty-braces/nginx.conf"
	t.Log("testing config:", conf)
	args := parser.ParseArgs{
		FileName: conf,
		Comments: true,
	}
	p, err := parser.Parse(args)
	if err != nil {
		t.Fatal(err)
	}
	var sc StringsCreator
	opts := &Options{Creator: &sc}
	t.Log("building using line numbers")
	_, err = BuildFiles(p, opts)
	if err != nil {
		t.Fatal(err)
	}

	// confirm the build file is effectively the same
	orig, _ := ioutil.ReadFile(conf)
	err = compareBuilds(t, string(orig), sc.Files[0].Contents)
	if err != nil {
		t.Fatal(err)
	}

	// build files has 2 internal versions,
	// one that respects line numbers, and one that doesn't
	// it uses a simple check to see if the first directive
	// has a line number, and if not uses the latter
	t.Log("building without line numbers")
	p.Config[0].Parsed[1].Line = 0
	var sc2 StringsCreator
	opts = &Options{Creator: &sc2}
	_, err = BuildFiles(p, opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range sc.Files {
		t.Log(f.Contents)
	}
}

func prep(s string) []string {
	s = whitespace.ReplaceAllString(s, " ")
	s = newlines.ReplaceAllString(s, "\n")
	s = empty.ReplaceAllString(s, "\n")
	s = newlines.ReplaceAllString(s, "\n")
	scanner := bufio.NewScanner(strings.NewReader(s))
	var list []string
	for scanner.Scan() {
		if text := scanner.Text(); text != "" {
			list = append(list, text)
		}
	}
	return list
}

func dump(t *testing.T, list []string) {
	for i, s := range list {
		t.Logf("%3d %s\n", i, s)
	}
}

func compareBuilds(t *testing.T, this, that string) error {
	c1 := prep(this)
	c2 := prep(that)
	if len(c1) != len(c2) {
		dump(t, c1)
		dump(t, c2)
		return fmt.Errorf("mismatch linecount, %d vs %d", len(c1), len(c2))
	}
	return nil
}
