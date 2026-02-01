package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (r *Registry) makeBashTool() Tool {
	return Tool{
		Description: "Run shell command",
		Schema: map[string]any{
			"cmd": "string",
		},
		Func: func(args map[string]any) string {
			cmdStr := args["cmd"].(string)

			cmd := exec.Command("bash", "-c", cmdStr)
			cmd.Dir = r.ctx.WorkDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String() + stderr.String()

			if err != nil {
				output += fmt.Sprintf("\n(exit: %v)", err)
			}

			if strings.TrimSpace(output) == "" {
				return "(empty)"
			}
			return output
		},
	}
}
