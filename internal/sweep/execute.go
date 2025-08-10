package sweep

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/jmelosegui/git-sweep/internal/git"
)

// ExecuteOptions controls how deletion is performed.
type ExecuteOptions struct {
	MaxParallel int
	ForceDelete bool // when true, use `git branch -D` instead of `-d`
}

// Result holds per-branch deletion outcomes.
type Result struct {
	Deleted []string
	Failed  map[string]error
}

// ExecuteDeletions deletes the selected branches with safety checks.
// - Never deletes the current branch
// - Uses `git branch -d` by default; can use -D when ForceDelete is true
// - Runs with bounded parallelism
func ExecuteDeletions(ctx context.Context, r git.Runner, plan Plan, execOpts ExecuteOptions) (Result, error) {
	if execOpts.MaxParallel <= 0 {
		execOpts.MaxParallel = maxInt(2, runtime.NumCPU())
	}

	res := Result{Failed: make(map[string]error)}
	if len(plan.Candidates) == 0 {
		return res, nil
	}

	sem := make(chan struct{}, execOpts.MaxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, b := range plan.Candidates {
		branchName := b.Name
		if branchName == plan.CurrentBranch {
			mu.Lock()
			res.Failed[branchName] = errors.New("refusing to delete current branch")
			mu.Unlock()
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			var err error
			if execOpts.ForceDelete {
				_, err = r.Run(ctx, "branch", "-D", branchName)
			} else {
				err = git.DeleteLocalBranch(ctx, r, branchName)
			}
			if err != nil {
				mu.Lock()
				res.Failed[branchName] = fmt.Errorf("delete failed: %w", err)
				mu.Unlock()
				return
			}
			mu.Lock()
			res.Deleted = append(res.Deleted, branchName)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return res, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
