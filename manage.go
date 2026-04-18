package main

import (
	"os"
	"github.com/pkshahid/JanGo/management"
)

func main() {
	management.Execute(os.Args[1:])
}
