// Command git-sweep provides a cross-platform CLI that removes local branches
// whose upstream has been removed. It is safe by default and supports dry-run
// and non-interactive operation.
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
