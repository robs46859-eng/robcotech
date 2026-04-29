package providers

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// LocalProvider implements the Provider interface for local model inference
// In production, this would connect to a local model server (ollama, vllm, etc.)
type LocalProvider struct {
	modelPath string
}

// NewLocalProvider creates a new local provider
func NewLocalProvider() *LocalProvider {
	return &LocalProvider{
		modelPath: "",
	}
}

// Name returns the provider name
func (p *LocalProvider) Name() string {
	return "local"
}

// Complete sends a completion request to a local model
// For now, returns mock responses - in production would call local inference server
func (p *LocalProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// In production, this would:
	// 1. Call ollama/vllm/llama.cpp local server
	// 2. Or run model directly with transformers.go
	// For now, return a simple mock response based on request content

	content := p.generateLocalResponse(req.Messages)

	return &CompletionResponse{
		ID:    fmt.Sprintf("local-%d", time.Now().Unix()),
		Model: req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     p.countTokens(req.Messages),
			CompletionTokens: p.countWords(content) / 4, // Rough estimate
			TotalTokens:      0,
		},
	}, nil
}

// generateLocalResponse generates a simple response for local models
// This is a placeholder - in production would use actual model inference
func (p *LocalProvider) generateLocalResponse(messages []Message) string {
	if len(messages) == 0 {
		return "I'm a local model. Please configure a local inference server."
	}

	lastMessage := messages[len(messages)-1].Content

	// Simple keyword-based responses for common tasks
	if p.containsAny(lastMessage, []string{"classify", "categorize", "tag"}) {
		return "Classification: This appears to be a categorization task. The content can be classified based on the provided criteria."
	}

	if p.containsAny(lastMessage, []string{"summarize", "summary"}) {
		return "Summary: The provided content has been processed. Key points have been extracted and condensed."
	}

	if p.containsAny(lastMessage, []string{"extract", "find"}) {
		return "Extraction: Relevant information has been identified and extracted from the input."
	}

	if p.containsAny(lastMessage, []string{"validate", "check", "schema"}) {
		return "Validation: The input has been checked against the specified schema. Format appears valid."
	}

	if p.containsAny(lastMessage, []string{"translate"}) {
		return "Translation: [Local translation would appear here. Configure a local translation model for actual translation.]"
	}

	// Default response
	return "I'm a local model processing your request. For full capabilities, configure a local inference server like Ollama or vLLM."
}

func (p *LocalProvider) containsAny(s string, substrs []string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func (p *LocalProvider) countTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += p.countWords(msg.Content)
	}
	// Rough estimate: 1 token ≈ 4 characters or 0.75 words
	return total * 4 / 3
}

func (p *LocalProvider) countWords(s string) int {
	return len(strings.Fields(s))
}
