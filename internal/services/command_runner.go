package services

import (
	"context"
	"os"
	"os/exec"
)

// CommandRunner abstracts running external commands for testability.
type CommandRunner interface {
	Run(ctx context.Context, dir, name string, args ...string) ([]byte, error)
}

// ExecCommandRunner executes commands using os/exec.
type ExecCommandRunner struct{}

func (ExecCommandRunner) Run(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	// Clear pager-related environment variables to prevent tools like delta from being used
	cmd.Env = removeEnvVars(os.Environ(), "PAGER", "GIT_PAGER", "DELTA_FEATURES")
	return cmd.CombinedOutput()
}

// removeEnvVars returns the environment with the specified variables removed
func removeEnvVars(env []string, vars ...string) []string {
	var result []string
	for _, e := range env {
		keep := true
		for _, v := range vars {
			if len(e) >= len(v) && e[:len(v)] == v && (len(e) == len(v) || e[len(v)] == '=') {
				keep = false
				break
			}
		}
		if keep {
			result = append(result, e)
		}
	}
	return result
}
