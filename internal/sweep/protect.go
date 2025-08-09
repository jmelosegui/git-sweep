package sweep

import (
	"os"
	"strings"
)

// ProtectedEnvVar is the environment variable used to extend protected branches.
// Value is a comma- or semicolon-separated list of branch names.
const ProtectedEnvVar = "GIT_SWEEP_PROTECTED"

// ProtectedNamesFromEnv parses the provided string (typically os.Getenv(ProtectedEnvVar))
// into a slice of branch names, splitting on commas/semicolons and trimming spaces.
func ProtectedNamesFromEnv(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	// Support both comma and semicolon separators
	value = strings.ReplaceAll(value, ";", ",")
	parts := strings.Split(value, ",")
	var out []string
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	return out
}

// ProtectedNamesFromEnvVar reads ProtectedEnvVar from the process environment.
func ProtectedNamesFromEnvVar() []string {
	return ProtectedNamesFromEnv(os.Getenv(ProtectedEnvVar))
}

// MergeProtectedNames returns a deduplicated union of base and extra names.
func MergeProtectedNames(base []string, extra []string) []string {
	set := make(map[string]struct{}, len(base)+len(extra))
	for _, n := range base {
		if n == "" {
			continue
		}
		set[n] = struct{}{}
	}
	for _, n := range extra {
		if n == "" {
			continue
		}
		set[n] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	return out
}
