package main

import (
	"flag"
	"fmt"
)

var (
	version = "v0.0.0-dev"
)

func main() {
	// Minimal flag set for now to ensure CLI wiring works
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	fmt.Println("git-sweep", version)
} 