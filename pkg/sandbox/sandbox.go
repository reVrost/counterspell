// Package sandbox provides secure command execution using bubblewrap (Linux) or pass-through (macOS).
package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/config"
)

// ErrOutputTruncated indicates the output was truncated due to size limits.
var ErrOutputTruncated = errors.New("output truncated")

// ErrTimeout indicates the command timed out.
var ErrTimeout = errors.New("command timed out")

// Result contains the output of a sandboxed command.
type Result struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Truncated bool
	TimedOut  bool
	Duration  time.Duration
}

// CombinedOutput returns stdout and stderr combined.
func (r *Result) CombinedOutput() string {
	if r.Stderr == "" {
		return r.Stdout
	}
	if r.Stdout == "" {
		return r.Stderr
	}
	return r.Stdout + "\n" + r.Stderr
}

// Executor executes commands with optional sandboxing.
type Executor struct {
	cfg            *config.Config
	bwrapAvailable bool
}

// NewExecutor creates a new sandbox executor.
func NewExecutor(cfg *config.Config) *Executor {
	bwrapAvailable := checkBwrapAvailable()
	if runtime.GOOS == "linux" && !bwrapAvailable {
		slog.Warn("Bubblewrap (bwrap) not found - sandboxing disabled on Linux!")
	}

	return &Executor{
		cfg:            cfg,
		bwrapAvailable: bwrapAvailable,
	}
}

// checkBwrapAvailable checks if bubblewrap is installed.
func checkBwrapAvailable() bool {
	_, err := exec.LookPath("bwrap")
	return err == nil
}

// Execute runs a command with sandboxing if available and appropriate.
//
// Parameters:
//   - ctx: Context for cancellation
//   - workDir: Working directory for the command
//   - cmd: Command to run (will be executed via bash -c)
//   - allowlist: If true, skip sandboxing for allowed commands
func (e *Executor) Execute(ctx context.Context, workDir, cmd string, allowlist bool) (*Result, error) {
	start := time.Now()

	// Check if command is in allowlist
	if allowlist && e.isAllowed(cmd) {
		slog.Debug("Running allowed command without sandbox", "cmd", truncateCmd(cmd, 50))
		return e.executeDirect(ctx, workDir, cmd, start)
	}

	// On Linux with bwrap available, use sandboxing
	if runtime.GOOS == "linux" && e.bwrapAvailable {
		slog.Debug("Running command in sandbox", "cmd", truncateCmd(cmd, 50))
		return e.executeSandboxed(ctx, workDir, cmd, start)
	}

	// On macOS or Linux without bwrap, run directly (dev mode)
	if runtime.GOOS == "darwin" {
		slog.Debug("Running command without sandbox (macOS dev mode)", "cmd", truncateCmd(cmd, 50))
	} else {
		slog.Warn("Running command without sandbox (bwrap unavailable)", "cmd", truncateCmd(cmd, 50))
	}
	return e.executeDirect(ctx, workDir, cmd, start)
}

// isAllowed checks if the command's base executable is in the allowlist.
func (e *Executor) isAllowed(cmd string) bool {
	// Extract first word (the command itself)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	baseCmd := parts[0]
	// Handle paths like /usr/bin/git -> git
	if idx := strings.LastIndex(baseCmd, "/"); idx >= 0 {
		baseCmd = baseCmd[idx+1:]
	}

	return e.cfg.IsCommandAllowed(baseCmd)
}

// executeDirect runs a command without sandboxing.
func (e *Executor) executeDirect(ctx context.Context, workDir, cmdStr string, start time.Time) (*Result, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.cfg.SandboxTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{buf: &stdout, limit: e.cfg.SandboxOutputLimit}
	cmd.Stderr = &limitedWriter{buf: &stderr, limit: e.cfg.SandboxOutputLimit}

	err := cmd.Run()
	duration := time.Since(start)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result, ErrTimeout
	}

	// Check for exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	// Check for truncation
	if lw, ok := cmd.Stdout.(*limitedWriter); ok && lw.truncated {
		result.Truncated = true
	}
	if lw, ok := cmd.Stderr.(*limitedWriter); ok && lw.truncated {
		result.Truncated = true
	}

	return result, nil
}

// executeSandboxed runs a command inside a bubblewrap sandbox.
func (e *Executor) executeSandboxed(ctx context.Context, workDir, cmdStr string, start time.Time) (*Result, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.cfg.SandboxTimeout)
	defer cancel()

	// Build bwrap arguments
	bwrapArgs := []string{
		// Die when parent dies
		"--die-with-parent",
		// New PID namespace
		"--unshare-pid",
		// Read-only bind mounts for system directories
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/etc", "/etc",
		// Writable /tmp
		"--tmpfs", "/tmp",
		// Bind mount the workspace
		"--bind", workDir, "/workspace",
		// Set working directory
		"--chdir", "/workspace",
		// Dev null for /dev
		"--dev", "/dev",
		// Proc filesystem
		"--proc", "/proc",
		// Run bash
		"bash", "-c", cmdStr,
	}

	cmd := exec.CommandContext(ctx, "bwrap", bwrapArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{buf: &stdout, limit: e.cfg.SandboxOutputLimit}
	cmd.Stderr = &limitedWriter{buf: &stderr, limit: e.cfg.SandboxOutputLimit}

	err := cmd.Run()
	duration := time.Since(start)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result, ErrTimeout
	}

	// Check for exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute sandboxed command: %w", err)
		}
	}

	// Check for truncation
	if lw, ok := cmd.Stdout.(*limitedWriter); ok && lw.truncated {
		result.Truncated = true
	}
	if lw, ok := cmd.Stderr.(*limitedWriter); ok && lw.truncated {
		result.Truncated = true
	}

	return result, nil
}

// IsSandboxed returns true if sandboxing is available on this system.
func (e *Executor) IsSandboxed() bool {
	return runtime.GOOS == "linux" && e.bwrapAvailable
}

// truncateCmd truncates a command string for logging.
func truncateCmd(cmd string, maxLen int) string {
	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen] + "..."
}

// limitedWriter wraps a buffer and enforces a size limit.
type limitedWriter struct {
	buf       *bytes.Buffer
	limit     int64
	written   int64
	truncated bool
}

func (w *limitedWriter) Write(p []byte) (n int, err error) {
	if w.truncated {
		return len(p), nil // Discard but don't error
	}

	remaining := w.limit - w.written
	if int64(len(p)) > remaining {
		// Write what we can
		if remaining > 0 {
			n, _ = w.buf.Write(p[:remaining])
			w.written += int64(n)
		}
		w.truncated = true
		w.buf.WriteString("\n... [output truncated] ...")
		return len(p), nil
	}

	n, err = w.buf.Write(p)
	w.written += int64(n)
	return n, err
}
