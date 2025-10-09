package clearcast

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"strings"
	"text/template"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const AgentModePlan = "plan"
const AgentModeLoop = "loop"
const AgentModeSingle = "single"

const availableToolsPrompt = `
## Iterations guideline
	If you are one off to the max iterations, make the best out of what you have and end your plan with done: true to produce the final output.

CURRENT ITERATION: {{.iteration}}
MAX ITERATIONS: {{.max_iterations}}

{{if .next_hints}}
	## PREVIOUS HINTS
	{{.next_hints}}
{{end}}

## AVAILABLE TOOLS
{{range .tools}}
- **{{.id}}**: {{.description}}
	USAGE: {{.usage}}
{{end}}

## TOOL USAGE INSTRUCTIONS
You can use the available tools to gather information and accomplish your task. Follow these steps:

1. **To use a tool**, respond with ONLY a JSON object (no markdown, no code blocks):
   {
     "tool": "exact_tool_id_from_above",
     "params": {"param_name": "param_value"}
   }

   **IMPORTANT**: Use the EXACT tool ID as listed above (e.g., "serper_dev", NOT "search" or any other name)

**Critical Requirements:**
- Respond with PLAIN JSON only - do NOT wrap in markdown code blocks (no ` + "```" + `)
- Each response must be a single valid JSON object
- Use the EXACT tool IDs as listed in "AVAILABLE TOOLS" above
- Use tools iteratively to gather information
`

// DebugContextKey for context-based debug flag
type DebugContextKey struct{}

// AgentOption defines a functional option for configuring Prompt
type AgentOption func(*Agent)

// WithSlogLogger sets the slog.Logger for Prompt
func WithSlogLogger(logger *slog.Logger) AgentOption {
	return func(p *Agent) {
		p.logger = logger
	}
}

// WithOTEL enables or disables OTEL tracing
func WithOTEL(enabled bool) AgentOption {
	return func(p *Agent) {
		p.enableOTEL = enabled
		if enabled {
			p.tracer = otel.Tracer("clearcast/prompt")
		} else {
			p.tracer = nil
		}
	}
}

func WithDefaultTools(IDs ...string) AgentOption {
	defaultToolsMap := map[string]*Tool{
		"web_search": WebSearchTool(),
	}
	return func(r *Agent) {
		for _, ID := range IDs {
			if tool, ok := defaultToolsMap[ID]; ok {
				r.tools[ID] = tool
			}
		}
	}
}

func WithTools(tools ...*Tool) AgentOption {
	return func(r *Agent) {
		for _, tool := range tools {
			r.tools[tool.ID] = tool
		}
	}
}

func WithTool(tool *Tool) AgentOption {
	return func(r *Agent) {
		r.tools[tool.ID] = tool
	}
}

// ExecuteTool executes a tool by ID with the given parameters
func (a *Agent) ExecuteTool(ctx context.Context, toolID string, params map[string]any) (any, error) {
	tool, ok := a.tools[toolID]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolID)
	}
	return tool.Execute(ctx, params)
}

// GetTools returns the agent's tools
func (a *Agent) GetTools() Tools {
	return a.tools
}

// Agent struct with slog and OTEL toggle
type Agent struct {
	ID         string
	Prompt     string
	Iteration  int
	Model      string
	llm        LLMProvider
	logger     *slog.Logger
	tracer     trace.Tracer
	enableOTEL bool
	tools      Tools
}

// NewAgent creates a new Prompt with slog (JSON handler) and OTEL enabled by default
func NewAgent(id, model, prompt string, llm LLMProvider, opts ...AgentOption) *Agent {
	p := &Agent{
		ID:         id,
		Model:      model,
		Prompt:     prompt,
		llm:        llm,
		logger:     slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		tracer:     otel.Tracer("clearcast/prompt"),
		enableOTEL: true,
		tools:      make(Tools),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// recordUsageAttributes adds usage-related attributes to an OTEL span
func (t *Agent) recordUsageAttributes(span trace.Span, usage Usage) {
	span.SetAttributes(
		attribute.Int("usage.prompt_tokens", usage.PromptTokens),
		attribute.Int("usage.completion_tokens", usage.CompletionTokens),
		attribute.Int("usage.total_tokens", usage.TotalTokens),
		attribute.Int("usage.reasoning_tokens", usage.CompletionTokenDetails.ReasoningTokens),
		attribute.Int("usage.cached_tokens", usage.PromptTokenDetails.CachedTokens),
		attribute.Float64("cost.total", usage.Cost),
		attribute.Float64("cost.upstream_inference", usage.CostDetails.UpstreamInferenceCost),
	)
}

// debugLog checks if debug is enabled and returns keyvals for logging
func (t *Agent) debugLog(ctx context.Context, model string) []any {
	if debugEnabled, _ := ctx.Value(DebugContextKey{}).(bool); debugEnabled {
		return []any{"model", model}
	}
	return nil
}

func (t *Agent) render(ctx context.Context, promptTemplate string, values map[string]any) (string, error) {
	var span trace.Span
	if t.enableOTEL {
		ctx, span = t.tracer.Start(ctx, "Prompt.Render")
		defer span.End()
		span.SetAttributes(
			attribute.String("template.content", promptTemplate),
			attribute.Int("values.count", len(values)),
		)
	}

	parsedTmpl, err := template.New("template").
		Option("missingkey=error").
		Parse(promptTemplate)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Failed to parse template", "template", promptTemplate, "error", err)
		return "", err
	}

	sb := new(strings.Builder)
	if err := parsedTmpl.Execute(sb, values); err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Failed to execute template", "values", values, "error", err)
		return "", err
	}

	rendered := sb.String()
	if keyvals := t.debugLog(ctx, ""); keyvals != nil {
		t.logger.Debug("Template rendered successfully", append(keyvals, "rendered", rendered, "values", values)...)
	}

	return rendered, nil
}

type ResponseOption func(*ChatCompletionRequest)

func WithResponseFormat(format ResponseFormat) ResponseOption {
	return func(req *ChatCompletionRequest) {
		req.ResponseFormat = &format
	}
}

// Non-streaming completion
func (t *Agent) Step(ctx context.Context, args map[string]any, opts ...ResponseOption) (ChatCompletionResponse, error) {
	var span trace.Span
	if t.enableOTEL {
		ctx, span = t.tracer.Start(ctx, "Agent.Plan")
		defer span.End()
		span.SetAttributes(attribute.String("model", t.Model))
	}

	prompt := t.Prompt + availableToolsPrompt
	// convert t.tools to map[string]any
	maps.Copy(args, t.tools.toMap())

	rendered, err := t.render(ctx, prompt, args)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Error rendering template for plan completion", "model", t.Model, "error", err)
		return ChatCompletionResponse{}, fmt.Errorf("error rendering template: %w", err)
	}

	// Build messages: start with system prompt, then add session messages
	messages := []ChatMessage{SystemMessage(rendered)}

	req := ChatCompletionRequest{
		Model:    t.Model,
		Messages: messages,
	}
	for _, opt := range opts {
		opt(&req)
	}

	startTime := time.Now()
	resp, err := t.llm.ChatCompletion(ctx, req)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Chat completion failed", "model", t.Model, "error", err, "request", req)
		return ChatCompletionResponse{}, err
	}

	ttft := float64(time.Since(startTime).Nanoseconds()) / 1e6
	if t.enableOTEL {
		t.recordUsageAttributes(span, resp.Usage)
		span.SetAttributes(attribute.Float64("metrics.ttft_ms", ttft))
	}

	if keyvals := t.debugLog(ctx, t.Model); keyvals != nil {
		t.logger.Debug("Chat completion successful", append(keyvals, "response", resp, "ttft_ms", ttft, "usage", resp.Usage)...)
	}

	t.logger.Info("Chat completion successful", "response", resp, "ttft_ms", ttft, "usage", resp.Usage)

	return resp, nil
}

// Streaming completion
// func (t *Agent) Reflect(ctx context.Context, args map[string]any, opts ...ResponseOption) (<-chan ChatCompletionChunk, error) {
// 	var span trace.Span
// 	if t.enableOTEL {
// 		ctx, span = t.tracer.Start(ctx, "Agent.Reflect")
// 		defer span.End()
// 		span.SetAttributes(attribute.String("model", t.ReflectModel))
// 	}
//
// 	rendered, err := t.render(ctx, t.ReflectPrompt, args)
// 	if err != nil {
// 		if t.enableOTEL {
// 			span.RecordError(err)
// 		}
// 		t.logger.Error("Error rendering template for streaming chat completion", "model", t.ReflectModel, "error", err)
// 		return nil, fmt.Errorf("error rendering template: %w", err)
// 	}
//
// 	// Build messages: start with system prompt, then add session messages
// 	messages := []ChatMessage{SystemMessage(rendered)}
// 	req := ChatCompletionRequest{
// 		Model:    t.ReflectModel,
// 		Messages: messages,
// 	}
// 	for _, opt := range opts {
// 		opt(&req)
// 	}
//
// 	t.logger.Info("Reflecting", "prompt", rendered, "args", args)
// 	// startTime := time.Now()
// 	stream, err := t.llm.ChatCompletionStream(ctx, req)
// 	if err != nil {
// 		if t.enableOTEL {
// 			span.RecordError(err)
// 		}
// 		t.logger.Error("Streaming chat completion failed", "model", t.ReflectModel, "error", err)
// 		return nil, err
// 	}

// if keyvals := t.debugLog(ctx, t.ReflectModel); keyvals != nil {
// 	t.logger.Debug("Streaming chat completion started", keyvals...)
// }

// resultChan := make(chan ChatCompletionChunk)
// go func() {
// 	defer close(resultChan)
// 	var totalUsage Usage
// 	firstChunkReceived := false
// 	for chunk := range stream {
// 		if !firstChunkReceived {
// 			ttft := float64(time.Since(startTime).Nanoseconds()) / 1e6
// 			if t.enableOTEL {
// 				span.SetAttributes(attribute.Float64("metrics.ttft_ms", ttft))
// 			}
// 			if keyvals := t.debugLog(ctx, t.ReflectModel); keyvals != nil {
// 				t.logger.Debug("First chunk received for streaming chat completion", append(keyvals, "ttft_ms", ttft)...)
// 			}
// 			firstChunkReceived = true
// 		}
//
// 		totalUsage.PromptTokens += chunk.Usage.PromptTokens
// 		totalUsage.CompletionTokens += chunk.Usage.CompletionTokens
// 		totalUsage.TotalTokens += chunk.Usage.TotalTokens
// 		totalUsage.Cost += chunk.Usage.Cost
// 		totalUsage.CompletionTokenDetails.ReasoningTokens += chunk.Usage.CompletionTokenDetails.ReasoningTokens
// 		totalUsage.PromptTokenDetails.CachedTokens += chunk.Usage.PromptTokenDetails.CachedTokens
// 		totalUsage.CostDetails.UpstreamInferenceCost += chunk.Usage.CostDetails.UpstreamInferenceCost
// 		resultChan <- chunk
// 	}
//
// 	if t.enableOTEL {
// 		t.recordUsageAttributes(span, totalUsage)
// 	}
// 	if keyvals := t.debugLog(ctx, t.ReflectModel); keyvals != nil {
// 		t.logger.Debug("Streaming chat completion completed", append(keyvals, "usage", totalUsage)...)
// 	}
// }()

// return stream, nil
// }
