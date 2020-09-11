package parser

import (
	"os"
	"regexp"
	"testing"
)

func TestApply(t *testing.T) {
	filename := "config/moar.conf"
	p, err := ParseFile(filename, nil, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	tm := TreeMap{Payload: p}
	tm.buildTree()
	re := regexp.MustCompile(".*baby.*")
	matcher := func(b *Directive) bool {
		t.Log("CHECKING:", b.Directive)
		if b.Directive == "location" && len(b.Args) > 0 {
			arg := b.Args[0]
			t.Logf("possibly (%t): %s\n", re.MatchString(arg), arg)
			return re.MatchString(b.Args[0])
		}
		return false
	}
	adding := []*Directive{
		{Directive: "biteme"},
	}
	if err := tm.Apply("/http", matcher, adding...); err != nil {
		t.Fatal(err)
	}
	if testing.Verbose() {
		tm.ShowTree(os.Stdout)
	}
}
