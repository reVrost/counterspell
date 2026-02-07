package tools

import (
	"fmt"
	"strings"

	"github.com/revrost/counterspell/internal/agent/skills"
)

// makeSkillTool creates a tool that recalls skill instructions by name.
func (r *Registry) makeSkillTool() Tool {
	return Tool{
		Description: "Recall a skill to get detailed instructions for a specific task. Skills provide step-by-step guidance for complex operations like merging, testing, or deployment without cluttering the chat history. Available skills: " + strings.Join(skills.List(), ", "),
		Schema: map[string]any{
			"name": "string",
		},
		Func: func(args map[string]any) string {
			name := args["name"].(string)

			skill, err := skills.Get(name)
			if err != nil {
				return fmt.Sprintf("error: %v", err)
			}

			return skill
		},
	}
}

// GetAvailableSkills returns a list of all available skill names.
func GetAvailableSkills() []string {
	return skills.List()
}

// RegisterSkill allows registering a new skill at runtime (for testing/extensibility).
func RegisterSkill(name, instructions string) {
	skills.Register(name, instructions)
}
