package git

import "context"

// DeleteLocalBranch deletes a local branch using `git branch -d <branch>`.
// It never force deletes; callers should handle non-merged errors.
func DeleteLocalBranch(ctx context.Context, r Runner, branch string) error {
	_, err := r.Run(ctx, "branch", "-d", branch)
	return err
}
