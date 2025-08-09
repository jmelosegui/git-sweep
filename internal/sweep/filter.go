package sweep

import (
	"regexp"
	"sort"

	"github.com/jmelosegui/git-sweep/internal/git"
)

// FilterOptions controls how branches are selected for deletion.
// Include/Exclude are optional regex patterns applied to branch names.
// ProtectedNames are exact matches that must never be deleted.
type FilterOptions struct {
	IncludePattern  string
	ExcludePattern  string
	ProtectedNames  []string
	ProtectCurrent  bool
	ProtectUpstream bool
}

// SelectBranchesToDelete returns branches that are marked gone and pass filters/protections.
func SelectBranchesToDelete(branches []git.Branch, current string, currentUpstream string, opts FilterOptions) ([]git.Branch, error) {
	var includeRe, excludeRe *regexp.Regexp
	var err error
	if opts.IncludePattern != "" {
		includeRe, err = regexp.Compile(opts.IncludePattern)
		if err != nil {
			return nil, err
		}
	}
	if opts.ExcludePattern != "" {
		excludeRe, err = regexp.Compile(opts.ExcludePattern)
		if err != nil {
			return nil, err
		}
	}

	protected := make(map[string]struct{}, len(opts.ProtectedNames)+2)
	for _, n := range opts.ProtectedNames {
		protected[n] = struct{}{}
	}
	if opts.ProtectCurrent && current != "" {
		protected[current] = struct{}{}
	}
	if opts.ProtectUpstream && currentUpstream != "" {
		protected[currentUpstream] = struct{}{}
	}

	var selected []git.Branch
	for _, b := range branches {
		if !b.IsGone {
			continue
		}
		if _, isProt := protected[b.Name]; isProt {
			continue
		}
		if includeRe != nil && !includeRe.MatchString(b.Name) {
			continue
		}
		if excludeRe != nil && excludeRe.MatchString(b.Name) {
			continue
		}
		selected = append(selected, b)
	}

	sort.Slice(selected, func(i, j int) bool { return selected[i].Name < selected[j].Name })
	return selected, nil
}
