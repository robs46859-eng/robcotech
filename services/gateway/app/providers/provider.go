// Package providers implements LLM provider integrations
package providers

import (
	"context"
	"fmt"
)

// Provider interface for LLM providers
type Provider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}

// CompletionRequest is a standardized request format
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionResponse is a standardized response format
type CompletionResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderType identifies the provider
type ProviderType string

const (
	ProviderAnthropic  ProviderType = "anthropic"
	ProviderOpenAI     ProviderType = "openai"
	ProviderGoogle     ProviderType = "google"
	ProviderLocal      ProviderType = "local"
)

// NewProvider creates a provider instance by type
func NewProvider(providerType ProviderType, apiKey string) (Provider, error) {
	switch providerType {
	case ProviderAnthropic:
		return NewAnthropicProvider(apiKey), nil
	case ProviderOpenAI:
		return NewOpenAIProvider(apiKey), nil
	case ProviderGoogle:
		return NewGoogleProvider(apiKey), nil
	case ProviderLocal:
		return NewLocalProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetProviderForModel returns the appropriate provider for a model name
func GetProviderForModel(modelName string, apiKeys map[string]string) (Provider, error) {
	modelLower := modelName

	// Anthropic models
	if containsAny(modelLower, []string{"claude", "haiku", "sonnet", "opus"}) {
		return NewProvider(ProviderAnthropic, apiKeys["anthropic"])
	}

	// OpenAI models
	if containsAny(modelLower, []string{"gpt-3.5", "gpt-4", "o1"}) {
		return NewProvider(ProviderOpenAI, apiKeys["openai"])
	}

	// Google models
	if containsAny(modelLower, []string{"gemini", "palm"}) {
		return NewProvider(ProviderGoogle, apiKeys["google"])
	}

	// Local models
	if containsAny(modelLower, []string{"local", "phi", "llama", "mistral"}) {
		return NewProvider(ProviderLocal, "")
	}

	// Default to Anthropic
	return NewProvider(ProviderAnthropic, apiKeys["anthropic"])
}

func containsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
