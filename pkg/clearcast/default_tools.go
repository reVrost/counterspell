package clearcast

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// SerperDevTool creates a new tool for interacting with the Serper.dev API.
func SerperDevTool() *Tool {
	return &Tool{
		ID: "serper_dev",
		Usage: `You can use serper_dev. To use serper_dev, respond with JSON in this format:
      {
        "tool": "serper_dev",
        "params": {"query": "what ever you would like to ask"}
      }
		`,
		Description: "Searches the web using Serper.dev",
		Execute: func(ctx context.Context, params map[string]any) (any, error) {
			query, ok := params["query"].(string)
			if !ok {
				return nil, fmt.Errorf("query parameter is required and must be a string")
			}

			apiKey := os.Getenv("SERPER_API_KEY")
			if apiKey == "" {
				return nil, fmt.Errorf("SERPER_API_KEY environment variable not set")
			}

			// Prepare the request body
			requestBody, err := json.Marshal(map[string]string{"q": query})
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}

			// Create the HTTP request
			req, err := http.NewRequestWithContext(ctx, "POST", "https://google.serper.dev/search", bytes.NewBuffer(requestBody))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			// Set headers
			req.Header.Set("X-API-KEY", apiKey)
			req.Header.Set("Content-Type", "application/json")

			// Execute the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to execute request: %w", err)
			}
			defer resp.Body.Close()

			// Read the response body
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
			}

			// Unmarshal the response into a generic map
			var result any
			if err := json.Unmarshal(responseBody, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return result, nil
		},
	}
}
