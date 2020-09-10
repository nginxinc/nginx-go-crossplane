// __uild nopes

package conf

import (
	"fmt"
	"os"
	"testing"

	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/builder"
	"gitswarm.f5net.com/indigo/poc/crossplane-go/pkg/parser"
)

func TestIncludesRegular(t *testing.T) {
	const file = "./configs/includes-regular/nginx.conf"
	catcherr := true
	single := false
	comment := true
	payload, err := parser.ParseFile(file, nil, catcherr, single, comment)
	if err != nil {
		panic(err)
	}
	fmt.Println("PAYLOAD!")
	payload.Dump(os.Stdout)
	dirname := "fake"
	os.RemoveAll(dirname)
	if err := os.MkdirAll(dirname, os.ModePerm); err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
	indent := 4
	tabs := true
	header := true
	s, err := builder.BuildFiles(*payload, dirname, indent, tabs, header)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CONFIG:")
	fmt.Println(s)
	singular, err := payload.Unify()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println()
	fmt.Println("Singular:")
	singular.Dump(os.Stdout)
	fmt.Println("RENDERED")
	if err := parser.RenderDirectives(os.Stdout, singular.Config[0].Parsed); err != nil {
		t.Fatal(err)
	}
	/*
		fmt.Println("TREE TIME")
		singular.ShowTree()
		fmt.Println("TREE TIME TOO")
		payload.ShowTree()
	*/
}
