// Package routing implements model selection and cost ladder routing
package routing

import (
	"context"
	"strings"
)

// Model tiers from cheapest to most expensive
const (
	TierLocal      = "local"       // Local models - ~$0.0001/1K tokens
	TierCheap      = "cheap"       // Haiku, GPT-3.5 - ~$0.0005/1K tokens
	TierMid        = "mid"         // Sonnet, GPT-4 - ~$0.003/1K tokens
	TierPremium    = "premium"     // Opus, GPT-4-turbo - ~$0.03/1K tokens
)

// TaskClassification represents the classified task type
type TaskClassification struct {
	Tier         string
	TaskType     string
	Confidence   float64
	ShouldEscalate bool
}

// SelectModelTier classifies the request and selects the appropriate model tier
// 
// Implements the cost ladder from the master architecture:
// - Tier 0: Deterministic (no LLM) - validation, field mapping, status checks
// - Tier 1: Very cheap - classification, intent detection, tagging, rewriting
// - Tier 2: Mid-cost - moderate reasoning, messy extraction, controlled drafting
// - Tier 3: Premium - high-stakes reasoning, complex synthesis, ambiguous planning
func SelectModelTier(ctx context.Context, req interface{}, tenantID string) string {
	// Extract request details (type assertion for InferenceRequest)
	messages := extractMessages(req)
	
	// Classify the task based on message content
	classification := classifyTask(messages)
	
	// Check tenant policy for model restrictions
	// (premium tenants might skip straight to higher tiers)
	tenantTier := getTenantTier(tenantID)
	
	// Apply escalation rules
	if classification.ShouldEscalate {
		return escalateTier(classification.Tier)
	}
	
	// Return the minimum of requested tier and tenant-allowed tier
	return minTier(classification.Tier, tenantTier)
}

// extractMessages extracts messages from various request types
func extractMessages(req interface{}) []string {
	var messages []string
	
	// Type switch to handle different request structures
	switch r := req.(type) {
	case map[string]interface{}:
		if msgs, ok := r["messages"].([]interface{}); ok {
			for _, m := range msgs {
				if msg, ok := m.(map[string]interface{}); ok {
					if content, ok := msg["content"].(string); ok {
						messages = append(messages, content)
					}
				}
			}
		}
	}
	
	return messages
}

// classifyTask analyzes messages to determine task type and appropriate tier
func classifyTask(messages []string) TaskClassification {
	if len(messages) == 0 {
		return TaskClassification{
			Tier:       TierMid,
			TaskType:   "unknown",
			Confidence: 0.5,
		}
	}
	
	// Analyze the last message (user's request)
	request := strings.ToLower(messages[len(messages)-1])
	
	// Tier 0: Deterministic tasks (should be handled before LLM)
	if isDeterministic(request) {
		return TaskClassification{
			Tier:       TierLocal,
			TaskType:   "deterministic",
			Confidence: 0.9,
		}
	}
	
	// Tier 1: Simple tasks suitable for cheap models
	if isSimpleTask(request) {
		return TaskClassification{
			Tier:       TierCheap,
			TaskType:   "simple",
			Confidence: 0.8,
		}
	}
	
	// Tier 2: Moderate reasoning tasks
	if isModerateTask(request) {
		return TaskClassification{
			Tier:       TierMid,
			TaskType:   "moderate",
			Confidence: 0.7,
		}
	}
	
	// Tier 3: Complex tasks requiring premium models
	return TaskClassification{
		Tier:       TierPremium,
		TaskType:   "complex",
		Confidence: 0.6,
	}
}

// isDeterministic checks if task can be handled without LLM
func isDeterministic(request string) bool {
	deterministicKeywords := []string{
		"validate", "format", "parse json", "check schema",
		"convert to", "transform", "extract field",
		"calculate", "count", "sum",
	}
	
	for _, keyword := range deterministicKeywords {
		if strings.Contains(request, keyword) {
			return true
		}
	}
	return false
}

// isSimpleTask checks if task is suitable for cheap models
func isSimpleTask(request string) bool {
	simpleKeywords := []string{
		"classify", "categorize", "tag", "label",
		"summarize", "rewrite", "paraphrase",
		"extract", "find", "search",
		"sentiment", "intent", "language",
		"translate", "fix grammar", "spell check",
	}
	
	for _, keyword := range simpleKeywords {
		if strings.Contains(request, keyword) {
			return true
		}
	}
	return false
}

// isModerateTask checks if task needs mid-tier models
func isModerateTask(request string) bool {
	moderateKeywords := []string{
		"analyze", "compare", "explain", "describe",
		"generate code", "write function", "debug",
		"extract from", "parse pdf", "read document",
		"draft email", "write report", "create outline",
	}
	
	for _, keyword := range moderateKeywords {
		if strings.Contains(request, keyword) {
			return true
		}
	}
	return false
}

// getTenantTier returns the maximum tier allowed for a tenant
func getTenantTier(tenantID string) string {
	// TODO: Look up tenant plan and return allowed tier
	// Free tier: TierCheap
	// Basic tier: TierMid
	// Pro/Enterprise: TierPremium
	
	// Default to allowing all tiers for now
	return TierPremium
}

// escalateTier returns the next higher tier for escalation
func escalateTier(currentTier string) string {
	switch currentTier {
	case TierLocal:
		return TierCheap
	case TierCheap:
		return TierMid
	case TierMid:
		return TierPremium
	default:
		return TierPremium
	}
}

// GetFallbackTier returns the fallback tier when a request fails
func GetFallbackTier(currentTier string) string {
	// Escalate to next tier on failure
	return escalateTier(currentTier)
}

// GetModelForTier maps a tier to a specific model name
func GetModelForTier(tier string) string {
	switch tier {
	case TierLocal:
		return "gemini-2.0-flash"
	case TierCheap:
		return "gemini-2.0-flash"
	case TierMid:
		return "gemini-2.0-flash"
	case TierPremium:
		return "gemini-2.0-flash"
	default:
		return "gemini-2.0-flash"
	}
}

// minTier returns the cheaper of two tiers
func minTier(tier1, tier2 string) string {
	tierOrder := map[string]int{
		TierLocal:   0,
		TierCheap:   1,
		TierMid:     2,
		TierPremium: 3,
	}
	
	if tierOrder[tier1] <= tierOrder[tier2] {
		return tier1
	}
	return tier2
}
