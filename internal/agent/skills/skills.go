// Package skills provides embedded skill instructions for the agent.
// Skills are markdown files with frontmatter that provide detailed guidance
// for complex multi-step operations.
package skills

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed *.md
var skillsFS embed.FS

// skillsCache is loaded once at init time from embedded files
var skillsCache map[string]string

func init() {
	skillsCache = make(map[string]string)

	// Load all skills from embedded filesystem
	entries, err := skillsFS.ReadDir(".")
	if err != nil {
		// If skills directory doesn't exist or can't be read, use fallback
		skillsCache["merge"] = defaultMergeSkill()
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Read the skill file
		content, err := skillsFS.ReadFile(entry.Name())
		if err != nil {
			continue
		}

		// Extract skill name from filename (remove .md extension)
		skillName := strings.TrimSuffix(entry.Name(), ".md")
		skillsCache[skillName] = string(content)
	}

	// Ensure merge skill is available even if file loading fails
	if _, ok := skillsCache["merge"]; !ok {
		skillsCache["merge"] = defaultMergeSkill()
	}
}

// Get returns a skill by name.
func Get(name string) (string, error) {
	skill, ok := skillsCache[name]
	if !ok {
		available := List()
		return "", fmt.Errorf("skill '%s' not found. Available skills: %s", name, strings.Join(available, ", "))
	}
	return skill, nil
}

// List returns all available skill names.
func List() []string {
	names := make([]string, 0, len(skillsCache))
	for name := range skillsCache {
		names = append(names, name)
	}
	return names
}

// Register allows registering a new skill at runtime (for testing/extensibility).
func Register(name, instructions string) {
	skillsCache[name] = instructions
}

// defaultMergeSkill provides a fallback if the embedded file can't be loaded
func defaultMergeSkill() string {
	return `You are a merge specialist. Your task is to merge the current task's changes into the main branch.

Follow these steps:

1. DETECT REPO TYPE
   Run: ls -la
   Check for .git directory → git repo
   Check for .jj directory → jj repo
   If neither exists, this is not a version-controlled workspace

2. INITIALIZE GIT IF NEEDED
   If no VCS found:
   - Run: git init
   - Run: git add -A
   - Run: git commit -m "Initial commit"

3. FOR GIT REPOSITORIES:
   a. Get current branch: git branch --show-current
   b. Checkout main: git checkout main || git checkout master
   c. Pull latest: git pull origin main || git pull origin master
   d. Merge: git merge <branch-name> --no-edit
   e. If conflicts: read files, resolve with edit tool, git add, git commit
   f. Push: git push origin main || git push origin master
   g. Cleanup: git branch -d <branch-name>

4. FOR JJ REPOSITORIES:
   a. Get change ID: jj log -n 1 --template 'change_id'
   b. IMPORTANT: If @ has changes, run: jj new
   c. Squash: jj squash --from <change-id> --into @
   d. Set bookmark: jj bookmark set main -r @
   e. Push: jj git push --bookmark main
   f. Cleanup: jj bookmark delete <bookmark>, jj workspace forget <workspace>, rm -rf ../jjws-*

5. REPORT RESULTS
   Always report what was done, any conflicts resolved, and final status.`
}
