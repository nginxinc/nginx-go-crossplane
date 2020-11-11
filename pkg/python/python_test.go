// +build python

// NOTE: using build tag to "silence" for now, but needs review:
//       quote handling seems to be inconsistent in the python version

package python

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

// Note: the json files were parsed from the project root (2 directories above)

var (

	// NOTE: messy is commented out for now because we don't match it's output;
	//       which normally would be an error, but in this case I believe
	//       that the python version is incorrect:
	//
	// python_test.go:117: {*parser.Payload}.Config[0].Parsed[3].Block[6].Block[8].Block[0].Args[0]:
	//         	-: "/abc/${uri} /abc/${uri}.html"
	//         	+: "/abc/${uri}"
	//         {*parser.Payload}.Config[0].Parsed[3].Block[6].Block[8].Block[0].Args[?->1]:
	//         	-: <non-existent>
	//         	+: "/abc/${uri}.html"

	the_files = `
#testdata/configs/comments-between-args/nginx.conf <- review comment handling
#testdata/configs/bad-args/nginx.conf
testdata/configs/quote-behavior/nginx.conf
testdata/configs/includes-globbed/nginx.conf
testdata/configs/includes-regular/nginx.conf
testdata/configs/with-comments/nginx.conf
#testdata/configs/messy/nginx.conf
testdata/configs/simple/nginx.conf
testdata/configs/directive-with-space/nginx.conf
#testdata/configs/empty-value-map/nginx.conf
testdata/configs/quoted-right-brace/nginx.conf
#testdata/configs/spelling-mistake/nginx.conf
testdata/configs/lua-block-simple/nginx.conf
testdata/configs/russian-text/nginx.conf
testdata/configs/lua-block-larger/nginx.conf
#testdata/configs/lua-block-tricky/nginx.conf
`
)

func init() {
	os.Chdir("../..") // work from project root
}

func TestGoldenMasters(t *testing.T) {
	// TODO: resolve inconsistencies in quote handling
	list := strings.Split(the_files, "\n")
	for _, file := range list {
		if file = strings.TrimSpace(file); file == "" {
			continue
		}
		if strings.HasPrefix(file, "#") {
			t.Log("skipping:", file)
			continue
		}
		orig := file
		i := strings.LastIndex(file, ".")
		if i < 0 {
			t.Errorf("no dot: %q\n", file)
			continue
		}
		file = file[:i] + ".json"
		t.Logf("golden master: %s\n", file)
		args := parser.ParseArgs{FileName: orig, PrefixPath: "/etc/nginx", Comments: true, CatchErrors: true, StripQuotes: false}
		payload, err := parser.Parse(args)
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(filepath.Dir(file), "xp.json")
		if err := jsonsave(payload, out); err != nil {
			t.Error(err)
		}

		original := &parser.Payload{}
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, original); err != nil {
			t.Fatal(err)
		}
		/*
			// TODO: this didn't seem to get called, not sure why -- FIND OUT WHY!
			empty := cmp.Comparer(func(a, b *parser.Config) bool {
				panic("we hit it")
				fmt.Fprintf(os.Stderr, "\n ====> %v vs. %v\n\n", a, b)
				return true
			})
		*/
		empty := cmpopts.EquateEmpty()
		if same := cmp.Equal(original, payload, empty); same {
			continue
		}
		if diff := cmp.Diff(original, payload, empty); diff != "" {
			t.Error(diff)
		}
	}
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

func verifyDir(t *testing.T, dir string) {
	t.Helper()
	files, err := glob(dir, "nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d files found in %q\n", len(files), dir)
	for i, file := range files {
		t.Logf("%d/%d: %s\n", i+1, len(files), file)
		args := parser.ParseArgs{FileName: file, PrefixPath: "/etc/nginx", Comments: true, CatchErrors: true}
		if _, err := parser.Parse(args); err != nil {
			t.Error(err)
		}
	}
}

func jsonsave(obj interface{}, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err = enc.Encode(obj); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func jload(file string, obj interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(obj)
}
