// Command git-sweep provides a cross-platform CLI that removes local branches
// whose upstream has been removed. It is safe by default and supports dry-run
// and non-interactive operation.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	gitpkg "github.com/jmelosegui/git-sweep/internal/git"
	sweeppkg "github.com/jmelosegui/git-sweep/internal/sweep"
	uipkg "github.com/jmelosegui/git-sweep/internal/ui"
	updatepkg "github.com/jmelosegui/git-sweep/internal/update"
	pflag "github.com/spf13/pflag"
)

const updateCheckRepo = "jmelosegui/git-sweep"

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
		yes         bool
	)

	pflag.BoolVarP(&showHelp, "help", "h", false, "show help")
	pflag.BoolVarP(&showVersion, "version", "V", false, "print version and exit")
	pflag.StringVarP(&remote, "remote", "r", "origin", "git remote to use for fetch --prune")
	pflag.StringVarP(&include, "include", "i", "", "regex to include branch names")
	pflag.StringVarP(&exclude, "exclude", "x", "", "regex to exclude branch names")
	pflag.BoolVarP(&jsonOut, "json", "j", false, "print plan as JSON")
	pflag.BoolVarP(&yes, "yes", "y", false, "execute deletions (otherwise dry-run)")
	pflag.Parse()

	if showHelp {
		printUsage()
		return
	}

	if showVersion {
		fmt.Println(version)
		return
	}

	// Best-effort update check: kicks off a short background fetch and
	// prints a one-line notice on stderr when a newer release is found.
	// Skipped for --json output, when the user opts out, and on
	// non-interactive stderr (CI, redirects).
	updateResult := startUpdateCheck(jsonOut)
	defer printUpdateNotice(updateResult)

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

	count, err := uipkg.PrintPlan(plan, uipkg.Options{JSON: jsonOut})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if count == 0 {
		return
	}

	// Interactive confirm if not --yes
	if !yes {
		ok, err := uipkg.ConfirmDeletion(count)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		if !ok {
			return
		}
	}

	// Execute deletions; --yes or confirmed implies force-delete (-D)
	res, err := sweeppkg.ExecuteDeletions(ctx, r, plan, sweeppkg.ExecuteOptions{MaxParallel: 0, ForceDelete: true})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := uipkg.PrintDeletionResult(res); err != nil {
		fmt.Println("error:", err)
	}
}

// startUpdateCheck runs the GitHub release lookup in the background and
// returns a channel that yields at most one result. The channel is closed
// when the goroutine finishes (or immediately when the check is skipped).
func startUpdateCheck(jsonOut bool) <-chan updatepkg.Result {
	out := make(chan updatepkg.Result, 1)
	go func() {
		defer close(out)
		if shouldSkipUpdateCheck(jsonOut) {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		out <- updatepkg.Check(ctx, updatepkg.Options{Repo: updateCheckRepo})
	}()
	return out
}

// printUpdateNotice waits briefly for the background check to finish and,
// if a newer release is available, prints a one-line notice to stderr.
// We never block the user for more than a short grace period: if the
// background check has not completed by then, this run silently skips the
// notice and the cache will likely be warm next time.
func printUpdateNotice(ch <-chan updatepkg.Result) {
	select {
	case res := <-ch:
		if !updatepkg.IsNewer(res.LatestTag, version) {
			return
		}
		fmt.Fprintf(os.Stderr, "\ngit-sweep: a newer version %s is available (you have %s).\n", res.LatestTag, version)
		fmt.Fprintln(os.Stderr, "  Linux/macOS: curl -fsSL https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.sh | bash")
		fmt.Fprintln(os.Stderr, "  Windows:     irm https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.ps1 | iex")
	case <-time.After(250 * time.Millisecond):
	}
}

func shouldSkipUpdateCheck(jsonOut bool) bool {
	if jsonOut {
		return true
	}
	if v := os.Getenv("GIT_SWEEP_NO_UPDATE_CHECK"); v != "" && v != "0" {
		return true
	}
	return !isStderrTTY()
}

func isStderrTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
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
	fmt.Println("    -h, --help              show this help")
}
