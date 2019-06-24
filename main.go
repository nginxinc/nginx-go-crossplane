package main

import (
	"os"

	"github.com/nginxinc/crossplane-go/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
