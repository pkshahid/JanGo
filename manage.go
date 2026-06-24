package main

import (
	"github.com/pkshahid/JanGo/management"
	"os"
)

func main() {
	management.Execute(os.Args[1:])
}
