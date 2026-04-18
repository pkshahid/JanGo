package main

import (
	"fmt"
	"os"

	_ "github.com/pkshahid/JanGo/examples/blog/app"
	"github.com/pkshahid/JanGo/management"
)

func main() {
	// Execute management command
	if err := management.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
