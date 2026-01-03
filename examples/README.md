# 🧪 Orion Library - Simple Example

A minimal example demonstrating Orion's core features.

## Overview

This example shows:
- Creating an agent with in-memory storage
- Running a simple conversation
- Handling streaming responses
- Managing sessions and messages

## Prerequisites

```bash
go get github.com/charmbracelet/orion
go get github.com/charmbracelet/fantasy
go get github.com/charmbracelet/fantasy/providers/openai
```

## Code

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/orion/pkg/orion"
	"github.com/charmbracelet/orion/pkg/orion/events"
	"github.com/charmbracelet/orion/pkg/orion/message"
	"github.com/charmbracelet/orion/pkg/orion/session"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openai"
)

func main() {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Initialize event broker for real-time updates
	eventBroker := events.NewBroker[orion.Message]()

	// Subscribe to message events
	eventBroker.Subscribe(func(event string, msg orion.Message) {
		switch event {
		case "created":
			fmt.Printf("📝 Message created: %s\n", msg.ID)
		case "updated":
			fmt.Printf("✏️  Message updated: %s\n", msg.ID)
		case "deleted":
			fmt.Printf("🗑️  Message deleted: %s\n", msg.ID)
		}
	})

	// Initialize in-memory stores
	sessionStore := session.NewService(nil) // nil for session broker
	messageStore := message.NewService(eventBroker)

	// Initialize OpenAI provider
	provider := openai.New(openai.WithAPIKey(apiKey))
	model := provider.LanguageModel("gpt-4")

	// Create agent
	agent := orion.NewAgent(orion.AgentOptions{
		LargeModel:   model,
		SmallModel:    nil, // Optional small model for auxiliary tasks
		SystemPrompt: "You are a helpful, friendly assistant.",
		Sessions:     sessionStore,
		Messages:     messageStore,
		Tools:        []fantasy.AgentTool{}, // Add tools here
		EventBroker:  eventBroker,
	})

	// Create a conversation session
	sess, err := sessionStore.Create(ctx, "Hello World")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	fmt.Printf("🎉 Created session: %s (Title: %s)\n", sess.ID, sess.Title)

	// Run the agent with a simple prompt
	fmt.Println("\n🤖 Running agent...")
	result, err := agent.Run(ctx, orion.AgentCall{
		SessionID:        sess.ID,
		Prompt:           "Hello! Please tell me a joke about programming.",
		MaxOutputTokens:  256,
	})

	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	// Display the response
	fmt.Println("\n💬 Response:")
	fmt.Println(result.Response.Content.Text())

	// Show token usage
	if result.Usage != nil {
		fmt.Printf("\n📊 Usage:\n")
		fmt.Printf("  Input tokens:  %d\n", result.Usage.InputTokens)
		fmt.Printf("  Output tokens: %d\n", result.Usage.OutputTokens)
		fmt.Printf("  Total tokens:  %d\n", result.Usage.InputTokens+result.Usage.OutputTokens)
	}

	// List messages in the session
	fmt.Println("\n📜 Session Messages:")
	messages, _ := messageStore.List(ctx, sess.ID)
	for i, msg := range messages {
		role := "👤 User"
		if msg.Role == orion.RoleAssistant {
			role = "🤖 Assistant"
		}
		content := msg.Content()
		fmt.Printf("  %d. %s: %s\n", i+1, role, content.Text)
	}

	fmt.Println("\n✨ Example completed!")
}
```

## Run

```bash
export OPENAI_API_KEY=your-key-here
go run main.go
```

## Expected Output

```
🎉 Created session: abc123 (Title: Hello World)
📝 Message created: msg-001
📝 Message created: msg-002
🤖 Running agent...
✏️  Message updated: msg-002
✏️  Message updated: msg-002

💬 Response:
Why do programmers prefer dark mode?

Because light attracts bugs!

📊 Usage:
  Input tokens:  25
  Output tokens:  18
  Total tokens:  43

📜 Session Messages:
  1. 👤 User: Hello! Please tell me a joke about programming.
  2. 🤖 Assistant: Why do programmers prefer dark mode?

Because light attracts bugs!

✨ Example completed!
```

## What's Happening

1. **Setup**: We create event broker and in-memory stores
2. **Events**: We subscribe to message events for real-time updates
3. **Agent**: We initialize the agent with an OpenAI model
4. **Session**: We create a conversation session
5. **Execution**: We run the agent with our prompt
6. **Streaming**: The agent updates messages in real-time
7. **Results**: We display the response and usage statistics
8. **History**: We list all messages in the session

## Extensions

Try adding:
- **Tools**: Implement a calculator tool
- **Queueing**: Send multiple prompts concurrently
- **Cancellation**: Cancel a long-running request
- **Multi-turn**: Continue the conversation
- **Custom Storage**: Implement a PostgreSQL backend
