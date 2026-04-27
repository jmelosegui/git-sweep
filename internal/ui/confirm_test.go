package ui

import "testing"

func TestInterpretConfirm(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"  y  \n", true},
		{"\n", false},
		{"", false},
		{"n\n", false},
		{"no\n", false},
		{"asdf\n", false},
		{"yse\n", false},
	}
	for _, c := range cases {
		if got := interpretConfirm(c.input); got != c.want {
			t.Errorf("interpretConfirm(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
