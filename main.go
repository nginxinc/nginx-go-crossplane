package main

import (
	"os"

	"gitlab.com/f5/nginx/crossplane-go/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
