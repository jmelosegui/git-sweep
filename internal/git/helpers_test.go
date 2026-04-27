package git

import (
	"context"
	"errors"
	"testing"
)

type stubRunner struct {
	stdout string
	stderr string
	err    error
}

func (s stubRunner) Run(_ context.Context, _ ...string) (Result, error) {
	return Result{Stdout: s.stdout, Stderr: s.stderr}, s.err
}

func TestIsInsideWorkTree_NotARepository(t *testing.T) {
	r := stubRunner{
		stderr: "fatal: not a git repository (or any of the parent directories): .git",
		err:    errors.New("git rev-parse --is-inside-work-tree: exit code 128: exit status 128"),
	}
	inside, err := IsInsideWorkTree(context.Background(), r)
	if inside {
		t.Fatalf("expected inside=false, got true")
	}
	if !errors.Is(err, ErrNotGitRepository) {
		t.Fatalf("expected ErrNotGitRepository, got %v", err)
	}
}

func TestIsInsideWorkTree_OtherFailurePassesThrough(t *testing.T) {
	wantErr := errors.New("boom")
	r := stubRunner{stderr: "permission denied", err: wantErr}
	inside, err := IsInsideWorkTree(context.Background(), r)
	if inside {
		t.Fatalf("expected inside=false, got true")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected underlying error to propagate, got %v", err)
	}
	if errors.Is(err, ErrNotGitRepository) {
		t.Fatalf("did not expect ErrNotGitRepository for unrelated failures")
	}
}

func TestIsInsideWorkTree_True(t *testing.T) {
	r := stubRunner{stdout: "true\n"}
	inside, err := IsInsideWorkTree(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inside {
		t.Fatalf("expected inside=true")
	}
}
