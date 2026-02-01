package tools

import (
	"fmt"
	"os"
	"strings"
)

func (r *Registry) makeReadTool() Tool {
	return Tool{
		Description: "Read file with line numbers",
		Schema: map[string]any{
			"path":   "string",
			"offset": "number?",
			"limit":  "number?",
		},
		Func: func(args map[string]any) string {
			path := r.resolvePath(args["path"].(string))

			offset := 0
			if o, ok := args["offset"].(float64); ok {
				offset = int(o)
			}
			limit := 0
			if l, ok := args["limit"].(float64); ok {
				limit = int(l)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			lines := strings.Split(string(data), "\n")

			if limit == 0 {
				limit = len(lines)
			}

			end := min(offset+limit, len(lines))

			var sb strings.Builder
			for i := offset; i < end; i++ {
				fmt.Fprintf(&sb, "%4d| %s\n", i+1, lines[i])
			}
			return sb.String()
		},
	}
}
