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
				return findClosestMatch(text, oldStr)
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

// findClosestMatch searches the file content for the best fuzzy match to the target string.
// Returns an error message with helpful debugging information.
func findClosestMatch(content, target string) string {
	lines := strings.Split(content, "\n")
	targetLines := strings.Split(target, "\n")
	targetLineCount := len(targetLines)

	var bestMatch struct {
		text       string
		startLine  int
		similarity float64
	}

	// Slide a window of targetLineCount lines through the file
	for i := 0; i <= len(lines)-targetLineCount; i++ {
		window := strings.Join(lines[i:i+targetLineCount], "\n")
		sim := similarity(window, target)
		if sim > bestMatch.similarity {
			bestMatch.text = window
			bestMatch.startLine = i + 1
			bestMatch.similarity = sim
		}
	}

	// Also check individual lines for single-line targets
	if targetLineCount == 1 {
		for i, line := range lines {
			sim := similarity(line, target)
			if sim > bestMatch.similarity {
				bestMatch.text = line
				bestMatch.startLine = i + 1
				bestMatch.similarity = sim
			}
		}
	}

	if bestMatch.similarity < 0.3 {
		return "error: old_string not found (no similar text found in file)"
	}

	// Build error message with helpful diff info
	msg := fmt.Sprintf("error: old_string not found\nClosest match at line %d (%.0f%% similar):\n",
		bestMatch.startLine, bestMatch.similarity*100)

	// Show the closest match with quoting
	msg += fmt.Sprintf("  found: %q\n", truncate(bestMatch.text, 200))
	msg += fmt.Sprintf("  wanted: %q\n", truncate(target, 200))

	// Add specific hints about common differences
	hints := detectDifferences(bestMatch.text, target)
	if len(hints) > 0 {
		msg += "Hints:\n"
		for _, hint := range hints {
			msg += fmt.Sprintf("  - %s\n", hint)
		}
	}

	return msg
}

// similarity calculates a simple similarity ratio between two strings.
// Returns a value between 0.0 (completely different) and 1.0 (identical).
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Use longest common subsequence ratio for similarity
	lcs := longestCommonSubsequence(a, b)
	maxLen := max(len(a), len(b))
	return float64(lcs) / float64(maxLen)
}

// longestCommonSubsequence returns the length of the LCS of two strings.
func longestCommonSubsequence(a, b string) int {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return 0
	}

	// Use two rows instead of full matrix to save memory
	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = max(prev[j], curr[j-1])
			}
		}
		prev, curr = curr, prev
	}
	return prev[n]
}

// detectDifferences analyzes two strings and returns hints about their differences.
func detectDifferences(found, wanted string) []string {
	var hints []string

	// Check for whitespace-only differences
	if strings.TrimSpace(found) == strings.TrimSpace(wanted) {
		hints = append(hints, "content matches but whitespace differs (leading/trailing spaces or blank lines)")
	}

	// Check indentation differences
	foundIndent := countLeadingWhitespace(found)
	wantedIndent := countLeadingWhitespace(wanted)
	if foundIndent != wantedIndent {
		hints = append(hints, fmt.Sprintf("indentation differs: found %d spaces/tabs, wanted %d", foundIndent, wantedIndent))
	}

	// Check for tab vs space issues
	foundHasTabs := strings.Contains(found, "\t")
	wantedHasTabs := strings.Contains(wanted, "\t")
	if foundHasTabs != wantedHasTabs {
		if foundHasTabs {
			hints = append(hints, "file uses tabs but old_string uses spaces")
		} else {
			hints = append(hints, "file uses spaces but old_string uses tabs")
		}
	}

	// Check line ending differences
	foundCRLF := strings.Contains(found, "\r\n")
	wantedCRLF := strings.Contains(wanted, "\r\n")
	if foundCRLF != wantedCRLF {
		if foundCRLF {
			hints = append(hints, "file uses CRLF (\\r\\n) line endings but old_string uses LF (\\n)")
		} else {
			hints = append(hints, "file uses LF (\\n) line endings but old_string uses CRLF (\\r\\n)")
		}
	}

	// Check for trailing whitespace differences
	foundLines := strings.Split(found, "\n")
	wantedLines := strings.Split(wanted, "\n")
	for i := 0; i < min(len(foundLines), len(wantedLines)); i++ {
		if strings.TrimRight(foundLines[i], " \t") == strings.TrimRight(wantedLines[i], " \t") &&
			foundLines[i] != wantedLines[i] {
			hints = append(hints, fmt.Sprintf("trailing whitespace differs on line %d", i+1))
			break
		}
	}

	// Check line count difference
	if len(foundLines) != len(wantedLines) {
		hints = append(hints, fmt.Sprintf("line count differs: found %d lines, wanted %d lines", len(foundLines), len(wantedLines)))
	}

	return hints
}

// countLeadingWhitespace returns the number of leading whitespace characters.
func countLeadingWhitespace(s string) int {
	for i, c := range s {
		if c != ' ' && c != '\t' {
			return i
		}
	}
	return len(s)
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
