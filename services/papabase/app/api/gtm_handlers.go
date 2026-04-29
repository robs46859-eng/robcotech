package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
)

// GenerateGTMStrategyRequest is the request body for GTM strategy generation
type GenerateGTMStrategyRequest struct {
	ProductDescription string `json:"product_description"`
	TargetMarket       string `json:"target_market"`
}

// GenerateGTMStrategyHandler handles POST /api/v1/ai/gtm/strategy
func GenerateGTMStrategyHandler(w http.ResponseWriter, r *http.Request) {
	gtmAgent := r.Context().Value(ContextKey("gtm_agent")).(*agents.GTMAgent)

	var req GenerateGTMStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	response, err := gtmAgent.GenerateStrategy(r.Context(), req.ProductDescription, req.TargetMarket)
	if err != nil {
		http.Error(w, `{"error": "failed to generate strategy: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"result":    response,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GenerateGTMOutboundRequest is the request body for GTM outbound generation
type GenerateGTMOutboundRequest struct {
	ICP        string `json:"icp"`
	ValueProp  string `json:"value_prop"`
}

// GenerateGTMOutboundHandler handles POST /api/v1/ai/gtm/outbound
func GenerateGTMOutboundHandler(w http.ResponseWriter, r *http.Request) {
	gtmAgent := r.Context().Value(ContextKey("gtm_agent")).(*agents.GTMAgent)

	var req GenerateGTMOutboundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	response, err := gtmAgent.GenerateOutbound(r.Context(), req.ICP, req.ValueProp)
	if err != nil {
		http.Error(w, `{"error": "failed to generate outbound strategy: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"result":    response,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
