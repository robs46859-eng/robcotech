// Package agents implements AI agents for Papabase
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

// LeadScore represents the AI-generated lead qualification
type LeadScore struct {
	Score           int     `json:"score"`            // 0-100
	ConversionProb  float64 `json:"conversion_prob"`  // 0.0-1.0
	RecommendedAction string  `json:"recommended_action"`
	Reasoning       string  `json:"reasoning"`
	Tier            string  `json:"tier"` // hot, warm, cold
}

// LeadScorer analyzes and scores leads using AI
type LeadScorer struct {
	gatewayURL string
	httpClient *http.Client
}

// NewLeadScorer creates a new lead scoring agent
func NewLeadScorer(gatewayURL string) *LeadScorer {
	return &LeadScorer{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ScoreLead analyzes a lead and returns a qualification score
func (ls *LeadScorer) ScoreLead(ctx context.Context, lead *Lead) (*LeadScore, error) {
	prompt := ls.buildScoringPrompt(lead)

	response, err := ls.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ls.parseScoreResponse(response)
}

// ScoreLeadsBatch scores multiple leads in one call
func (ls *LeadScorer) ScoreLeadsBatch(ctx context.Context, leads []*Lead) ([]*LeadScore, error) {
	prompt := ls.buildBatchScoringPrompt(leads)

	response, err := ls.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ls.parseBatchScoreResponse(response, len(leads))
}

// GetLeadInsights returns AI insights about a specific lead
func (ls *LeadScorer) GetLeadInsights(ctx context.Context, lead *Lead) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this business lead and provide insights:

Lead Information:
- Name: %s
- Company: %s
- Email: %s
- Phone: %s
- Source: %s
- Notes: %s
- Status: %s

Provide insights on:
1. Best contact method and timing
2. Likely service needs based on profile
3. Potential objections or concerns
4. Recommended next steps
5. Upsell opportunities

Return as JSON with keys: contact_strategy, likely_needs, objections, next_steps, upsell_opportunities`,
		lead.Name, lead.Company, lead.Email, lead.Phone, lead.Source, lead.Notes, lead.Status)

	response, err := ls.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var insights map[string]interface{}
	if err := json.Unmarshal([]byte(response), &insights); err != nil {
		return map[string]interface{}{"analysis": response}, nil
	}

	return insights, nil
}

// PredictLeadOutcome predicts the likely outcome of a lead
func (ls *LeadScorer) PredictLeadOutcome(ctx context.Context, lead *Lead) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Predict the outcome for this business lead:

Lead Information:
- Name: %s
- Company: %s
- Source: %s
- Status: %s
- Notes: %s

Based on historical patterns, predict:
1. Likely outcome (converted, lost, stalled)
2. Estimated time to conversion (days)
3. Estimated project value range
4. Confidence level

Return as JSON with keys: predicted_outcome, estimated_days, estimated_value, confidence`,
		lead.Name, lead.Company, lead.Source, lead.Status, lead.Notes)

	response, err := ls.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var prediction map[string]interface{}
	if err := json.Unmarshal([]byte(response), &prediction); err != nil {
		return map[string]interface{}{"prediction": response}, nil
	}

	return prediction, nil
}

func (ls *LeadScorer) buildScoringPrompt(lead *Lead) string {
	return fmt.Sprintf(`You are a lead qualification expert. Score this business lead from 0-100.

Lead Information:
- Name: %s
- Company: %s
- Email: %s
- Phone: %s
- Source: %s
- Notes: %s
- Current Status: %s

Scoring Criteria:
- 80-100 (Hot): Complete info, strong intent, good company, warm source
- 60-79 (Warm): Good info, some interest, decent company
- 40-59 (Cool): Basic info, unclear intent
- 0-39 (Cold): Missing info, low quality, spam indicators

Consider:
1. Contact completeness (email + phone = higher score)
2. Company name quality (real business vs generic)
3. Lead source (referral = higher, cold = lower)
4. Notes showing intent or budget mentions
5. Professional email domain

Return ONLY valid JSON:
{
  "score": <number 0-100>,
  "conversion_prob": <number 0.0-1.0>,
  "recommended_action": "<string: immediate_call, email_followup, nurture, disqualify>",
  "reasoning": "<string: brief explanation>",
  "tier": "<string: hot, warm, cold>"
}`,
		lead.Name,
		lead.Company,
		lead.Email,
		lead.Phone,
		lead.Source,
		lead.Notes,
		lead.Status)
}

func (ls *LeadScorer) buildBatchScoringPrompt(leads []*Lead) string {
	var sb strings.Builder
	sb.WriteString("Score these business leads from 0-100. Return a JSON array.\n\n")

	for i, lead := range leads {
		sb.WriteString(fmt.Sprintf("Lead %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("- Name: %s\n", lead.Name))
		sb.WriteString(fmt.Sprintf("- Company: %s\n", lead.Company))
		sb.WriteString(fmt.Sprintf("- Source: %s\n", lead.Source))
		sb.WriteString(fmt.Sprintf("- Notes: %s\n", lead.Notes))
		sb.WriteString("\n")
	}

	sb.WriteString("Return JSON array of objects with keys: score, conversion_prob, recommended_action, reasoning, tier")

	return sb.String()
}

func (ls *LeadScorer) parseScoreResponse(response string) (*LeadScore, error) {
	var score LeadScore

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(response), &score); err != nil {
		// Extract score from text response
		score = ls.extractScoreFromText(response)
	}

	// Validate score
	if score.Score < 0 {
		score.Score = 0
	}
	if score.Score > 100 {
		score.Score = 100
	}
	if score.ConversionProb < 0 {
		score.ConversionProb = 0
	}
	if score.ConversionProb > 1 {
		score.ConversionProb = 1
	}

	// Set tier based on score if not provided
	if score.Tier == "" {
		if score.Score >= 80 {
			score.Tier = "hot"
		} else if score.Score >= 60 {
			score.Tier = "warm"
		} else {
			score.Tier = "cold"
		}
	}

	// Set default action if not provided
	if score.RecommendedAction == "" {
		switch score.Tier {
		case "hot":
			score.RecommendedAction = "immediate_call"
		case "warm":
			score.RecommendedAction = "email_followup"
		default:
			score.RecommendedAction = "nurture"
		}
	}

	return &score, nil
}

func (ls *LeadScorer) parseBatchScoreResponse(response string, count int) ([]*LeadScore, error) {
	var scores []*LeadScore

	if err := json.Unmarshal([]byte(response), &scores); err != nil {
		// If parsing fails, return a single score
		score, err := ls.parseScoreResponse(response)
		if err != nil {
			return nil, err
		}
		return []*LeadScore{score}, nil
	}

	return scores, nil
}

func (ls *LeadScorer) extractScoreFromText(text string) LeadScore {
	score := LeadScore{
		Score:  50,
		Tier:   "warm",
	}

	// Look for score patterns in text
	text = strings.ToLower(text)

	if strings.Contains(text, "score") {
		// Try to extract number after "score"
		parts := strings.Split(text, "score")
		if len(parts) > 1 {
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if len(part) > 0 {
					// Extract first number
					for i, c := range part {
						if c >= '0' && c <= '9' {
							// Found a digit, extract the number
							numStr := ""
							for j := i; j < len(part) && part[j] >= '0' && part[j] <= '9'; j++ {
								numStr += string(part[j])
							}
							if numStr != "" {
								fmt.Sscanf(numStr, "%d", &score.Score)
								break
							}
						}
					}
					break
				}
			}
		}
	}

	// Look for keywords to determine tier
	if strings.Contains(text, "hot") || strings.Contains(text, "excellent") || strings.Contains(text, "high") {
		score.Tier = "hot"
		if score.Score < 70 {
			score.Score = 80
		}
	} else if strings.Contains(text, "cold") || strings.Contains(text, "poor") || strings.Contains(text, "low") {
		score.Tier = "cold"
		if score.Score > 50 {
			score.Score = 30
		}
	}

	score.ConversionProb = float64(score.Score) / 100.0
	score.Reasoning = text[:min(200, len(text))]

	return score
}

func (ls *LeadScorer) callGateway(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a business lead qualification expert. Return responses in valid JSON format."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  500,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/ai", ls.gatewayURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "papabase-service-key")

	resp, err := ls.httpClient.Do(req)
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

	if choices, ok := gatewayResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return strings.TrimSpace(content), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no content in gateway response")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
