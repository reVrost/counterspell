package codex

import "encoding/json"

// ThreadEvent is a top-level JSONL event emitted by codex exec.
type ThreadEvent struct {
	Type     string
	ThreadID string
	Usage    *Usage
	Error    *ThreadError
	Message  string
	Item     ThreadItem
	ItemRaw  map[string]any
	Raw      map[string]any
}

// Usage describes token usage during a turn.
type Usage struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens"`
	OutputTokens      int `json:"output_tokens"`
}

// ThreadError describes a failure for a turn.
type ThreadError struct {
	Message string `json:"message"`
}

// ParseThreadEvent parses a JSONL line into a ThreadEvent.
func ParseThreadEvent(line string) (ThreadEvent, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return ThreadEvent{}, err
	}
	event := ThreadEvent{Raw: raw}
	event.Type = getString(raw, "type")
	event.ThreadID = getString(raw, "thread_id")
	if usageRaw, ok := raw["usage"].(map[string]any); ok {
		var usage Usage
		if data, err := json.Marshal(usageRaw); err == nil {
			_ = json.Unmarshal(data, &usage)
			event.Usage = &usage
		}
	}
	if errorRaw, ok := raw["error"].(map[string]any); ok {
		var threadError ThreadError
		if data, err := json.Marshal(errorRaw); err == nil {
			_ = json.Unmarshal(data, &threadError)
			event.Error = &threadError
		}
	}
	if msg, ok := raw["message"].(string); ok {
		event.Message = msg
	}
	if itemRaw, ok := raw["item"].(map[string]any); ok {
		event.ItemRaw = itemRaw
		if item, err := ParseThreadItem(itemRaw); err == nil {
			event.Item = item
		}
	}
	return event, nil
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}
