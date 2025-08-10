// Package integration contains end-to-end tests that exercise git-sweep against
// real Git repositories created in temporary directories. These tests verify
// the full flow: remote/local setup → discovery of gone upstreams → safe deletion.
package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	gitpkg "github.com/jmelosegui/git-sweep/internal/git"
	sweeppkg "github.com/jmelosegui/git-sweep/internal/sweep"
)

// TestDiscoverGoneAndDeleteMergedBranch proves that a branch whose upstream
// was deleted remotely is discovered as "gone" and can be safely deleted with
// `git branch -d` when it has been merged into the default branch. The test:
// 1) Creates a bare remote and a local repo.
// 2) Creates and pushes a feature branch.
// 3) Merges the feature branch into main (so `-d` is allowed later).
// 4) Deletes the remote feature branch to simulate a gone upstream.
// 5) Builds a plan and executes deletions; expects the local feature branch to be removed.
func TestDiscoverGoneAndDeleteMergedBranch(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Ensure git is present; skip gracefully if not on PATH in CI.
		if _, err := exec.LookPath("git"); err != nil {
			t.Skip("git not available in PATH")
		}
	}

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tmp := t.TempDir()
	remotePath := filepath.Join(tmp, "remote.git")
	localPath := filepath.Join(tmp, "local")

	runGit(t, tmp, "init", "--bare", remotePath)
	mustMkdir(t, localPath)
	runGit(t, localPath, "init")
	runGit(t, localPath, "config", "user.name", "Test User")
	runGit(t, localPath, "config", "user.email", "test@example.com")

	writeFile(t, filepath.Join(localPath, "README.md"), "hello\n")
	runGit(t, localPath, "add", ".")
	runGit(t, localPath, "commit", "-m", "init")
	runGit(t, localPath, "branch", "-M", "main")

	remoteURL := toFileURL(remotePath)
	runGit(t, localPath, "remote", "add", "origin", remoteURL)
	runGit(t, localPath, "push", "-u", "origin", "main")

	// Create feature branch, commit, push
	runGit(t, localPath, "checkout", "-b", "feat/a")
	writeFile(t, filepath.Join(localPath, "feature.txt"), "feature\n")
	runGit(t, localPath, "add", ".")
	runGit(t, localPath, "commit", "-m", "feat commit")
	runGit(t, localPath, "push", "-u", "origin", "feat/a")

	// Merge into main so -d is allowed later
	runGit(t, localPath, "checkout", "main")
	runGit(t, localPath, "merge", "--no-ff", "-m", "merge feat/a", "feat/a")
	runGit(t, localPath, "push", "origin", "main")

	// Delete remote branch to simulate gone upstream
	runGit(t, localPath, "push", "origin", ":feat/a")

	// Build plan and execute
	r := gitpkg.ExecRunner{WorkDir: localPath}
	plan, err := sweeppkg.BuildPlan(ctx, r, sweeppkg.Options{Remote: "origin", ProtectCurrent: true, ProtectUpstream: true})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	// Expect feat/a to appear
	found := false
	for _, b := range plan.Candidates {
		if b.Name == "feat/a" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected feat/a in candidates, got %+v", plan.Candidates)
	}

	res, err := sweeppkg.ExecuteDeletions(ctx, r, plan, sweeppkg.ExecuteOptions{})
	if err != nil {
		t.Fatalf("ExecuteDeletions error: %v", err)
	}
	if len(res.Failed) != 0 {
		t.Fatalf("unexpected failures: %+v", res.Failed)
	}
	// Verify branch is gone locally
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/feat/a")
	cmd.Dir = localPath
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected local branch to be deleted")
	}
}

// runGit executes a git command in the given directory and fails the test with
// a helpful message (including combined output) on error. This avoids hiding
// errors that would otherwise appear only in subprocess output.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

// writeFile writes a file with standard permissions and fails the test on error.
func writeFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

// mustMkdir creates a directory tree and fails the test on error.
func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

// toFileURL converts a filesystem path to a file:// URL usable by Git remotes.
// Handles both POSIX and Windows paths (C:/...).
func toFileURL(p string) string {
	p = filepath.ToSlash(p)
	if strings.HasPrefix(p, "/") {
		return "file://" + p
	}
	// Windows paths like C:/...
	return "file:///" + p
}
