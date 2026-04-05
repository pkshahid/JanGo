package main

import (
	"os"
	"github.com/godjango/godjango/management"
)

func main() {
	management.Execute(os.Args[1:])
}
