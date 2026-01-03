package tools

import (
	"context"
	"fmt"

	"charm.land/fantasy"
)

// CalculatorParams defines parameters for the calculator tool.
type CalculatorParams struct {
	Expression string `json:"expression" description:"Mathematical expression to evaluate (e.g., '2+2', '10*5')"`
}

// NewCalculatorTool creates a tool that can evaluate mathematical expressions.
func NewCalculatorTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"calculator",
		"Evaluate mathematical expressions",
		func(ctx context.Context, params CalculatorParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Evaluate expression
			result, err := evaluateExpression(params.Expression)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Error: %v", err)), nil
			}

			return fantasy.NewTextResponse(fmt.Sprintf("Result: %s = %s", params.Expression, result)), nil
		},
	)
}

// evaluateExpression is a simple expression evaluator.
// For production, use a proper math expression parser.
func evaluateExpression(expr string) (string, error) {
	// Very simple evaluator - only handles basic operations
	// In real implementation, use a proper expression parser like
	// github.com/Knetic/govaluate or github.com/alecthomas/participle

	// For demo, just return the expression with a note
	return fmt.Sprintf("%d (demo - implement real evaluator)", len(expr)), nil
}

// WeatherParams defines parameters for the weather tool.
type WeatherParams struct {
	City string `json:"city" description:"City name to get weather for"`
	Unit string `json:"unit" description:"Temperature unit: 'celsius' or 'fahrenheit' (default: celsius)"`
}

// NewWeatherTool creates a tool that can fetch weather information.
func NewWeatherTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"weather",
		"Get current weather information for a city",
		func(ctx context.Context, params WeatherParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.City == "" {
				return fantasy.NewTextErrorResponse("Error: city parameter is required"), nil
			}

			// In real implementation, call a weather API
			// For demo, return mock data
			weatherData := fmt.Sprintf("Weather in %s: 72°F, Sunny\nHumidity: 45%%\nWind: 8 mph", params.City)

			return fantasy.NewTextResponse(weatherData), nil
		},
	)
}

// TimeParams defines parameters for the time tool.
type TimeParams struct {
	Timezone string `json:"timezone" description:"IANA timezone name (e.g., 'America/New_York', 'UTC'). If empty, returns UTC."`
}

// NewTimeTool creates a tool that can get current time in a timezone.
func NewTimeTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"time",
		"Get current time in a specific timezone",
		func(ctx context.Context, params TimeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// In real implementation, use time.LoadLocation
			// For demo, return a formatted string

			if params.Timezone == "" {
				params.Timezone = "UTC"
			}

			return fantasy.NewTextResponse(fmt.Sprintf("Current time in %s: [demo - implement real time lookup]", params.Timezone)), nil
		},
	)
}

// WebSearchParams defines parameters for the web search tool.
type WebSearchParams struct {
	Query  string `json:"query" description:"Search query"`
	MaxResults int  `json:"max_results" description:"Maximum number of results to return (default: 5)"`
}

// NewWebSearchTool creates a tool that can search the web.
func NewWebSearchTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"web_search",
		"Search the web for information",
		func(ctx context.Context, params WebSearchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" {
				return fantasy.NewTextErrorResponse("Error: query parameter is required"), nil
			}

			if params.MaxResults <= 0 || params.MaxResults > 10 {
				params.MaxResults = 5
			}

			// In real implementation, call a search API
			// For demo, return mock results
			results := fmt.Sprintf("Web search results for '%s' (showing %d of %d):\n\n",
				params.Query, params.MaxResults, params.MaxResults)

			for i := 1; i <= params.MaxResults; i++ {
				results += fmt.Sprintf("%d. [Result %d] - Title placeholder\n   https://example.com/result%d\n\n", i, i, i)
			}

			return fantasy.NewTextResponse(results), nil
		},
	)
}

// EchoParams defines parameters for the echo tool.
type EchoParams struct {
	Message string `json:"message" description:"Message to echo back"`
	Times   int    `json:"times" description:"Number of times to repeat the message (default: 1)"`
}

// NewEchoTool creates a tool that repeats a message.
func NewEchoTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"echo",
		"Repeat a message multiple times",
		func(ctx context.Context, params EchoParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Message == "" {
				return fantasy.NewTextErrorResponse("Error: message parameter is required"), nil
			}

			if params.Times <= 0 || params.Times > 10 {
				params.Times = 1
			}

			var result string
			for i := 0; i < params.Times; i++ {
				result += fmt.Sprintf("%s [%d]\n", params.Message, i+1)
			}

			return fantasy.NewTextResponse(result), nil
		},
	)
}

// StringLengthParams defines parameters for the string length tool.
type StringLengthParams struct {
	Text string `json:"text" description:"Text to analyze"`
}

// NewStringLengthTool creates a tool that analyzes text length.
func NewStringLengthTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"string_length",
		"Analyze the length of a text string",
		func(ctx context.Context, params StringLengthParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			charCount := len(params.Text)
			wordCount := len([]rune(params.Text)) // Count Unicode characters
			lineCount := 1
			for _, c := range params.Text {
				if c == '\n' {
					lineCount++
				}
			}

			result := fmt.Sprintf("Text analysis:\n"+
				"  Characters: %d\n"+
				"  Words (Unicode): %d\n"+
				"  Lines: %d\n"+
				"  Bytes: %d",
				charCount, wordCount, lineCount, len([]byte(params.Text)))

			return fantasy.NewTextResponse(result), nil
		},
	)
}

// RandomNumberParams defines parameters for the random number tool.
type RandomNumberParams struct {
	Min int `json:"min" description:"Minimum value (inclusive, default: 1)"`
	Max int `json:"max" description:"Maximum value (inclusive, default: 100)"`
}

// NewRandomNumberTool creates a tool that generates random numbers.
func NewRandomNumberTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"random_number",
		"Generate a random number in a range",
		func(ctx context.Context, params RandomNumberParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Min > params.Max {
				return fantasy.NewTextErrorResponse("Error: min cannot be greater than max"), nil
			}

			// Set defaults
			if params.Min == 0 {
				params.Min = 1
			}
			if params.Max == 0 {
				params.Max = 100
			}

			// In real implementation, use crypto/rand or math/rand
			// For demo, return a fixed value
			result := params.Min + (params.Max-params.Min)/2

			return fantasy.NewTextResponse(fmt.Sprintf("Random number between %d and %d: %d", params.Min, params.Max, result)), nil
		},
	)
}

// Base64EncodeParams defines parameters for base64 encoding.
type Base64EncodeParams struct {
	Text string `json:"text" description:"Text to encode"`
}

// NewBase64EncodeTool creates a tool that encodes text to base64.
func NewBase64EncodeTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"base64_encode",
		"Encode text to base64",
		func(ctx context.Context, params Base64EncodeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Text == "" {
				return fantasy.NewTextErrorResponse("Error: text parameter is required"), nil
			}

			// In real implementation, use encoding/base64
			encoded := "[base64 encoded text - implement real encoding]"

			return fantasy.NewTextResponse(fmt.Sprintf("Base64 encoded: %s", encoded)), nil
		},
	)
}

// Base64DecodeParams defines parameters for base64 decoding.
type Base64DecodeParams struct {
	Encoded string `json:"encoded" description:"Base64 encoded string to decode"`
}

// NewBase64DecodeTool creates a tool that decodes base64 to text.
func NewBase64DecodeTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"base64_decode",
		"Decode base64 encoded string",
		func(ctx context.Context, params Base64DecodeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Encoded == "" {
				return fantasy.NewTextErrorResponse("Error: encoded parameter is required"), nil
			}

			// In real implementation, use encoding/base64
			decoded := "[decoded text - implement real decoding]"

			return fantasy.NewTextResponse(fmt.Sprintf("Decoded: %s", decoded)), nil
		},
	)
}

// UUIDParams defines parameters for UUID generation.
type UUIDParams struct {
	Count int `json:"count" description:"Number of UUIDs to generate (default: 1, max: 10)"`
}

// NewUUIDTool creates a tool that generates UUIDs.
func NewUUIDTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"uuid",
		"Generate random UUIDs",
		func(ctx context.Context, params UUIDParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Count <= 0 || params.Count > 10 {
				params.Count = 1
			}

			var results string
			for i := 0; i < params.Count; i++ {
				results += fmt.Sprintf("%d. [UUID %d - implement real UUID generation]\n", i+1, i+1)
			}

			return fantasy.NewTextResponse(results), nil
		},
	)
}

// AllExampleTools returns all example tools.
func AllExampleTools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		NewCalculatorTool(),
		NewWeatherTool(),
		NewTimeTool(),
		NewWebSearchTool(),
		NewEchoTool(),
		NewStringLengthTool(),
		NewRandomNumberTool(),
		NewBase64EncodeTool(),
		NewBase64DecodeTool(),
		NewUUIDTool(),
	}
}
