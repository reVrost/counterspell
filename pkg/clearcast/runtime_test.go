package clearcast

import (
	"testing"
)

func TestStripMarkdownCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"tool": "search", "params": {"q": "test"}}`,
			expected: `{"tool": "search", "params": {"q": "test"}}`,
		},
		{
			name:  "JSON with ```json blocks",
			input: "```json\n{\n  \"tool\": \"search\",\n  \"params\": {\"q\": \"test\"}\n}\n```",
			expected: `{
  "tool": "search",
  "params": {"q": "test"}
}`,
		},
		{
			name:  "JSON with ``` blocks",
			input: "```\n{\n  \"tool\": \"search\"\n}\n```",
			expected: `{
  "tool": "search"
}`,
		},
		{
			name:     "JSON with whitespace",
			input:    "  \n  {\"tool\": \"search\"}  \n  ",
			expected: `{"tool": "search"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripMarkdownCodeBlock(tt.input)
			if result != tt.expected {
				t.Errorf("stripMarkdownCodeBlock() = %q, want %q", result, tt.expected)
			}
		})
	}
}
