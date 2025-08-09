package sweep

import "testing"

func TestProtectedNamesFromEnv(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"main, develop ; hotfix/x", []string{"main", "develop", "hotfix/x"}},
		{"  ", nil},
	}
	for _, c := range cases {
		got := ProtectedNamesFromEnv(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("len mismatch: got %v want %v", got, c.want)
		}
	}
}

func TestMergeProtectedNames(t *testing.T) {
	base := []string{"main", "master"}
	extra := []string{"develop", "main"}
	merged := MergeProtectedNames(base, extra)
	if len(merged) != 3 {
		t.Fatalf("unexpected merge size: %v", merged)
	}
}
