package main_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/nginxinc/crossplane-go/pkg/parser"
)

var TestIncludePerf = false

func pythonCrossplane(command string, file string) (res []byte, err error) {
	cmd := []string{command, file}
	c := exec.Command("crossplane", cmd...)
	res, err = c.Output()
	if err != nil {
		e := fmt.Sprintf("Cannot execute python crossplane: %v", err)
		return res, errors.New(e)
	}
	return res, err
}

func goCrossplane(command string, file string) (res []byte, err error) {
	cmd := []string{command, "-n", file}
	c := exec.Command("./crossplane-go", cmd...)
	res, err = c.Output()
	if err != nil {
		e := fmt.Sprintf("Cannot execute command: crossplane-go %v", cmd)
		return res, errors.New(e)
	}
	return res, err
}

func byteToPayload(b []byte) (p parser.Payload, err error) {
	p = parser.Payload{}
	err = json.Unmarshal(b, &p)
	return p, err
}

func setupTest() (err error) {
	// check for python crossplane first.. this is expected to be in /bin now
	cmd := exec.Command("crossplane", "--help")
	err = cmd.Run()
	if err != nil {
		e := fmt.Sprintf("Cannot execute this test: need python crossplane installed. Err: %v", err)
		return errors.New(e)
	}

	// try to build binary so we can test off that
	err = exec.Command("go", "build").Start()
	if err != nil {
		e := fmt.Sprintf("Cannot build go binary: %v", err)
		return errors.New(e)
	}
	return err
}

func TestCrossplanesParse(t *testing.T) {

	err := setupTest()
	if err != nil {
		t.Fatal(err)
	}
	err = filepath.Walk("./cmd/configs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Errorf("Cannot access path %s: %v\n", path, err)
			return err
		}

		if strings.Contains(info.Name(), "nginx.conf") {
			p, e := filepath.Abs(path)
			if e != nil {
				t.Errorf("Cannot convert to absolute path: %s", path)
			}

			p1, e := pythonCrossplane("parse", p)
			if e != nil {
				t.Errorf("Cannot parse using crossplane-python: %v", e)
			}
			o1, e := byteToPayload(p1)
			if e != nil {
				t.Errorf("Cannot convert []byte to payload: %v", e)
			}

			p2, e := goCrossplane("parse", p)
			if e != nil {
				t.Errorf("Cannot parse using crossplane-go: %v", e)
			}
			o2, e := byteToPayload(p2)
			if e != nil {
				t.Errorf("Cannot convert []byte to payload: %v", e)
			}

			if diff := deep.Equal(o1, o2); diff != nil {
				t.Fatal(diff)
			}

		}
		return nil
	})
	if err != nil {
		t.Fatalf("Cannot walk dir ./cmd/configs: %v", err)
	}
}

func TestCrossplanesLex(t *testing.T) {

	err := setupTest()
	if err != nil {
		t.Fatal(err)
	}
	casesFailed := []string{}
	err = filepath.Walk("./cmd/configs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Errorf("Cannot access path %s: %v\n", path, err)
			return err
		}

		skipCases := []string{"lua"}

		for _, skip := range skipCases {
			if strings.Contains(info.Name(), skip) {
				return filepath.SkipDir
			}
		}

		if strings.Contains(info.Name(), "nginx.conf") {
			p, e := filepath.Abs(path)
			if e != nil {
				t.Errorf("Cannot convert to absolute path: %s", path)
			}

			pstart := time.Now()
			o1, e := pythonCrossplane("lex", p)
			pelapsed := time.Since(pstart)
			if e != nil {
				t.Errorf("Cannot lex using crossplane-python: %v", e)
			}

			gstart := time.Now()
			o2, e := goCrossplane("lex", p)
			gelapsed := time.Since(gstart)
			if e != nil {
				t.Errorf("Cannot lex using crossplane-go: %v", e)
			}

			if string(o1) != string(o2) {
				var jo1, jo2 = []interface{}{}, []interface{}{}

				_ = json.Unmarshal(o1, &jo1)
				_ = json.Unmarshal(o2, &jo2)

				if diff := deep.Equal(jo1, jo2); diff != nil {
					casesFailed = append(casesFailed, p)
					e := fmt.Sprintf("Found different strings for case %s: (python left, golang right)", p)
					t.Error(e)
					m := len(diff)
					for i, d := range diff {
						t.Errorf("\t difference (%d of %d) found: %v", i+1, m, d)
					}
				}
				t.Log("=============================\n")
			}
			if TestIncludePerf && (gelapsed > pelapsed) {
				t.Errorf("Lex took %s slower", gelapsed-pelapsed)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Cannot walk dir ./cmd/configs: %v", err)
	}

	if len(casesFailed) > 0 {
		t.Error("Cases failed:")
		for _, cf := range casesFailed {
			t.Logf("%s\n", cf)
		}
	}
}
