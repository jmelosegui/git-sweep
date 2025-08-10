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
func PrintPlan(plan sweep.Plan, opts Options) error {
	if opts.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(plan)
	}
	w := os.Stdout
	if _, err := fmt.Fprintf(w, "Repository: %s\n", plan.RepoRoot); err != nil {
		return err
	}
	if plan.Remote != "" {
		if _, err := fmt.Fprintf(w, "Remote: %s\n", plan.Remote); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Current: %s (upstream: %s)\n\n", plan.CurrentBranch, plan.CurrentUpstream); err != nil {
		return err
	}

	if len(plan.Candidates) == 0 {
		if _, err := fmt.Fprintln(w, "No local branches with gone upstreams found."); err != nil {
			return err
		}
		return nil
	}
	if _, err := fmt.Fprintln(w, "Branches to delete (gone upstream):"); err != nil {
		return err
	}
	for _, b := range plan.Candidates {
		// For now we just print the branch name. Stage 6+ annotations can be added here
		// when we include remote-merge checks in the plan (e.g., appears merged: yes/no).
		if _, err := fmt.Fprintf(w, "  - %s\n", b.Name); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "\nTotal: %d\n", len(plan.Candidates)); err != nil {
		return err
	}
	return nil
}
