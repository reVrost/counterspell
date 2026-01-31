package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
)

// CloudflareTunnel represents a running cloudflared process.
type CloudflareTunnel struct {
	cmd *exec.Cmd
}

// StartCloudflare starts a cloudflared tunnel using a token.
func StartCloudflare(ctx context.Context, token, localURL, binaryPath string, logger *slog.Logger) (*CloudflareTunnel, error) {
	if token == "" {
		return nil, fmt.Errorf("missing tunnel token")
	}

	path := binaryPath
	if path == "" {
		p, err := exec.LookPath("cloudflared")
		if err != nil {
			return nil, fmt.Errorf("cloudflared not found in PATH")
		}
		path = p
	}

	args := []string{"tunnel", "--no-autoupdate", "run", "--token", token}
	if localURL != "" {
		args = append(args, "--url", localURL)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start cloudflared: %w", err)
	}

	logPipe(logger, "cloudflared", stdout)
	logPipe(logger, "cloudflared", stderr)

	return &CloudflareTunnel{cmd: cmd}, nil
}

// Stop terminates the cloudflared process.
func (t *CloudflareTunnel) Stop() error {
	if t == nil || t.cmd == nil || t.cmd.Process == nil {
		return nil
	}
	return t.cmd.Process.Kill()
}

func logPipe(logger *slog.Logger, label string, pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				logger.Info("tunnel", "provider", label, "line", line)
			}
		}
	}()
}
