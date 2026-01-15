package agent

import "strings"

// ToolDef is the schema for a single tool, sent to the LLM.
type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines tool parameters in JSON Schema format.
type InputSchema struct {
	Type       string         `json:"type"`       // Always "object"
	Properties map[string]any `json:"properties"` // Param name -> {type: "..."}
	Required   []string       `json:"required"`   // Names of required params
}

// ToolFunc is the signature for all tool implementations.
// Takes arguments from the LLM and returns the result as a string.
type ToolFunc func(map[string]any) string

// Tool represents a function the LLM can call.
// - Description: What the tool does (shown to LLM)
// - Schema: Parameter types and whether optional (trailing "?" means optional)
// - Func: The actual Go function to execute
type Tool struct {
	Description string
	Schema      map[string]any
	Func        ToolFunc
}

// makeSchema converts our Tool definitions to API-compatible ToolDef.
// This maps our simple "string?" notation to JSON Schema format.
func makeSchema(tools map[string]Tool) []ToolDef {
	result := []ToolDef{}
	for name, tool := range tools {
		props := map[string]any{}
		required := []string{}

		// Process each parameter in the tool's schema
		for paramName, paramType := range tool.Schema {
			typeStr, ok := paramType.(string)
			if !ok {
				continue
			}

			// Remove "?" suffix to get base type
			baseType := strings.TrimSuffix(typeStr, "?")

			// Convert our type names to JSON Schema types
			resultType := "string"
			if baseType == "number" {
				resultType = "integer"
			}
			if baseType == "boolean" {
				resultType = "boolean"
			}

			props[paramName] = map[string]any{"type": resultType}

			// If no "?" suffix, parameter is required
			if !strings.HasSuffix(typeStr, "?") {
				required = append(required, paramName)
			}
		}

		result = append(result, ToolDef{
			Name:        name,
			Description: tool.Description,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: props,
				Required:   required,
			},
		})
	}
	return result
}
