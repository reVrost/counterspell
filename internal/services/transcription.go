package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	openrouter "github.com/revrost/go-openrouter"
)

const (
	voxtralModel = "mistralai/voxtral-small-24b-2507"
)

// TranscriptionService handles audio transcription via OpenRouter.
type TranscriptionService struct {
	client *openrouter.Client
}

// NewTranscriptionService creates a new transcription service using the server's OpenRouter API key.
func NewTranscriptionService() *TranscriptionService {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		slog.Warn("OPENROUTER_API_KEY not set, transcription will not work")
		return &TranscriptionService{}
	}

	client := openrouter.NewClient(
		apiKey,
		openrouter.WithXTitle("Counterspell"),
		openrouter.WithHTTPReferer("https://counterspell.dev"),
	)

	return &TranscriptionService{client: client}
}

// convertToMp3 converts audio data to mp3 format using ffmpeg.
func convertToMp3(input []byte) ([]byte, error) {
	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "mp3", "-ab", "128k", "-ar", "44100", "pipe:1")
	cmd.Stdin = bytes.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		slog.Error("ffmpeg conversion failed", "stderr", stderr.String())
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	return stdout.Bytes(), nil
}

// TranscribeAudio transcribes audio data to text.
// audioData should be the raw audio bytes (webm/mp3/wav format).
func (s *TranscriptionService) TranscribeAudio(ctx context.Context, audioData io.Reader, format string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("transcription service not configured (missing OPENROUTER_API_KEY)")
	}

	// Read all audio data
	data, err := io.ReadAll(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to read audio data: %w", err)
	}

	// Convert webm to mp3 using ffmpeg (Voxtral only supports mp3/wav)
	if strings.Contains(format, "webm") {
		converted, err := convertToMp3(data)
		if err != nil {
			return "", fmt.Errorf("failed to convert audio: %w", err)
		}
		data = converted
	}

	// Base64 encode the audio
	b64Audio := base64.StdEncoding.EncodeToString(data)

	// Normalize format
	audioFormat := openrouter.AudioFormatMp3
	if strings.Contains(format, "wav") {
		audioFormat = openrouter.AudioFormatWav
	}

	// Create chat completion with audio input
	resp, err := s.client.CreateChatCompletion(ctx, openrouter.ChatCompletionRequest{
		Model: voxtralModel,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role: openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{
					Multi: []openrouter.ChatMessagePart{
						{
							Type: openrouter.ChatMessagePartTypeInputAudio,
							InputAudio: &openrouter.ChatMessageInputAudio{
								Data:   b64Audio,
								Format: audioFormat,
							},
						},
						{
							Type: openrouter.ChatMessagePartTypeText,
							Text: "Transcribe this audio. Output ONLY the transcription, nothing else. If you cannot understand the audio, output an empty string.",
						},
					},
				},
			},
		},
		MaxTokens:   1024,
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("transcription API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no transcription returned")
	}

	transcription := strings.TrimSpace(resp.Choices[0].Message.Content.Text)
	slog.Info("Transcription complete", "length", len(transcription))

	return transcription, nil
}
