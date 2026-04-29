package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CodeResponse represents an AI-generated code solution or analysis
type CodeResponse struct {
	Language    string            `json:"language"`
	Code        string            `json:"code"`
	Explanation string            `json:"explanation"`
	Tests       string            `json:"tests,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
}

// CodeAgent handles autonomous software engineering tasks
type CodeAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// NewCodeAgent creates a new code agent
func NewCodeAgent(gatewayURL string) *CodeAgent {
	return &CodeAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// GenerateFeature generates code for a specific feature request
func (ca *CodeAgent) GenerateFeature(ctx context.Context, requirement string, technicalStack string) (*CodeResponse, error) {
	prompt := fmt.Sprintf(`Implement the following feature:

Requirement: %s
Stack: %s

Please provide high-quality, production-ready code, including necessary imports and brief explanations.
Return JSON with keys: language, code, explanation, tests.`,
		requirement, technicalStack)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ca.parseCodeResponse(response)
}

// RefactorCode analyzes and improves existing code
func (ca *CodeAgent) RefactorCode(ctx context.Context, existingCode string, objective string) (*CodeResponse, error) {
	prompt := fmt.Sprintf(`Refactor the following code:

Goal: %s
Source Code:
%s

Focus on readability, performance, and best practices for the specific language.
Return JSON with keys: language, code, explanation.`,
		objective, existingCode)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ca.parseCodeResponse(response)
}

func (ca *CodeAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":       "gemini-2.0-flash",
		"prompt":      prompt,
		"system":      "You are a Senior Software Engineer AI. You specialize in Go, TypeScript, React, and Azure Architecture. Fulfill coding requests with surgical precision and idiomatic patterns.",
		"temperature": 0.1,
	}
	
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", ca.gatewayURL+"/v1/ai", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer fsa_dev_key_12345")
	
	resp, err := ca.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gateway error: %s", string(respBody))
	}
	
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	return result.Response, nil
}

func (ca *CodeAgent) parseCodeResponse(response string) (*CodeResponse, error) {
	response = strings.TrimPrefix(response, "```json\n")
	response = strings.TrimSuffix(response, "\n```")
	response = strings.TrimSpace(response)
	
	var result CodeResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse code response: %w", err)
	}
	
	return &result, nil
}
