package main

import (
	"fmt"
	"os"

	"github.com/k2tzumi/openapi-mock-server/cmd"
)

func main() {
	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
