package tools

import (
	"fmt"
	"os"
	"strings"
)

func (r *Registry) makeLsTool() Tool {
	return Tool{
		Description: "List directory contents",
		Schema: map[string]any{
			"path": "string?",
		},
		Func: func(args map[string]any) string {
			path := r.ctx.WorkDir
			if p, ok := args["path"].(string); ok {
				path = r.resolvePath(p)
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return fmt.Sprintf("error: %v", err)
			}

			var sb strings.Builder
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() {
					name += "/"
				}
				sb.WriteString(name + "\n")
			}
			return sb.String()
		},
	}
}
