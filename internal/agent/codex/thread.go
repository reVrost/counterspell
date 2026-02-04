package codex

import (
	"context"
	"errors"
	"fmt"
)

type Turn struct {
	Items         []ThreadItem
	FinalResponse string
	Usage         *Usage
}

type StreamedTurn struct {
	Events <-chan ThreadEvent
	Done   <-chan error
}

type UserInput struct {
	Type string
	Text string
	Path string
}

type Input any

type Thread struct {
	exec          *Exec
	options       Options
	id            string
	threadOptions ThreadOptions
}

func NewThread(exec *Exec, options Options, threadOptions ThreadOptions, id string) *Thread {
	return &Thread{
		exec:          exec,
		options:       options,
		id:            id,
		threadOptions: threadOptions,
	}
}

func (t *Thread) ID() string {
	return t.id
}

func (t *Thread) RunStreamed(ctx context.Context, input Input, turnOptions TurnOptions) *StreamedTurn {
	events := make(chan ThreadEvent, 32)
	done := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(done)

		schemaFile, err := CreateOutputSchemaFile(turnOptions.OutputSchema)
		if err != nil {
			done <- err
			return
		}
		defer func() {
			_ = schemaFile.Cleanup()
		}()

		prompt, images, err := normalizeInput(input)
		if err != nil {
			done <- err
			return
		}

		runCtx := ctx
		var cancel context.CancelFunc
		if turnOptions.Signal != nil {
			runCtx, cancel = context.WithCancel(ctx)
			go func() {
				select {
				case <-turnOptions.Signal:
					cancel()
				case <-runCtx.Done():
				}
			}()
		}

		stream := t.exec.Run(runCtx, ExecArgs{
			Input:                 prompt,
			BaseURL:               t.options.BaseURL,
			APIKey:                t.options.APIKey,
			ThreadID:              t.id,
			Images:                images,
			Model:                 t.threadOptions.Model,
			SandboxMode:           t.threadOptions.SandboxMode,
			WorkingDirectory:      t.threadOptions.WorkingDirectory,
			AdditionalDirectories: t.threadOptions.AdditionalDirectories,
			SkipGitRepoCheck:      t.threadOptions.SkipGitRepoCheck,
			OutputSchemaFile:      schemaFile.SchemaPath,
			ModelReasoningEffort:  t.threadOptions.ModelReasoningEffort,
			NetworkAccessEnabled:  t.threadOptions.NetworkAccessEnabled,
			WebSearchMode:         t.threadOptions.WebSearchMode,
			WebSearchEnabled:      t.threadOptions.WebSearchEnabled,
			ApprovalPolicy:        t.threadOptions.ApprovalPolicy,
		})

		for line := range stream.Lines {
			if line == "" {
				continue
			}
			event, err := ParseThreadEvent(line)
			if err != nil {
				done <- fmt.Errorf("failed to parse item: %w", err)
				return
			}
			if event.Type == "thread.started" && event.ThreadID != "" {
				t.id = event.ThreadID
			}
			events <- event
		}
		if cancel != nil {
			cancel()
		}
		done <- <-stream.Done
	}()

	return &StreamedTurn{Events: events, Done: done}
}

func (t *Thread) Run(ctx context.Context, input Input, turnOptions TurnOptions) (*Turn, error) {
	stream := t.RunStreamed(ctx, input, turnOptions)
	items := []ThreadItem{}
	var finalResponse string
	var usage *Usage
	var turnErr error

	for event := range stream.Events {
		switch event.Type {
		case "item.completed":
			if event.Item != nil {
				if msg, ok := event.Item.(AgentMessageItem); ok {
					finalResponse = msg.Text
				}
				items = append(items, event.Item)
			}
		case "turn.completed":
			usage = event.Usage
		case "turn.failed":
			if event.Error != nil {
				turnErr = errors.New(event.Error.Message)
			} else {
				turnErr = errors.New("turn failed")
			}
		case "error":
			if event.Message != "" {
				turnErr = errors.New(event.Message)
			} else {
				turnErr = errors.New("turn failed")
			}
		}
	}

	if err := <-stream.Done; err != nil {
		return nil, err
	}
	if turnErr != nil {
		return nil, turnErr
	}
	return &Turn{
		Items:         items,
		FinalResponse: finalResponse,
		Usage:         usage,
	}, nil
}

func normalizeInput(input Input) (string, []string, error) {
	switch v := input.(type) {
	case string:
		return v, nil, nil
	case []UserInput:
		promptParts := []string{}
		images := []string{}
		for _, item := range v {
			switch item.Type {
			case "text":
				if item.Text != "" {
					promptParts = append(promptParts, item.Text)
				}
			case "local_image":
				if item.Path != "" {
					images = append(images, item.Path)
				}
			default:
				return "", nil, fmt.Errorf("unsupported input type %q", item.Type)
			}
		}
		return joinPrompt(promptParts), images, nil
	default:
		return "", nil, fmt.Errorf("unsupported input: %T", input)
	}
}

func joinPrompt(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	prompt := parts[0]
	for i := 1; i < len(parts); i++ {
		prompt += "\n\n" + parts[i]
	}
	return prompt
}
