package lexer_test

import (
	"io/ioutil"
	"path"
	"testing"
)

func HelperReadConfig(t testing.TB, filepath string) string {
	t.Helper()
	f, err := ioutil.ReadFile(path.Join("testdata", filepath))
	if err != nil {
		t.Fatalf("Unable to read config file, %s", err)
	}
	return string(f)
}
