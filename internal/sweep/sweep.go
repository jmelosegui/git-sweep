package sweep

import (
	"context"
	"fmt"

	"github.com/jmelosegui/git-sweep/internal/git"
)

// Options controls how sweep selects branches to delete. Deletion itself is not performed here.
// Remote is used for fetch --prune and discovery scoping where applicable.
// ExtraProtected extends the default protected names and environment-derived names.
type Options struct {
	Remote          string
	IncludePattern  string
	ExcludePattern  string
	ExtraProtected  []string
	ProtectCurrent  bool
	ProtectUpstream bool
}

// Plan contains the branches selected for deletion along with context information.
type Plan struct {
	RepoRoot        string
	Remote          string
	CurrentBranch   string
	CurrentUpstream string
	Candidates      []git.Branch
}

// BuildPlan discovers gone branches and filters them according to Options and protections.
func BuildPlan(ctx context.Context, r git.Runner, opts Options) (Plan, error) {
	var plan Plan

	inside, err := git.IsInsideWorkTree(ctx, r)
	if err != nil {
		return plan, err
	}
	if !inside {
		return plan, fmt.Errorf("not inside a git work tree")
	}

	root, err := git.RepoRoot(ctx, r)
	if err != nil {
		return plan, err
	}
	plan.RepoRoot = root

	// Fetch prune for the selected remote
	if err := git.FetchPrune(ctx, r, opts.Remote); err != nil {
		return plan, err
	}
	plan.Remote = opts.Remote

	// Discover branches
	branches, err := git.ListLocalBranches(ctx, r)
	if err != nil {
		return plan, err
	}

	// Protections
	current, err := git.CurrentBranch(ctx, r)
	if err != nil {
		return plan, err
	}
	upstream, _ := git.BranchUpstream(ctx, r, current) // no upstream is not an error

	plan.CurrentBranch = current
	plan.CurrentUpstream = upstream

	baseProtected := git.DefaultProtectedNames()
	envProtected := ProtectedNamesFromEnvVar()
	protected := MergeProtectedNames(baseProtected, envProtected)
	protected = MergeProtectedNames(protected, opts.ExtraProtected)

	selected, err := SelectBranchesToDelete(branches, current, upstream, FilterOptions{
		IncludePattern:  opts.IncludePattern,
		ExcludePattern:  opts.ExcludePattern,
		ProtectedNames:  protected,
		ProtectCurrent:  opts.ProtectCurrent,
		ProtectUpstream: opts.ProtectUpstream,
	})
	if err != nil {
		return plan, err
	}
	plan.Candidates = selected
	return plan, nil
}
