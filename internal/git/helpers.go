package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrNotGitRepository is returned when a git command fails because the
// working directory is not inside a git repository. The error message
// mirrors Git's own "fatal: not a git repository" output so it can be
// surfaced to users without further translation.
var ErrNotGitRepository = errors.New("fatal: not a git repository (or any of the parent directories): .git")

// Version returns the output of `git --version`.
func Version(ctx context.Context, r Runner) (string, error) {
	res, err := r.Run(ctx, "--version")
	if err != nil {
		return "", err
	}
	return res.Stdout, nil
}

// IsInsideWorkTree reports whether the working directory is inside a git work tree.
// It calls: git rev-parse --is-inside-work-tree
func IsInsideWorkTree(ctx context.Context, r Runner) (bool, error) {
	res, err := r.Run(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if strings.Contains(res.Stderr, "not a git repository") {
			return false, ErrNotGitRepository
		}
		return false, err
	}
	s := strings.TrimSpace(strings.ToLower(res.Stdout))
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	return false, fmt.Errorf("unexpected output for rev-parse --is-inside-work-tree: %q", res.Stdout)
}

// RepoRoot returns the absolute path of the repository root directory.
// It calls: git rev-parse --show-toplevel
func RepoRoot(ctx context.Context, r Runner) (string, error) {
	res, err := r.Run(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res.Stdout), nil
}
