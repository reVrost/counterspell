package clearcast

import (
	"context"
	"fmt"
	"log/slog"
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

// Agent struct with slog and OTEL toggle
type Agent struct {
	ID         string
	Model      string
	Prompt     string
	mode       string
	llm        LLMProvider
	logger     *slog.Logger
	tracer     trace.Tracer
	enableOTEL bool
}

// NewAgent creates a new Prompt with slog (JSON handler) and OTEL enabled by default
func NewAgent(id, mode, model, template string,
	llm LLMProvider, opts ...AgentOption) *Agent {
	p := &Agent{
		ID:         id,
		Model:      model,
		mode:       mode,
		Prompt:     template,
		llm:        llm,
		logger:     slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		tracer:     otel.Tracer("clearcast/prompt"),
		enableOTEL: true,
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

func (t *Agent) render(ctx context.Context, values map[string]any) (string, error) {
	var span trace.Span
	if t.enableOTEL {
		ctx, span = t.tracer.Start(ctx, "Prompt.Render")
		defer span.End()
		span.SetAttributes(
			attribute.String("template.content", t.Prompt),
			attribute.Int("values.count", len(values)),
		)
	}

	parsedTmpl, err := template.New("template").
		Option("missingkey=error").
		Parse(t.Prompt)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Failed to parse template", "template", t.Prompt, "error", err)
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

type StepOption func(*ChatCompletionRequest)

func WithResponseFormat(format ResponseFormat) StepOption {
	return func(req *ChatCompletionRequest) {
		req.ResponseFormat = &format
	}
}

// Non-streaming completion
func (t *Agent) Step(ctx context.Context, args map[string]any, opts ...StepOption) (ChatCompletionResponse, error) {
	var span trace.Span
	if t.enableOTEL {
		ctx, span = t.tracer.Start(ctx, "Prompt.ChatCompletion")
		defer span.End()
		span.SetAttributes(attribute.String("model", t.Model))
	}

	rendered, err := t.render(ctx, args)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Error rendering template for chat completion", "model", t.Model, "error", err)
		return ChatCompletionResponse{}, fmt.Errorf("error rendering template: %w", err)
	}

	// TODO: add provide for JSON format (structured response)
	req := ChatCompletionRequest{
		Model:    t.Model,
		Messages: []ChatMessage{SystemMessage(rendered)},
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
		t.logger.Error("Chat completion failed", "model", t.Model, "error", err)
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

	return resp, nil
}

// Streaming completion
func (t *Agent) StepStream(ctx context.Context, model string, args map[string]any) (<-chan ChatCompletionChunk, error) {
	var span trace.Span
	if t.enableOTEL {
		ctx, span = t.tracer.Start(ctx, "Prompt.ChatCompletionStream")
		defer span.End()
		span.SetAttributes(attribute.String("model", model))
	}

	rendered, err := t.render(ctx, args)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Error rendering template for streaming chat completion", "model", model, "error", err)
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	req := ChatCompletionRequest{
		Model:    model,
		Messages: []ChatMessage{SystemMessage(rendered)},
	}

	startTime := time.Now()
	stream, err := t.llm.ChatCompletionStream(ctx, req)
	if err != nil {
		if t.enableOTEL {
			span.RecordError(err)
		}
		t.logger.Error("Streaming chat completion failed", "model", model, "error", err)
		return nil, err
	}

	if keyvals := t.debugLog(ctx, model); keyvals != nil {
		t.logger.Debug("Streaming chat completion started", keyvals...)
	}

	usageChan := make(chan ChatCompletionChunk)
	go func() {
		defer close(usageChan)
		var totalUsage Usage
		firstChunkReceived := false
		for chunk := range stream {
			if !firstChunkReceived {
				ttft := float64(time.Since(startTime).Nanoseconds()) / 1e6
				if t.enableOTEL {
					span.SetAttributes(attribute.Float64("metrics.ttft_ms", ttft))
				}
				if keyvals := t.debugLog(ctx, model); keyvals != nil {
					t.logger.Debug("First chunk received for streaming chat completion", append(keyvals, "ttft_ms", ttft)...)
				}
				firstChunkReceived = true
			}

			totalUsage.PromptTokens += chunk.Usage.PromptTokens
			totalUsage.CompletionTokens += chunk.Usage.CompletionTokens
			totalUsage.TotalTokens += chunk.Usage.TotalTokens
			totalUsage.Cost += chunk.Usage.Cost
			totalUsage.CompletionTokenDetails.ReasoningTokens += chunk.Usage.CompletionTokenDetails.ReasoningTokens
			totalUsage.PromptTokenDetails.CachedTokens += chunk.Usage.PromptTokenDetails.CachedTokens
			totalUsage.CostDetails.UpstreamInferenceCost += chunk.Usage.CostDetails.UpstreamInferenceCost
			usageChan <- chunk
		}

		if t.enableOTEL {
			t.recordUsageAttributes(span, totalUsage)
		}
		if keyvals := t.debugLog(ctx, model); keyvals != nil {
			t.logger.Debug("Streaming chat completion completed", append(keyvals, "usage", totalUsage)...)
		}
	}()

	return usageChan, nil
}
