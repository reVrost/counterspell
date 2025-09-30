package clearcast

import "time"

type Session struct {
	RootAgentID   string         `json:"root_agent_id"`
	Workspace     map[string]any `json:"workspace"`
	Mission       string         `json:"mission"`
	Memory        map[string]any `json:"memory"`
	Messages      []Message      `json:"messages"`
	CreatedAt     time.Time      `json:"created_at"`
	MaxIterations int            `json:"max_iterations"`

	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost"`
	Tools        []*Tool `json:"tools"`
}

type Message struct {
	Role    string
	Content string
	// Summaryy is the summary of all the messages up to this point, act as a compaction
	Summary string
}

// to map[string]any
func (s *Session) ToMap() map[string]any {
	result := map[string]any{
		"workspace":      s.Workspace,
		"mission":        s.Mission,
		"memory":         s.Memory,
		"messages":       s.Messages,
		"created_at":     s.CreatedAt,
		"max_iterations": s.MaxIterations,

		"input_tokens":  s.InputTokens,
		"output_tokens": s.OutputTokens,
		"cost":          s.Cost,
		"tools":         s.Tools,
	}

	// Merge workspace fields at top level for template access
	for k, v := range s.Workspace {
		result[k] = v
	}

	return result
}
