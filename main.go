package main

import (
	"os"

	"gitswarm.f5net.com/indigo/poc/crossplane-go.git/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
