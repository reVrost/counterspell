package tools

import (
	"fmt"
	"os"
	"strings"
)

func (r *Registry) makeEditTool() Tool {
	return Tool{
		Description: "Replace old with new in file",
		Schema: map[string]any{
			"path": "string",
			"old":  "string",
			"new":  "string",
			"all":  "boolean?",
		},
		Func: func(args map[string]any) string {
			path := r.resolvePath(args["path"].(string))
			oldStr := args["old"].(string)
			newStr := args["new"].(string)

			doAll := false
			if a, ok := args["all"].(bool); ok {
				doAll = a
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			text := string(data)

			if !strings.Contains(text, oldStr) {
				return "error: old_string not found"
			}

			count := strings.Count(text, oldStr)
			if !doAll && count > 1 {
				return fmt.Sprintf("error: old_string appears %d times, use all=true", count)
			}

			var replacement string
			if doAll {
				replacement = strings.ReplaceAll(text, oldStr, newStr)
			} else {
				replacement = strings.Replace(text, oldStr, newStr, 1)
			}

			if err := os.WriteFile(path, []byte(replacement), 0644); err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			return "ok"
		},
	}
}
