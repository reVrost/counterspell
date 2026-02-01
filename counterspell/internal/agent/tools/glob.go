package tools

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (r *Registry) makeGlobTool() Tool {
	return Tool{
		Description: "Find files by pattern",
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

			fullPat := filepath.Join(basePath, pat)
			matches, err := filepath.Glob(fullPat)
			if err != nil {
				return "error: " + err.Error()
			}

			type fileInfo struct {
				path  string
				mtime time.Time
			}
			fileInfos := []fileInfo{}
			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					continue
				}
				fileInfos = append(fileInfos, fileInfo{match, info.ModTime()})
			}

			sort.Slice(fileInfos, func(i, j int) bool {
				return fileInfos[i].mtime.After(fileInfos[j].mtime)
			})

			if len(fileInfos) == 0 {
				return "none"
			}

			var sb strings.Builder
			for _, fi := range fileInfos {
				rel, _ := filepath.Rel(r.ctx.WorkDir, fi.path)
				if rel == "" {
					rel = fi.path
				}
				sb.WriteString(rel + "\n")
			}
			return sb.String()
		},
	}
}
