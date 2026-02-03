// Package tools provides tool definitions and implementations for the agent.
package tools

import (
	"path/filepath"
	"strings"
)

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
type ToolFunc func(args map[string]any) string

// Tool represents a function the LLM can call.
type Tool struct {
	Description string
	Schema      map[string]any
	Func        ToolFunc
}

// Context provides tools with access to runner state.
type Context struct {
	WorkDir string
	// TodoState is set by the runner for the todo tool
	TodoState *TodoState
	// TodoEvents receives the latest todo list when it changes.
	TodoEvents chan<- []TodoItem
}

// Registry holds all available tools.
type Registry struct {
	ctx   *Context
	tools map[string]Tool
}

// NewRegistry creates a new tool registry with all tools.
func NewRegistry(ctx *Context) *Registry {
	r := &Registry{
		ctx:   ctx,
		tools: make(map[string]Tool),
	}
	r.registerAll()
	return r
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools.
func (r *Registry) All() map[string]Tool {
	return r.tools
}

// registerAll registers all tools in the registry.
func (r *Registry) registerAll() {
	r.tools["read"] = r.makeReadTool()
	r.tools["write"] = r.makeWriteTool()
	r.tools["edit"] = r.makeEditTool()
	r.tools["multiedit"] = r.makeMultieditTool()
	r.tools["glob"] = r.makeGlobTool()
	r.tools["grep"] = r.makeGrepTool()
	r.tools["bash"] = r.makeBashTool()
	r.tools["ls"] = r.makeLsTool()
	r.tools["todos"] = r.makeTodoTool()
}

// resolvePath resolves a path relative to the work directory.
func (r *Registry) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(r.ctx.WorkDir, path)
}

// MakeSchema converts Tool definitions to API-compatible ToolDef.
func MakeSchema(tools map[string]Tool) []ToolDef {
	result := []ToolDef{}
	for name, tool := range tools {
		props := map[string]any{}
		required := []string{}

		for paramName, paramType := range tool.Schema {
			if schemaMap, ok := paramType.(map[string]any); ok {
				props[paramName] = schemaMap
				required = append(required, paramName)
				continue
			}

			typeStr, ok := paramType.(string)
			if !ok {
				continue
			}

			baseType := strings.TrimSuffix(typeStr, "?")

			resultType := "string"
			if baseType == "number" {
				resultType = "integer"
			}
			if baseType == "boolean" {
				resultType = "boolean"
			}

			props[paramName] = map[string]any{"type": resultType}

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
