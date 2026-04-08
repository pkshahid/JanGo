package main

import (
	"fmt"
	"os"

	_ "github.com/godjango/godjango/examples/blog/app"
	"github.com/godjango/godjango/management"
)

func main() {
	// Execute management command
	if err := management.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
