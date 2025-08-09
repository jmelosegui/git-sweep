package git

// Branch represents a local branch and basic information about its upstream.
// Upstream may be empty if none is configured.
// Track contains Git's tracking status string (e.g., "[gone]", "[ahead 1]", "[behind 2]").
// IsGone is true when the upstream remote ref has been deleted ("[gone]").
// The Name and Upstream are short names (e.g., "feature/foo", "origin/main").
//
//nolint:revive // exported fields with clear descriptive names
type Branch struct {
	Name     string
	Upstream string
	Track    string
	IsGone   bool
}
