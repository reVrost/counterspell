package clearcast

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DebugContextKey for context-based debug flag
type DebugContextKey struct{}

// Prompt struct with zerolog and OTEL
type Prompt struct {
	llm    LLMProvider
	Prompt string
	logger zerolog.Logger
	tracer trace.Tracer
}

// NewPrompt creates a new Prompt with zerolog and OTEL
func NewPrompt(template string, llm LLMProvider, logger zerolog.Logger) *Prompt {
	return &Prompt{
		Prompt: template,
		llm:    llm,
		logger: logger,
		tracer: otel.Tracer("clearcast/prompt"),
	}
}

// recordUsageAttributes adds usage-related attributes to an OTEL span
func (t *Prompt) recordUsageAttributes(span trace.Span, usage Usage) {
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

// debugLog creates a zerolog event with common fields
func (t *Prompt) debugLog(ctx context.Context, model string) *zerolog.Event {
	if isDebugEnabled, _ := ctx.Value(DebugContextKey{}).(bool); isDebugEnabled {
		return t.logger.Debug().Str("model", model)
	}
	return nil
}

func (t *Prompt) render(ctx context.Context, values map[string]any) (string, error) {
	ctx, span := t.tracer.Start(ctx, "Prompt.Render")
	defer span.End()

	parsedTmpl, err := template.New("template").
		Option("missingkey=error").
		Parse(t.Prompt)
	if err != nil {
		span.RecordError(err)
		t.logger.Error().
			Err(err).
			Str("template", t.Prompt).
			Msg("Failed to parse template")
		return "", err
	}

	span.SetAttributes(
		attribute.String("template.content", t.Prompt),
		attribute.Int("values.count", len(values)),
	)

	sb := new(strings.Builder)
	if err := parsedTmpl.Execute(sb, values); err != nil {
		span.RecordError(err)
		t.logger.Error().
			Err(err).
			Interface("values", values).
			Msg("Failed to execute template")
		return "", err
	}

	rendered := sb.String()
	if log := t.debugLog(ctx, ""); log != nil {
		log.Str("rendered", rendered).
			Interface("values", values).
			Msg("Template rendered successfully")
	}

	return rendered, nil
}

// Non-streaming completion
func (t *Prompt) ChatCompletion(ctx context.Context, model string, args map[string]any) (ChatCompletionResponse, error) {
	ctx, span := t.tracer.Start(ctx, "Prompt.ChatCompletion")
	defer span.End()

	span.SetAttributes(attribute.String("model", model))

	rendered, err := t.render(ctx, args)
	if err != nil {
		span.RecordError(err)
		t.logger.Error().
			Err(err).
			Str("model", model).
			Msg("Error rendering template for chat completion")
		return ChatCompletionResponse{}, fmt.Errorf("error rendering template: %w", err)
	}

	req := ChatCompletionRequest{
		Model:    model,
		Messages: []ChatMessage{SystemMessage(rendered)},
	}

	startTime := time.Now()
	resp, err := t.llm.ChatCompletion(ctx, req)
	if err != nil {
		span.RecordError(err)
		t.logger.Error().
			Err(err).
			Str("model", model).
			Msg("Chat completion failed")
		return ChatCompletionResponse{}, err
	}

	ttft := float64(time.Since(startTime).Nanoseconds()) / 1e6
	t.recordUsageAttributes(span, resp.Usage)
	span.SetAttributes(attribute.Float64("metrics.ttft_ms", ttft))

	if log := t.debugLog(ctx, model); log != nil {
		log.Interface("response", resp).
			Float64("ttft_ms", ttft).
			Interface("usage", resp.Usage).
			Msg("Chat completion successful")
	}

	return resp, nil
}

// Streaming completion
func (t *Prompt) ChatCompletionStream(ctx context.Context, model string, args map[string]any) (<-chan ChatCompletionChunk, error) {
	ctx, span := t.tracer.Start(ctx, "Prompt.ChatCompletionStream")
	defer span.End()

	span.SetAttributes(attribute.String("model", model))

	rendered, err := t.render(ctx, args)
	if err != nil {
		span.RecordError(err)
		t.logger.Error().
			Err(err).
			Str("model", model).
			Msg("Error rendering template for streaming chat completion")
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	req := ChatCompletionRequest{
		Model:    model,
		Messages: []ChatMessage{SystemMessage(rendered)},
	}

	startTime := time.Now()
	stream, err := t.llm.ChatCompletionStream(ctx, req)
	if err != nil {
		span.RecordError(err)
		t.logger.Error().
			Err(err).
			Str("model", model).
			Msg("Streaming chat completion failed")
		return nil, err
	}

	if log := t.debugLog(ctx, model); log != nil {
		log.Msg("Streaming chat completion started")
	}

	usageChan := make(chan ChatCompletionChunk)
	go func() {
		defer close(usageChan)
		var totalUsage Usage
		firstChunkReceived := false
		for chunk := range stream {
			if !firstChunkReceived {
				ttft := float64(time.Since(startTime).Nanoseconds()) / 1e6
				span.SetAttributes(attribute.Float64("metrics.ttft_ms", ttft))
				if log := t.debugLog(ctx, model); log != nil {
					log.Float64("ttft_ms", ttft).
						Msg("First chunk received for streaming chat completion")
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

		t.recordUsageAttributes(span, totalUsage)
		if log := t.debugLog(ctx, model); log != nil {
			log.Interface("usage", totalUsage).
				Msg("Streaming chat completion completed")
		}
	}()

	return usageChan, nil
}
