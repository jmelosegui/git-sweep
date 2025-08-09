// Command git-sweep provides a cross-platform CLI that removes local branches
// whose upstream has been removed. It is safe by default and supports dry-run
// and non-interactive operation.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	gitpkg "github.com/jmelosegui/git-sweep/internal/git"
	sweeppkg "github.com/jmelosegui/git-sweep/internal/sweep"
	uipkg "github.com/jmelosegui/git-sweep/internal/ui"
)

var (
	version = "v0.0.0-dev"
)

func main() {
	// Flags
	showVersion := flag.Bool("version", false, "print version and exit")
	remote := flag.String("remote", "origin", "git remote to use for fetch --prune")
	include := flag.String("include", "", "regex to include branch names")
	exclude := flag.String("exclude", "", "regex to exclude branch names")
	jsonOut := flag.Bool("json", false, "print plan as JSON")
	noColor := flag.Bool("no-color", true, "no-op placeholder: color not yet implemented")
	yes := flag.Bool("yes", false, "execute deletions (otherwise dry-run)")
	debug := flag.Bool("debug", false, "enable debug output (not used yet)")
	flag.Parse()

	_ = noColor
	_ = debug

	if *showVersion {
		fmt.Println(version)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	r := gitpkg.ExecRunner{}
	plan, err := sweeppkg.BuildPlan(ctx, r, sweeppkg.Options{
		Remote:          *remote,
		IncludePattern:  *include,
		ExcludePattern:  *exclude,
		ExtraProtected:  nil,
		ProtectCurrent:  true,
		ProtectUpstream: true,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if err := uipkg.PrintPlan(plan, uipkg.Options{JSON: *jsonOut}); err != nil {
		fmt.Println("error:", err)
		return
	}

	if !*yes {
		return
	}

	// Execute deletions when --yes is provided
	res, err := sweeppkg.ExecuteDeletions(ctx, r, plan, sweeppkg.ExecuteOptions{MaxParallel: 0})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := uipkg.PrintDeletionResult(res); err != nil {
		fmt.Println("error:", err)
	}
}
