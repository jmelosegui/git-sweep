package git

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// FetchPrune runs `git fetch --prune` for the given remote (default origin if empty).
func FetchPrune(ctx context.Context, r Runner, remote string) error {
	remoteArg := remote
	if strings.TrimSpace(remoteArg) == "" {
		remoteArg = "origin"
	}
	_, err := r.Run(ctx, "fetch", "--prune", remoteArg)
	return err
}

// ListLocalBranches returns local branches with their upstream and tracking status.
// Prefer `for-each-ref` for structured output; fallback to parsing `git branch -vv` if needed.
func ListLocalBranches(ctx context.Context, r Runner) ([]Branch, error) {
	// Try for-each-ref with a custom format capturing: name, upstream, and upstream:track
	// %1: short refname; %2: upstream short; %3: upstream:track status
	format := "%(refname:short)\t%(upstream:short)\t%(upstream:track)"
	res, err := r.Run(ctx, "for-each-ref", "--format="+format, "refs/heads")
	if err == nil && strings.TrimSpace(res.Stdout) != "" {
		return parseForEachRef(res.Stdout), nil
	}
	// Fallback to `git branch -vv` parsing
	res, err = r.Run(ctx, "branch", "-vv")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	return parseBranchVV(res.Stdout), nil
}

func parseForEachRef(output string) []Branch {
	var branches []Branch
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, ln := range lines {
		parts := strings.SplitN(ln, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		upstream := strings.TrimSpace(parts[1])
		track := strings.TrimSpace(parts[2])
		branches = append(branches, Branch{
			Name:     name,
			Upstream: upstream,
			Track:    track,
			IsGone:   strings.Contains(track, "[gone]"),
		})
	}
	return branches
}

var branchVVRe = regexp.MustCompile(`^\*?\s*(?P<branch>\S+)\s+\S+\s+\[(?P<track>[^\]]+)\]`) // e.g., "feature 1234abcd [gone] msg"

func parseBranchVV(output string) []Branch {
	var branches []Branch
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, ln := range lines {
		m := branchVVRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		branch := strings.TrimSpace(m[1])
		track := strings.TrimSpace(m[2])
		branches = append(branches, Branch{
			Name:     branch,
			Upstream: "", // Not available from -vv reliably without more parsing
			Track:    track,
			IsGone:   strings.Contains(track, "gone"),
		})
	}
	return branches
}
