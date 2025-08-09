package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ExecRunner executes git by invoking the host git binary via os/exec without a shell.
// It is safe for cross-platform use and honors context cancellation.
//
// ExecRunner assumes `git` is on PATH. If not, Run returns an error from exec.LookPath/Command.
type ExecRunner struct {
	// WorkDir, if non-empty, sets the working directory for git commands.
	WorkDir string
}

// Run executes `git` with the provided arguments. Output is captured and returned in Result.
// When git exits non-zero, an error is returned and Result.ExitCode is set when available.
func (r ExecRunner) Run(ctx context.Context, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	res := Result{
		Stdout:   strings.TrimRight(stdoutBuf.String(), "\n"),
		Stderr:   strings.TrimRight(stderrBuf.String(), "\n"),
		ExitCode: 0,
	}

	var exitErr *exec.ExitError
	if err != nil {
		if errors.As(err, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
			return res, fmt.Errorf("git %v: exit code %d: %w", strings.Join(args, " "), exitErr.ExitCode(), err)
		}
		return res, fmt.Errorf("git %v: %w", strings.Join(args, " "), err)
	}

	return res, nil
}
