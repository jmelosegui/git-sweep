package git

import (
	"context"
)

// Result contains the outputs and exit code of a git command execution.
// Stdout and Stderr are captured fully as strings; callers should trim them when needed.
// ExitCode is 0 on success; when non-zero, the returned error will describe the failure.
// Context cancellations are reported via the returned error.
//
// The intent is to avoid exposing os/exec types to calling packages to keep
// the API stable and easy to mock in tests.
//
// High-verbosity code is used intentionally for clarity and maintainability.
// See code style guidelines.
//
// Note: Large outputs are acceptable for our use cases (branch listings). If
// extremely large outputs are needed in the future, consider a streaming API.
// This design optimizes for simplicity at this stage.
//
//nolint:revive // exported fields with clear names
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Runner defines the minimal interface required to execute git commands.
// All calls must be non-interactive and avoid shell features.
// Implementations should be concurrency-safe if used across goroutines.
// The working directory behavior is implementation-defined; ExecRunner
// allows setting it per-runner.
//
// Implementations should return a non-nil error when the command fails to
// start, is context-cancelled, or exits with a non-zero status. ExitCode in
// the Result should reflect the actual exit status when available.
//
// Commands should not print to the process stdout/stderr; all output goes into Result.
type Runner interface {
	Run(ctx context.Context, args ...string) (Result, error)
}
