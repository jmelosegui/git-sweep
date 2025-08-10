package sweep

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jmelosegui/git-sweep/internal/git"
)

type fakeRunner struct{
	calls [][]string
	failOn map[string]error
}

func (f *fakeRunner) Run(_ context.Context, args ...string) (git.Result, error) {
	f.calls = append(f.calls, append([]string{}, args...))
	key := ""
	if len(args) > 0 {
		key = args[0]
	}
	if err, ok := f.failOn[key]; ok {
		return git.Result{ExitCode: 1}, err
	}
	return git.Result{ExitCode: 0}, nil
}

func TestExecuteDeletions_SkipsCurrentBranch(t *testing.T) {
	r := &fakeRunner{}
	plan := Plan{
		CurrentBranch: "main",
		Candidates: []git.Branch{{Name: "main"}},
	}
	res, err := ExecuteDeletions(context.Background(), r, plan, ExecuteOptions{MaxParallel: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res.Failed) != 1 {
		t.Fatalf("expected 1 failed, got %+v", res.Failed)
	}
}

func TestExecuteDeletions_ForceUsesCapitalD(t *testing.T) {
	r := &fakeRunner{}
	plan := Plan{
		CurrentBranch: "main",
		Candidates: []git.Branch{{Name: "feature/x"}},
	}
	_, err := ExecuteDeletions(context.Background(), r, plan, ExecuteOptions{MaxParallel: 1, ForceDelete: true})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(r.calls) == 0 {
		t.Fatalf("expected calls, got none")
	}
	want := []string{"branch", "-D", "feature/x"}
	if !reflect.DeepEqual(r.calls[0], want) {
		t.Fatalf("unexpected args: got %v want %v", r.calls[0], want)
	}
}

func TestExecuteDeletions_ReportsDeleteFailure(t *testing.T) {
	r := &fakeRunner{failOn: map[string]error{
		"branch": errors.New("boom"),
	}}
	plan := Plan{
		CurrentBranch: "main",
		Candidates: []git.Branch{{Name: "feature/y"}},
	}
	res, err := ExecuteDeletions(context.Background(), r, plan, ExecuteOptions{MaxParallel: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res.Failed) != 1 {
		t.Fatalf("expected 1 failure, got %+v", res.Failed)
	}
}
