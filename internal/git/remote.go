package git

import (
	"context"
	"strings"
)

// RemoteDefaultRef returns the remote HEAD ref like "origin/main" for the given remote.
// It runs: git symbolic-ref refs/remotes/<remote>/HEAD
func RemoteDefaultRef(ctx context.Context, r Runner, remote string) (string, error) {
	if strings.TrimSpace(remote) == "" {
		remote = "origin"
	}
	res, err := r.Run(ctx, "symbolic-ref", "refs/remotes/"+remote+"/HEAD")
	if err != nil {
		return "", err
	}
	full := strings.TrimSpace(res.Stdout) // e.g., refs/remotes/origin/main
	const prefix = "refs/remotes/"
	if !strings.HasPrefix(full, prefix) {
		return "", nil
	}
	return strings.TrimPrefix(full, prefix), nil // e.g., origin/main
}

// IsAncestor reports whether commit-ish a is an ancestor of commit-ish b.
// It runs: git merge-base --is-ancestor a b
func IsAncestor(ctx context.Context, r Runner, a, b string) (bool, error) {
	_, err := r.Run(ctx, "merge-base", "--is-ancestor", a, b)
	if err != nil {
		// Non-zero exit means not ancestor; treat other errors the same
		return false, nil
	}
	return true, nil
}
