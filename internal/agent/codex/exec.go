package codex

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type ExecArgs struct {
	Input string

	BaseURL  string
	APIKey   string
	ThreadID string
	Images   []string

	Model                 string
	SandboxMode           SandboxMode
	WorkingDirectory      string
	AdditionalDirectories []string
	SkipGitRepoCheck      bool
	OutputSchemaFile      string
	ModelReasoningEffort  ModelReasoningEffort
	NetworkAccessEnabled  *bool
	WebSearchMode         WebSearchMode
	WebSearchEnabled      *bool
	ApprovalPolicy        ApprovalMode
	ExtraArgs             []string
}

type Exec struct {
	executablePath  string
	envOverride     map[string]string
	configOverrides ConfigObject
}

const (
	internalOriginatorEnv = "CODEX_INTERNAL_ORIGINATOR_OVERRIDE"
	goSDKOriginator       = "codex_sdk_go"
)

func NewExec(executablePath string, envOverride map[string]string, configOverrides ConfigObject) *Exec {
	path := executablePath
	if path == "" {
		path = "codex"
	}
	return &Exec{
		executablePath:  path,
		envOverride:     envOverride,
		configOverrides: configOverrides,
	}
}

type LineStream struct {
	Lines <-chan string
	Done  <-chan error
}

func (e *Exec) Run(ctx context.Context, args ExecArgs) *LineStream {
	lines := make(chan string, 64)
	done := make(chan error, 1)

	go func() {
		defer close(lines)
		defer close(done)

		cmdArgs := []string{"exec", "--experimental-json"}

		if e.configOverrides != nil {
			overrides, err := serializeConfigOverrides(e.configOverrides)
			if err != nil {
				done <- err
				return
			}
			for _, override := range overrides {
				cmdArgs = append(cmdArgs, "--config", override)
			}
		}

		if args.Model != "" {
			cmdArgs = append(cmdArgs, "--model", args.Model)
		}
		if args.SandboxMode != "" {
			cmdArgs = append(cmdArgs, "--sandbox", string(args.SandboxMode))
		}
		if args.WorkingDirectory != "" {
			cmdArgs = append(cmdArgs, "--cd", args.WorkingDirectory)
		}
		for _, dir := range args.AdditionalDirectories {
			if strings.TrimSpace(dir) == "" {
				continue
			}
			cmdArgs = append(cmdArgs, "--add-dir", dir)
		}
		if args.SkipGitRepoCheck {
			cmdArgs = append(cmdArgs, "--skip-git-repo-check")
		}
		if args.OutputSchemaFile != "" {
			cmdArgs = append(cmdArgs, "--output-schema", args.OutputSchemaFile)
		}
		if args.ModelReasoningEffort != "" {
			cmdArgs = append(cmdArgs, "--config", fmt.Sprintf("model_reasoning_effort=%q", args.ModelReasoningEffort))
		}
		if args.NetworkAccessEnabled != nil {
			cmdArgs = append(cmdArgs, "--config", fmt.Sprintf("sandbox_workspace_write.network_access=%t", *args.NetworkAccessEnabled))
		}
		if args.WebSearchMode != "" {
			cmdArgs = append(cmdArgs, "--config", fmt.Sprintf("web_search=%q", args.WebSearchMode))
		} else if args.WebSearchEnabled != nil {
			if *args.WebSearchEnabled {
				cmdArgs = append(cmdArgs, "--config", "web_search=\"live\"")
			} else {
				cmdArgs = append(cmdArgs, "--config", "web_search=\"disabled\"")
			}
		}
		if args.ApprovalPolicy != "" {
			cmdArgs = append(cmdArgs, "--config", fmt.Sprintf("approval_policy=%q", args.ApprovalPolicy))
		}
		for _, image := range args.Images {
			if strings.TrimSpace(image) == "" {
				continue
			}
			cmdArgs = append(cmdArgs, "--image", image)
		}
		if args.ThreadID != "" {
			cmdArgs = append(cmdArgs, "resume", args.ThreadID)
		}
		if len(args.ExtraArgs) > 0 {
			cmdArgs = append(cmdArgs, args.ExtraArgs...)
		}

		cmd := exec.CommandContext(ctx, e.executablePath, cmdArgs...)

		env := map[string]string{}
		if e.envOverride != nil {
			for key, value := range e.envOverride {
				env[key] = value
			}
		} else {
			for _, entry := range os.Environ() {
				parts := strings.SplitN(entry, "=", 2)
				if len(parts) == 2 {
					env[parts[0]] = parts[1]
				}
			}
		}
		if _, ok := env[internalOriginatorEnv]; !ok {
			env[internalOriginatorEnv] = goSDKOriginator
		}
		if args.BaseURL != "" {
			env["OPENAI_BASE_URL"] = args.BaseURL
		}
		if args.APIKey != "" {
			env["CODEX_API_KEY"] = args.APIKey
		}
		cmd.Env = mapToEnv(env)

		stdin, err := cmd.StdinPipe()
		if err != nil {
			done <- err
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			done <- err
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			done <- err
			return
		}

		if err := cmd.Start(); err != nil {
			done <- err
			return
		}

		if _, err := io.WriteString(stdin, args.Input); err != nil {
			_ = stdin.Close()
			_ = cmd.Wait()
			done <- err
			return
		}
		_ = stdin.Close()

		var stderrMu sync.Mutex
		var stderrChunks []string
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				stderrMu.Lock()
				stderrChunks = append(stderrChunks, line)
				stderrMu.Unlock()
			}
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		scanErr := scanner.Err()

		err = cmd.Wait()
		wg.Wait()

		if scanErr != nil {
			done <- scanErr
			return
		}
		if err != nil {
			stderrMu.Lock()
			stderrContent := strings.Join(stderrChunks, "\n")
			stderrMu.Unlock()
			if stderrContent != "" {
				done <- fmt.Errorf("%w: %s", err, stderrContent)
				return
			}
			done <- err
			return
		}
		done <- nil
	}()

	return &LineStream{Lines: lines, Done: done}
}

func mapToEnv(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for key, value := range env {
		result = append(result, key+"="+value)
	}
	return result
}
