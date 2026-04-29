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

// DebugPlan represents an AI-generated debugging plan
type DebugPlan struct {
	RootCause     string            `json:"root_cause"`
	SuggestedFix  string            `json:"suggested_fix"`
	CodeChanges   map[string]string `json:"code_changes,omitempty"`
	CommandsToRun []string          `json:"commands_to_run,omitempty"`
}

// DebugAgent handles autonomous debugging and code fixing
type DebugAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// NewDebugAgent creates a new debug agent
func NewDebugAgent(gatewayURL string) *DebugAgent {
	return &DebugAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// DebugError analyzes an error and codebase context to provide fixes
func (da *DebugAgent) DebugError(ctx context.Context, errorMessage string, codeContext string) (*DebugPlan, error) {
	prompt := fmt.Sprintf(`Debug the following error and provide a fix:

Error Message: %s
Code Context: %s

Return JSON with keys: root_cause, suggested_fix, code_changes (map of file to code), commands_to_run.`,
		errorMessage, codeContext)

	response, err := da.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return da.parseDebugPlan(response)
}

func (da *DebugAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":       "gemini-1.5-flash", // Using a reliable, fast model
		"prompt":      prompt,
		"system":      "You are an expert Debug Agent. You specialize in diagnosing software errors, fixing code, and managing deployments. Always return valid JSON matching the requested structure.",
		"temperature": 0.2, // Low temperature for precise code fixes
	}
	
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", da.gatewayURL+"/v1/ai", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer fsa_dev_key_12345")
	
	resp, err := da.httpClient.Do(req)
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

func (da *DebugAgent) parseDebugPlan(response string) (*DebugPlan, error) {
	// Clean markdown formatting if present
	response = strings.TrimPrefix(response, "```json\n")
	response = strings.TrimSuffix(response, "\n```")
	response = strings.TrimSpace(response)
	
	var result DebugPlan
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse debug plan: %w", err)
	}
	
	return &result, nil
}
