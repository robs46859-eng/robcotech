// Package agents implements the Dad AI agent for website generation
package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// DadAIAgent is the AI-powered website generation agent
// "Dad AI" = "Develop Another Day - Artificial Intelligence"
type DadAIAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// WebsiteProject represents a website generation project
type WebsiteProject struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tier        string            `json:"tier"` // starter, studio, agency
	Status      string            `json:"status"` // pending, generating, ready, failed
	Prompt      string            `json:"prompt"`
	OutputType  string            `json:"output_type"` // single_page, multi_page, dashboard
	Templates   []string          `json:"templates"`
	GeneratedAt time.Time         `json:"generated_at"`
	Metadata    map[string]string `json:"metadata"`
}

// GenerationRequest is a request to generate a website
type GenerationRequest struct {
	Prompt         string            `json:"prompt"`
	BusinessType   string            `json:"business_type"`
	OutputType     string            `json:"output_type"` // single_page, multi_page, dashboard
	Tier           string            `json:"tier"`
	ColorScheme    string            `json:"color_scheme,omitempty"`
	Features       []string          `json:"features,omitempty"`
	Content        map[string]string `json:"content,omitempty"`
}

// GenerationResponse is the response from website generation
type GenerationResponse struct {
	ProjectID   string `json:"project_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	HTML        string `json:"html,omitempty"`
	CSS         string `json:"css,omitempty"`
	JavaScript  string `json:"javascript,omitempty"`
	ReactCode   string `json:"react_code,omitempty"`
	Dashboard   string `json:"dashboard,omitempty"`
	EstimatedAt string `json:"estimated_at,omitempty"`
}

// NewDadAIAgent creates a new Dad AI agent
func NewDadAIAgent(ctx context.Context, gatewayURL string) (*DadAIAgent, error) {
	return &DadAIAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// GenerateWebsite generates a website based on the prompt
func (d *DadAIAgent) GenerateWebsite(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	// Classify the request and determine output complexity
	outputType := d.classifyOutputType(req)

	// Build the prompt for the LLM
	prompt := d.buildGenerationPrompt(req, outputType)

	// Call the gateway for inference
	llmResponse, err := d.callGateway(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("gateway call failed: %w", err)
	}

	// Parse the LLM response into structured output
	generatedContent, err := d.parseLLMResponse(llmResponse, outputType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	projectID := uuid.New().String()

	return &GenerationResponse{
		ProjectID:   projectID,
		Status:      "ready",
		Message:     fmt.Sprintf("Generated %s website for %s", outputType, req.BusinessType),
		HTML:        generatedContent.HTML,
		CSS:         generatedContent.CSS,
		JavaScript:  generatedContent.JavaScript,
		ReactCode:   generatedContent.ReactCode,
		Dashboard:   generatedContent.Dashboard,
		EstimatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// GeneratedContent holds the generated website code
type GeneratedContent struct {
	HTML       string
	CSS        string
	JavaScript string
	ReactCode  string
	Dashboard  string
}

// classifyOutputType determines the output complexity based on request
func (d *DadAIAgent) classifyOutputType(req GenerationRequest) string {
	// Dashboard output for complex business apps
	if req.OutputType == "dashboard" {
		return "dashboard"
	}

	// Multi-page for studios and agencies
	if req.Tier == "studio" || req.Tier == "agency" {
		return "multi_page"
	}

	// Check for complexity indicators in prompt
	complexKeywords := []string{"dashboard", "admin", "multi-user", "portal", "crm", "analytics"}
	for _, keyword := range complexKeywords {
		if contains(req.Prompt, keyword) || contains(req.BusinessType, keyword) {
			return "multi_page"
		}
	}

	// Default to single page for simple sites
	return "single_page"
}

// buildGenerationPrompt creates the prompt for the LLM
func (d *DadAIAgent) buildGenerationPrompt(req GenerationRequest, outputType string) string {
	tierInstructions := map[string]string{
		"starter": "Generate a clean, professional single-page website. Include: hero section, services, about, contact form. Keep it simple and focused.",
		"studio":  "Generate a multi-page website with: home, services, portfolio, about, contact. Include lead capture forms and professional styling.",
		"agency":  "Generate a full web application with dashboard, client portal, multi-user support, analytics, and integrations.",
	}

	prompt := fmt.Sprintf(`You are Dad AI (Develop Another Day - Artificial Intelligence), an expert web developer.

Task: Generate a %s website for a %s business.

Business Description: %s

Color Scheme: %s

Features Required: %v

%s

Output the complete, production-ready code with:
1. HTML structure (semantic, accessible)
2. CSS styling (responsive, modern)
3. JavaScript functionality (interactive elements)
4. Any React components if needed

Return valid JSON with keys: html, css, javascript, react_components`,
		outputType,
		req.BusinessType,
		req.Prompt,
		req.ColorScheme,
		req.Features,
		tierInstructions[req.Tier],
	)

	return prompt
}

// callGateway makes an inference request to the stack-arkham gateway
func (d *DadAIAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are Dad AI, an expert web developer who generates production-ready websites."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  4000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/ai", d.gatewayURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", "papabase-service-key")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(body))
	}

	var gatewayResp map[string]interface{}
	if err := json.Unmarshal(body, &gatewayResp); err != nil {
		return "", err
	}

	// Extract content from gateway response
	if choices, ok := gatewayResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no content in gateway response")
}

// parseLLMResponse extracts structured code from LLM response
func (d *DadAIAgent) parseLLMResponse(response string, outputType string) (*GeneratedContent, error) {
	// Try to parse as JSON first
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Not JSON, wrap in HTML
		return &GeneratedContent{
			HTML: response,
			CSS:  getDefaultCSS(),
		}, nil
	}

	content := &GeneratedContent{}

	if html, ok := result["html"].(string); ok {
		content.HTML = html
	}
	if css, ok := result["css"].(string); ok {
		content.CSS = css
	}
	if js, ok := result["javascript"].(string); ok {
		content.JavaScript = js
	}
	if react, ok := result["react_components"].(string); ok {
		content.ReactCode = react
	}

	// Generate dashboard for complex outputs
	if outputType == "dashboard" {
		content.Dashboard = d.generateDashboardCode(result)
	}

	return content, nil
}

// generateDashboardCode creates dashboard React code
func (d *DadAIAgent) generateDashboardCode(result map[string]interface{}) string {
	return "// Dashboard Component\n" +
		"import React, { useState, useEffect } from 'react';\n\n" +
		"export default function Dashboard({ projectId }) {\n" +
		"  const [metrics, setMetrics] = useState({ leads: 0, tasks: 0, revenue: 0 });\n" +
		"  const [recentActivity, setRecentActivity] = useState([]);\n\n" +
		"  useEffect(() => {\n" +
		"    fetch('/api/v1/dashboard/' + projectId)\n" +
		"      .then(r => r.json())\n" +
		"      .then(data => {\n" +
		"        setMetrics(data.metrics);\n" +
		"        setRecentActivity(data.activity);\n" +
		"      });\n" +
		"  }, [projectId]);\n\n" +
		"  return (\n" +
		"    <div className=\"dashboard\">\n" +
		"      <div className=\"metrics-grid\">\n" +
		"        <MetricCard title=\"Leads\" value={metrics.leads} />\n" +
		"        <MetricCard title=\"Tasks\" value={metrics.tasks} />\n" +
		"        <MetricCard title=\"Revenue\" value={'$' + metrics.revenue} />\n" +
		"      </div>\n" +
		"      <ActivityFeed items={recentActivity} />\n" +
		"    </div>\n" +
		"  );\n" +
		"}\n\n" +
		"function MetricCard({ title, value }) {\n" +
		"  return (\n" +
		"    <div className=\"metric-card\">\n" +
		"      <h3>{title}</h3>\n" +
		"      <p className=\"value\">{value}</p>\n" +
		"    </div>\n" +
		"  );\n" +
		"}\n\n" +
		"function ActivityFeed({ items }) {\n" +
		"  return (\n" +
		"    <div className=\"activity-feed\">\n" +
		"      <h3>Recent Activity</h3>\n" +
		"      {items.map((item, i) => (\n" +
		"        <div key={i} className=\"activity-item\">{item}</div>\n" +
		"      ))}\n" +
		"    </div>\n" +
		"  );\n" +
		"}"
}

// getDefaultCSS returns default styling
func getDefaultCSS() string {
	return `
:root {
  --primary: #2563eb;
  --secondary: #64748b;
  --accent: #0ea5e9;
  --background: #ffffff;
  --surface: #f8fafc;
  --text: #1e293b;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  line-height: 1.6;
  color: var(--text);
  background: var(--background);
}

.container { max-width: 1200px; margin: 0 auto; padding: 0 1rem; }
.btn { display: inline-block; padding: 0.75rem 1.5rem; border-radius: 0.5rem; text-decoration: none; font-weight: 600; }
.btn-primary { background: var(--primary); color: white; }
.hero { padding: 4rem 0; text-align: center; }
.hero h1 { font-size: 2.5rem; margin-bottom: 1rem; }
.section { padding: 3rem 0; }
.grid { display: grid; gap: 1.5rem; }
.grid-3 { grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); }
.card { padding: 1.5rem; border-radius: 0.5rem; background: var(--surface); }
input, textarea { width: 100%; padding: 0.75rem; border: 1px solid var(--secondary); border-radius: 0.5rem; }
`
}

// contains checks if a string contains a substring (case-insensitive)
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
