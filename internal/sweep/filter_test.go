package sweep

import (
	"testing"

	"github.com/jmelosegui/git-sweep/internal/git"
)

func TestSelectBranchesToDelete(t *testing.T) {
	branches := []git.Branch{
		{Name: "feature/a", IsGone: true},
		{Name: "feature/b", IsGone: true},
		{Name: "hotfix/x", IsGone: true},
		{Name: "main", IsGone: true},
		{Name: "develop", IsGone: false},
	}

	opts := FilterOptions{
		IncludePattern:  "^feature/",
		ExcludePattern:  "b$",
		ProtectedNames:  []string{"main"},
		ProtectCurrent:  true,
		ProtectUpstream: true,
	}

	selected, err := SelectBranchesToDelete(branches, "hotfix/x", "origin/main", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 1 || selected[0].Name != "feature/a" {
		t.Fatalf("unexpected selection: %#v", selected)
	}
}
