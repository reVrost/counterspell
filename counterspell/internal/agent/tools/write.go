package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

func (r *Registry) makeWriteTool() Tool {
	return Tool{
		Description: "Write content to file",
		Schema: map[string]any{
			"path":    "string",
			"content": "string",
		},
		Func: func(args map[string]any) string {
			path := r.resolvePath(args["path"].(string))
			content := args["content"].(string)

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Sprintf("error: %v", err)
			}

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			return "ok"
		},
	}
}
