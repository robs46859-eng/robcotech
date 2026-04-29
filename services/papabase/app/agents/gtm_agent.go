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

// GTMResponse represents an AI-generated Go-To-Market strategy or artifact
type GTMResponse struct {
	CampaignName string            `json:"campaign_name"`
	Strategy     string            `json:"strategy"`
	Artifacts    map[string]string `json:"artifacts,omitempty"`
	Metrics      []string          `json:"metrics,omitempty"`
	NextSteps    []string          `json:"next_steps,omitempty"`
}

// GTMAgent handles marketing, sales, and operations workflows
type GTMAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// NewGTMAgent creates a new GTM agent
func NewGTMAgent(gatewayURL string) *GTMAgent {
	return &GTMAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// GenerateStrategy generates positioning, ICP, pricing, and sales motion
func (ga *GTMAgent) GenerateStrategy(ctx context.Context, productDescription string, targetMarket string) (*GTMResponse, error) {
	prompt := fmt.Sprintf(`Develop a comprehensive Go-To-Market strategy for:

Product: %s
Target Market: %s

Please cover the following pillars:
1. Positioning & ICP (Ideal Customer Profile)
2. AI-Driven Pricing Strategy
3. Sales Motion Design

Return JSON with keys: campaign_name, strategy, metrics, next_steps.`,
		productDescription, targetMarket)

	response, err := ga.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ga.parseGTMResponse(response)
}

// GenerateOutbound generates cold outreach, SDR scripts, and enrichment strategies
func (ga *GTMAgent) GenerateOutbound(ctx context.Context, icp string, valueProp string) (*GTMResponse, error) {
	prompt := fmt.Sprintf(`Create an Outbound signal-based prospecting campaign:

ICP: %s
Value Proposition: %s

Include:
1. AI Cold Outreach templates
2. AI SDR logic
3. Lead Enrichment workflows
4. Video Outreach scripts

Return JSON with keys: campaign_name, strategy, artifacts (map of artifact name to content), next_steps.`,
		icp, valueProp)

	response, err := ga.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ga.parseGTMResponse(response)
}

// GenerateInbound creates multi-platform launch, SEO, and content strategies
func (ga *GTMAgent) GenerateInbound(ctx context.Context, productDescription string) (*GTMResponse, error) {
	prompt := fmt.Sprintf(`Design an Inbound marketing campaign for:

Product: %s

Cover:
1. Multi-platform launch plan
2. AI SEO strategy
3. Social Selling guidelines
4. Content-to-pipeline roadmap

Return JSON with keys: campaign_name, strategy, artifacts (map of artifact name to content), next_steps.`,
		productDescription)

	response, err := ga.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ga.parseGTMResponse(response)
}

// GeneratePaid designs paid creative and UGC ad strategies
func (ga *GTMAgent) GeneratePaid(ctx context.Context, campaignGoal string, budget string) (*GTMResponse, error) {
	prompt := fmt.Sprintf(`Formulate a Paid acquisition strategy:

Goal: %s
Budget: %s

Include:
1. AI UGC Ads briefs
2. Paid Creative AI generation prompts

Return JSON with keys: campaign_name, strategy, artifacts (map of artifact name to content), next_steps.`,
		campaignGoal, budget)

	response, err := ga.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ga.parseGTMResponse(response)
}

// GenerateRetention creates expansion and partner strategies
func (ga *GTMAgent) GenerateRetention(ctx context.Context, currentUserBase string) (*GTMResponse, error) {
	prompt := fmt.Sprintf(`Build a Retention and Expansion strategy for:

Current User Base: %s

Focus on:
1. Expansion & Retention tactics
2. Partner/Affiliate programs

Return JSON with keys: campaign_name, strategy, next_steps.`,
		currentUserBase)

	response, err := ga.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ga.parseGTMResponse(response)
}

// GenerateOperations formulates the automation layer
func (ga *GTMAgent) GenerateOperations(ctx context.Context, teamSize string, stack string) (*GTMResponse, error) {
	prompt := fmt.Sprintf(`Design the GTM Operations and Automation layer:

Team Size: %s
Current Stack: %s

Include:
1. GTM Engineering requirements
2. Solo-founder GTM automation workflow
3. Key GTM Metrics dashboard

Return JSON with keys: campaign_name, strategy, metrics, next_steps.`,
		teamSize, stack)

	response, err := ga.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ga.parseGTMResponse(response)
}

func (ga *GTMAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":  "gemini-2.0-flash",
		"prompt": prompt,
		"system": "You are an expert GTM (Go-To-Market) agent specializing in strategy, outbound, inbound, paid, retention, and operations. Always return valid JSON matching the requested structure.",
		"temperature": 0.7,
	}
	
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", ga.gatewayURL+"/v1/ai", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer fsa_dev_key_12345")
	
	resp, err := ga.httpClient.Do(req)
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

func (ga *GTMAgent) parseGTMResponse(response string) (*GTMResponse, error) {
	// Clean markdown formatting if present
	response = strings.TrimPrefix(response, "```json\n")
	response = strings.TrimSuffix(response, "\n```")
	response = strings.TrimSpace(response)
	
	var result GTMResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse GTM response: %w", err)
	}
	
	return &result, nil
}
