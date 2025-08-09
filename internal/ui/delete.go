package ui

import (
	"fmt"
	"os"

	"github.com/jmelosegui/git-sweep/internal/sweep"
)

// PrintDeletionResult prints a summary of deletions.
func PrintDeletionResult(res sweep.Result) error {
	w := os.Stdout
	if len(res.Deleted) > 0 {
		if _, err := fmt.Fprintf(w, "Deleted %d branch(es):\n", len(res.Deleted)); err != nil {
			return err
		}
		for _, name := range res.Deleted {
			if _, err := fmt.Fprintf(w, "  - %s\n", name); err != nil {
				return err
			}
		}
	}
	if len(res.Failed) > 0 {
		if _, err := fmt.Fprintf(w, "Failures (%d):\n", len(res.Failed)); err != nil {
			return err
		}
		for name, err := range res.Failed {
			if _, e := fmt.Fprintf(w, "  - %s: %v\n", name, err); e != nil {
				return e
			}
		}
	}
	return nil
}
