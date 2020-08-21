package parser

import (
	"fmt"
	"io/ioutil"
	"log"
)

// readTestData loads a testdata file by name. If the file cannot be loaded or read then
// the calling test will fail
func readTestData(filename string) []byte {
	b, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		log.Fatalf("Error loading from testdata, %s", err)
	}
	return b
}
