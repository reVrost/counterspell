package services

import (
	"context"
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
	return cmd.CombinedOutput()
}
