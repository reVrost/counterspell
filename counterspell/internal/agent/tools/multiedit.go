package tools

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed multiedit.md
var multieditDescription string

// FailedEdit represents an edit that failed to apply.
type FailedEdit struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

func (r *Registry) makeMultieditTool() Tool {
	return Tool{
		Description: multieditDescription,
		Schema: map[string]any{
			"file_path": "string",
			"edits": map[string]any{
				"type":        "array",
				"description": "Array of edit operations to perform sequentially on the file",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"old_string": map[string]any{
							"type":        "string",
							"description": "Text to replace (must match exactly including whitespace/indentation)",
						},
						"new_string": map[string]any{
							"type":        "string",
							"description": "Replacement text",
						},
						"replace_all": map[string]any{
							"type":        "boolean",
							"description": "Replace all occurrences (optional, defaults to false)",
						},
					},
					"required": []string{"old_string", "new_string"},
				},
			},
		},
		Func: r.toolMultiedit,
	}
}

func (r *Registry) toolMultiedit(args map[string]any) string {
	// Get file path
	filePathRaw, ok := args["file_path"]
	if !ok {
		return "error: file_path is required"
	}
	filePath := r.resolvePath(filePathRaw.(string))

	// Get edits array
	editsRaw, ok := args["edits"]
	if !ok {
		return "error: edits array is required"
	}
	editsArr, ok := editsRaw.([]any)
	if !ok {
		return "error: edits must be an array"
	}
	if len(editsArr) == 0 {
		return "error: at least one edit operation is required"
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("error: file not found: %s", filePath)
		}
		return fmt.Sprintf("error: failed to read file: %v", err)
	}

	currentContent := string(data)
	originalContent := currentContent

	// Track results
	var failedEdits []FailedEdit
	editsApplied := 0

	// Apply edits sequentially
	for i, editRaw := range editsArr {
		editMap, ok := editRaw.(map[string]any)
		if !ok {
			failedEdits = append(failedEdits, FailedEdit{
				Index: i + 1,
				Error: "edit must be an object",
			})
			continue
		}

		oldString, _ := editMap["old_string"].(string)
		newString, _ := editMap["new_string"].(string)
		replaceAll := false
		if ra, ok := editMap["replace_all"].(bool); ok {
			replaceAll = ra
		}

		// Apply the edit
		newContent, err := applyEdit(currentContent, oldString, newString, replaceAll)
		if err != nil {
			failedEdits = append(failedEdits, FailedEdit{
				Index: i + 1,
				Error: err.Error(),
			})
			continue
		}

		currentContent = newContent
		editsApplied++
	}

	// Check if anything changed
	if currentContent == originalContent {
		if len(failedEdits) > 0 {
			failedJSON, _ := json.Marshal(failedEdits)
			return fmt.Sprintf("error: no changes made - all %d edit(s) failed\nfailed_edits: %s", len(failedEdits), string(failedJSON))
		}
		return "error: no changes made - all edits resulted in identical content"
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(currentContent), 0644); err != nil {
		return fmt.Sprintf("error: failed to write file: %v", err)
	}

	// Build response
	if len(failedEdits) > 0 {
		failedJSON, _ := json.Marshal(failedEdits)
		return fmt.Sprintf("Applied %d of %d edits (%d failed)\nfailed_edits: %s",
			editsApplied, len(editsArr), len(failedEdits), string(failedJSON))
	}

	return fmt.Sprintf("Applied %d edits successfully", editsApplied)
}

// applyEdit applies a single edit operation to content.
func applyEdit(content, oldString, newString string, replaceAll bool) (string, error) {
	if oldString == "" {
		return "", fmt.Errorf("old_string cannot be empty")
	}

	if !strings.Contains(content, oldString) {
		return "", fmt.Errorf("old_string not found in content")
	}

	if replaceAll {
		return strings.ReplaceAll(content, oldString, newString), nil
	}

	// Check for multiple occurrences
	count := strings.Count(content, oldString)
	if count > 1 {
		return "", fmt.Errorf("old_string appears %d times, use replace_all=true or add more context", count)
	}

	return strings.Replace(content, oldString, newString, 1), nil
}
