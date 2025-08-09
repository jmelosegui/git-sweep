package git

import (
	"context"
	"strings"
)

// CurrentBranch returns the current branch short name (e.g., "main").
// If in detached HEAD, it returns "HEAD".
func CurrentBranch(ctx context.Context, r Runner) (string, error) {
	res, err := r.Run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res.Stdout), nil
}

// BranchUpstream returns the short upstream name for the given branch (e.g., "origin/main").
// If the branch has no upstream, it returns an empty string and nil error.
func BranchUpstream(ctx context.Context, r Runner, branch string) (string, error) {
	// Use symbolic-full-name to ensure we get short form; on some setups, --abbrev-ref shortens it
	res, err := r.Run(ctx, "rev-parse", "--abbrev-ref", "--symbolic-full-name", branch+"@{upstream}")
	if err != nil {
		// No upstream configured or other error; treat as no-upstream when output empty
		return "", nil
	}
	return strings.TrimSpace(res.Stdout), nil
}

// DefaultProtectedNames returns the baseline protected branch names.
func DefaultProtectedNames() []string {
	return []string{"main", "master", "develop"}
}
