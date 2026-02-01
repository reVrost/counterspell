package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (r *Registry) makeGrepTool() Tool {
	return Tool{
		Description: "Search files for regex pattern",
		Schema: map[string]any{
			"pat":  "string",
			"path": "string?",
		},
		Func: func(args map[string]any) string {
			pat := args["pat"].(string)
			basePath := r.ctx.WorkDir
			if p, ok := args["path"].(string); ok {
				basePath = r.resolvePath(p)
			}

			re, err := regexp.Compile(pat)
			if err != nil {
				return fmt.Sprintf("error: invalid regex: %v", err)
			}

			hits := []string{}
			_ = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if strings.Contains(path, "/.git/") || strings.Contains(path, "/node_modules/") {
					return nil
				}

				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				lines := strings.Split(string(data), "\n")
				for lineNum, line := range lines {
					if re.MatchString(line) {
						rel, _ := filepath.Rel(r.ctx.WorkDir, path)
						if rel == "" {
							rel = path
						}
						hits = append(hits, fmt.Sprintf("%s:%d:%s", rel, lineNum+1, strings.TrimSpace(line)))
					}
				}
				return nil
			})

			if len(hits) == 0 {
				return "none"
			}

			if len(hits) > 50 {
				hits = hits[:50]
			}
			return strings.Join(hits, "\n")
		},
	}
}
