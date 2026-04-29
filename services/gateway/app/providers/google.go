package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GoogleProvider implements the Provider interface for Google AI (Gemini)
type GoogleProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(apiKey string) *GoogleProvider {
	return &GoogleProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1",
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}

// Complete sends a completion request to Google AI
func (p *GoogleProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Map model names
	model := p.mapModelName(req.Model)

	// Build Google request (Gemini format)
	contents := p.convertMessagesToGemini(req.Messages)
	if len(contents) == 0 {
		return nil, fmt.Errorf("no messages provided for inference")
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048 // Default to 2K tokens if not specified
	}

	googleReq := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     req.Temperature,
			"maxOutputTokens": maxTokens,
		},
	}

	reqBody, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("google request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var googleResp map[string]interface{}
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract content from Gemini response
	content := ""
	if candidates, ok := googleResp["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if content_parts, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content_parts["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						content, _ = part["text"].(string)
					}
				}
			}
		}
	}

	return &CompletionResponse{
		ID:    fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Model: model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "model",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     0, // Google doesn't always return token counts
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}, nil
}

// convertMessagesToGemini converts standard messages to Gemini format
func (p *GoogleProvider) convertMessagesToGemini(messages []Message) []map[string]interface{} {
	var contents []map[string]interface{}

	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}

		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{"text": msg.Content},
			},
		})
	}

	return contents
}

// mapModelName maps generic model names to Google-specific names
func (p *GoogleProvider) mapModelName(model string) string {
	switch model {
	case "gemini-pro":
		return "gemini-pro"
	case "gemini-1.5":
		return "gemini-1.5-pro"
	case "gemini-2.0-flash":
		return "gemini-2.0-flash"
	case "gemini-1.5-flash", "cheap":
		return "gemini-1.5-flash"
	case "gemini-1.5-pro", "mid":
		return "gemini-1.5-pro"
	case "gemini-ultra", "premium":
		return "gemini-ultra"
	default:
		return "gemini-1.5-flash"
	}
}
