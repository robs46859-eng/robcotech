package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/auth"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/providers"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/routing"
)

// InferenceRequest represents a request to the inference gateway
type InferenceRequest struct {
	Model      string                 `json:"model,omitempty"`
	Messages   []Message              `json:"messages"`
	Temperature float32               `json:"temperature,omitempty"`
	MaxTokens  int                    `json:"max_tokens,omitempty"`
	Stream     bool                   `json:"stream,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// InferenceResponse represents the response from the inference gateway
type InferenceResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
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

// InferenceHandler handles POST /v1/ai requests
// This is the main gateway entry point for all inference requests
// 
// Implements the cost ladder:
// 1. Check semantic cache (free)
// 2. Route to cheapest viable model
// 3. Escalate only if confidence/policy requires
func InferenceHandler(w http.ResponseWriter, r *http.Request) {
	// Get tenant from context (set by auth middleware)
	tenantID, ok := r.Context().Value(auth.TenantIDKey).(string)
	if !ok {
		// No tenant in context - use default for development
		tenantID = "default"
	}
	
	// Parse request
	var req InferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	
	// Step 1: Check semantic cache
	cacheResult, cacheHit := checkSemanticCache(r.Context(), tenantID, req)
	if cacheHit {
		// Return cached response
		response := InferenceResponse{
			ID:      "cache-" + time.Now().Format("20060102150405"),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "cache",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: cacheResult.Content,
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Hit", "true")
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Step 2: Classify request and select model tier
	modelTier := routing.SelectModelTier(r.Context(), req, tenantID)
	
	// Step 3: Execute inference through selected provider
	result, err := executeInference(r.Context(), req, modelTier)
	if err != nil {
		// Step 4: Fallback to next tier if available
		modelTier = routing.GetFallbackTier(modelTier)
		if modelTier != "" {
			result, err = executeInference(r.Context(), req, modelTier)
		}
		
		if err != nil {
			http.Error(w, `{"error": "inference failed: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
	}
	
	// Step 5: Record usage for billing
	recordUsage(r.Context(), tenantID, req.Model, result)
	
	// Step 6: Store in cache if cacheable
	if shouldCache(req, result) {
		storeInCache(r.Context(), tenantID, req, result)
	}
	
	// Build response
	response := InferenceResponse{
		ID:      result.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   result.Model,
		Choices: result.Choices,
		Usage:   result.Usage,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Model-Tier", modelTier)
	w.Header().Set("X-Cache-Hit", "false")
	json.NewEncoder(w).Encode(response)
}

// CacheResult holds cached response data
type CacheResult struct {
	Content string
}

// InferenceResult holds the result from model execution
type InferenceResult struct {
	ID      string
	Model   string
	Choices []Choice
	Usage   Usage
}

// checkSemanticCache checks if request has a cached response
func checkSemanticCache(ctx context.Context, tenantID string, req InferenceRequest) (*CacheResult, bool) {
	// TODO: Call semantic cache service
	// For now, always return cache miss
	return nil, false
}

// executeInference calls the selected model provider
func executeInference(ctx context.Context, req InferenceRequest, modelTier string) (*InferenceResult, error) {
	// Map tier to actual model
	model := routing.GetModelForTier(modelTier)
	
	// Get API keys from environment
	apiKeys := map[string]string{
		"anthropic":  os.Getenv("ANTHROPIC_API_KEY"),
		"openai":     os.Getenv("OPENAI_API_KEY"),
		"google":     os.Getenv("GOOGLE_API_KEY"),
	}
	
	// Get provider for model
	provider, err := providers.GetProviderForModel(model, apiKeys)
	if err != nil {
		// Fall back to local provider if no API key configured
		provider = providers.NewLocalProvider()
	}
	
	// Convert request to provider format
	providerReq := providers.CompletionRequest{
		Model:       model,
		Messages:    make([]providers.Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}
	
	for i, msg := range req.Messages {
		providerReq.Messages[i] = providers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	
	// Call provider
	resp, err := provider.Complete(ctx, providerReq)
	if err != nil {
		return nil, err
	}
	
	// Convert response to our format
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = Choice{
			Index: choice.Index,
			Message: Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}
	
	return &InferenceResult{
		ID:      resp.ID,
		Model:   resp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// recordUsage records token usage for billing
func recordUsage(ctx context.Context, tenantID, model string, result *InferenceResult) {
	// TODO: Call billing service to record usage
	_ = ctx
	_ = tenantID
	_ = model
	_ = result
}

// shouldCache determines if response should be cached
func shouldCache(req InferenceRequest, result *InferenceResult) bool {
	// Don't cache streaming responses
	if req.Stream {
		return false
	}
	// Don't cache if temperature is high (non-deterministic)
	if req.Temperature > 0.7 {
		return false
	}
	return true
}

// storeInCache stores response in semantic cache
func storeInCache(ctx context.Context, tenantID string, req InferenceRequest, result *InferenceResult) {
	// TODO: Call semantic cache service to store
	_ = ctx
	_ = tenantID
	_ = req
	_ = result
}

// ListTenantsHandler handles GET /v1/tenants
func ListTenantsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tenant listing
	response := map[string]interface{}{
		"tenants": []string{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTenantHandler handles GET /v1/tenants/{id}
func GetTenantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["id"]

	// TODO: Implement tenant retrieval
	response := map[string]interface{}{
		"id": tenantID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUsageHandler handles GET /v1/usage
func GetUsageHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement usage retrieval
	response := map[string]interface{}{
		"tokens_used": 0,
		"requests":    0,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBillingHandler handles GET /v1/billing
func GetBillingHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement billing retrieval
	response := map[string]interface{}{
		"current_charges": 0.0,
		"currency":        "USD",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
