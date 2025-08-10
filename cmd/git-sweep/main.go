// Command git-sweep provides a cross-platform CLI that removes local branches
// whose upstream has been removed. It is safe by default and supports dry-run
// and non-interactive operation.
package main

import (
	"context"
	"fmt"
	"time"

	gitpkg "github.com/jmelosegui/git-sweep/internal/git"
	sweeppkg "github.com/jmelosegui/git-sweep/internal/sweep"
	uipkg "github.com/jmelosegui/git-sweep/internal/ui"
	pflag "github.com/spf13/pflag"
)

var (
	version = "v0.0.0-dev"
)

func main() {
	// GNU-style flags via pflag
	var (
		showHelp    bool
		showVersion bool
		remote      string
		include     string
		exclude     string
		jsonOut     bool
		noColor     bool
		yes         bool
		debug       bool
	)

	pflag.BoolVarP(&showHelp, "help", "h", false, "show help")
	pflag.BoolVarP(&showVersion, "version", "V", false, "print version and exit")
	pflag.StringVarP(&remote, "remote", "r", "origin", "git remote to use for fetch --prune")
	pflag.StringVarP(&include, "include", "i", "", "regex to include branch names")
	pflag.StringVarP(&exclude, "exclude", "x", "", "regex to exclude branch names")
	pflag.BoolVarP(&jsonOut, "json", "j", false, "print plan as JSON")
	pflag.BoolVar(&noColor, "no-color", true, "disable color output (placeholder)")
	pflag.BoolVarP(&yes, "yes", "y", false, "execute deletions (otherwise dry-run)")
	pflag.BoolVarP(&debug, "debug", "d", false, "enable debug output (placeholder)")
	pflag.Parse()

	if showHelp {
		printUsage()
		return
	}

	_ = noColor
	_ = debug

	if showVersion {
		fmt.Println(version)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	r := gitpkg.ExecRunner{}
	plan, err := sweeppkg.BuildPlan(ctx, r, sweeppkg.Options{
		Remote:          remote,
		IncludePattern:  include,
		ExcludePattern:  exclude,
		ExtraProtected:  nil,
		ProtectCurrent:  true,
		ProtectUpstream: true,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if err := uipkg.PrintPlan(plan, uipkg.Options{JSON: jsonOut}); err != nil {
		fmt.Println("error:", err)
		return
	}

	if !yes {
		return
	}

	// Execute deletions when --yes is provided; treat as consent to force-delete (-D)
	res, err := sweeppkg.ExecuteDeletions(ctx, r, plan, sweeppkg.ExecuteOptions{MaxParallel: 0, ForceDelete: true})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := uipkg.PrintDeletionResult(res); err != nil {
		fmt.Println("error:", err)
	}
}

func printUsage() {
	fmt.Println("usage: git sweep [<options>]")
	fmt.Println()
	fmt.Println("    -V, --version           print version and exit")
	fmt.Println("    -r, --remote <name>     git remote to use for fetch --prune (default: origin)")
	fmt.Println("    -i, --include <regex>   include branches matching regex")
	fmt.Println("    -x, --exclude <regex>   exclude branches matching regex")
	fmt.Println("    -j, --json              machine-readable plan output (JSON)")
	fmt.Println("    -y, --yes               execute deletions (consent to force-delete -D)")
	fmt.Println("    -d, --debug             enable debug output (placeholder)")
	fmt.Println("        --no-color          disable color output (placeholder)")
	fmt.Println("    -h, --help              show this help")
}
