package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmelosegui/git-sweep/internal/sweep"
)

// Options controls how UI prints information.
type Options struct {
	JSON bool
}

// PrintPlan prints a sweep.Plan either as JSON or a human-readable summary.
// It returns the number of candidates printed in the human-readable mode.
func PrintPlan(plan sweep.Plan, opts Options) (int, error) {
	if opts.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return 0, enc.Encode(plan)
	}
	w := os.Stdout
	if _, err := fmt.Fprintf(w, "On branch %s\n", plan.CurrentBranch); err != nil {
		return 0, err
	}
	if plan.CurrentUpstream != "" {
		if _, err := fmt.Fprintf(w, "Your branch is up to date with '%s'.\n\n", plan.CurrentUpstream); err != nil {
			return 0, err
		}
	} else {
		if _, err := fmt.Fprintln(w, "(no upstream configured)"); err != nil {
			return 0, err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return 0, err
		}
	}

	if len(plan.Candidates) == 0 {
		if _, err := fmt.Fprintln(w, "nothing to sweep, local branches are clean"); err != nil {
			return 0, err
		}
		return 0, nil
	}
	if _, err := fmt.Fprintln(w, "The following local branches have a gone upstream:"); err != nil {
		return 0, err
	}
	for _, b := range plan.Candidates {
		if _, err := fmt.Fprintf(w, "  %s\n", b.Name); err != nil {
			return 0, err
		}
	}
	if _, err := fmt.Fprintf(w, "\n(%d to delete)\n", len(plan.Candidates)); err != nil {
		return 0, err
	}
	return len(plan.Candidates), nil
}
