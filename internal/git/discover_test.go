package git

import "testing"

func TestParseForEachRef(t *testing.T) {
	input := "" +
		"feature/foo\torigin/main\t[ahead 1]\n" +
		"bugfix/bar\torigin/main\t[behind 2]\n" +
		"old/baz\torigin/main\t[gone]\n"

	branches := parseForEachRef(input)
	if len(branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(branches))
	}
	if !branches[2].IsGone {
		t.Fatalf("expected third branch to be gone, got %+v", branches[2])
	}
}

func TestParseBranchVV(t *testing.T) {
	input := "" +
		"  feature/foo 1234abcd [ahead 1] message\n" +
		"* main       deadbeef [origin/main] message\n" +
		"  old/baz     cafe0000 [gone] message\n"

	branches := parseBranchVV(input)
	if len(branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(branches))
	}
	if !branches[2].IsGone {
		t.Fatalf("expected third branch to be gone, got %+v", branches[2])
	}
}
