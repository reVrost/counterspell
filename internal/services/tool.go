package services

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// ToolService executes tool operations for agents.
type ToolService struct {
	dataDir    string
	fileService *FileService
}

// NewToolService creates a new tool service.
func NewToolService(dataDir string) *ToolService {
	return &ToolService{
		dataDir:    dataDir,
		fileService: NewFileService(dataDir),
	}
}

// ExecuteTool executes a tool operation.
func (s *ToolService) ExecuteTool(ctx context.Context, tool string, input map[string]any) (string, error) {
	slog.Info("[TOOL] Executing", "tool", tool)

	switch tool {
	case "bash":
		return s.executeBash(ctx, input)
	case "read":
		return s.executeRead(ctx, input)
	case "write":
		return s.executeWrite(ctx, input)
	case "list":
		return s.executeList(ctx, input)
	case "search":
		return s.executeSearch(ctx, input)
	case "grep":
		return s.executeGrep(ctx, input)
	default:
		return "", fmt.Errorf("unknown tool: %s", tool)
	}
}

// executeBash executes a shell command.
func (s *ToolService) executeBash(ctx context.Context, input map[string]any) (string, error) {
	command, ok := input["command"].(string)
	if !ok {
		return "", fmt.Errorf("command required")
	}

	slog.Info("[TOOL] Executing bash", "command", command)

	// Create command
	var cmd *exec.Cmd
	if strings.Contains(command, ";") || strings.Contains(command, "&&") {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	} else {
		parts := strings.Fields(command)
		cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	}

	// Execute
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// executeRead reads a file.
func (s *ToolService) executeRead(ctx context.Context, input map[string]any) (string, error) {
	path, ok := input["path"].(string)
	if !ok {
		return "", fmt.Errorf("path required")
	}

	content, err := s.fileService.Read(ctx, path)
	if err != nil {
		return "", err
	}

	return content, nil
}

// executeWrite writes to a file.
func (s *ToolService) executeWrite(ctx context.Context, input map[string]any) (string, error) {
	path, ok := input["path"].(string)
	if !ok {
		return "", fmt.Errorf("path required")
	}
	content, ok := input["content"].(string)
	if !ok {
		return "", fmt.Errorf("content required")
	}

	if err := s.fileService.Write(ctx, path, content); err != nil {
		return "", err
	}

	return fmt.Sprintf("Written to %s", path), nil
}

// executeList lists files in a directory.
func (s *ToolService) executeList(ctx context.Context, input map[string]any) (string, error) {
	directory := ""
	if d, ok := input["directory"].(string); ok {
		directory = d
	}

	files, err := s.fileService.List(ctx, directory)
	if err != nil {
		return "", err
	}

	// Format output
	var builder strings.Builder
	for _, file := range files {
		if file.IsDir {
			builder.WriteString(fmt.Sprintf("[DIR]  %s\n", file.Name))
		} else {
			builder.WriteString(fmt.Sprintf("[FILE] %s (%d bytes)\n", file.Name, file.Size))
		}
	}

	return builder.String(), nil
}

// executeSearch searches for files.
func (s *ToolService) executeSearch(ctx context.Context, input map[string]any) (string, error) {
	pattern := ""
	if p, ok := input["pattern"].(string); ok {
		pattern = p
	}
	directory := ""
	if d, ok := input["directory"].(string); ok {
		directory = d
	}

	files, err := s.fileService.Search(ctx, pattern, directory, 20)
	if err != nil {
		return "", err
	}

	// Format output
	var builder strings.Builder
	for _, file := range files {
		builder.WriteString(fmt.Sprintf("%s\n", file.Path))
	}

	return builder.String(), nil
}

// executeGrep searches for text in files.
func (s *ToolService) executeGrep(ctx context.Context, input map[string]any) (string, error) {
	pattern, ok := input["pattern"].(string)
	if !ok {
		return "", fmt.Errorf("pattern required")
	}
	path := "."
	if p, ok := input["path"].(string); ok {
		path = p
	}

	slog.Info("[TOOL] Grep", "pattern", pattern, "path", path)

	// Use find/grep combination
	cmd := exec.CommandContext(ctx, "find", path, "-type", "f", "-exec", "grep", "-l", pattern, "{}", "+")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("grep failed: %w", err)
	}

	return string(output), nil
}

// GetAvailableTools returns list of available tools.
func (s *ToolService) GetAvailableTools() []string {
	return []string{
		"bash",
		"read",
		"write",
		"list",
		"search",
		"grep",
	}
}
