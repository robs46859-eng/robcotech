package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
)

// GenerateCodeRequest is the request body for feature implementation
type GenerateCodeRequest struct {
	Requirement   string `json:"requirement"`
	TechnicalStack string `json:"technical_stack"`
}

// GenerateCodeHandler handles POST /api/v1/ai/code/generate
func GenerateCodeHandler(w http.ResponseWriter, r *http.Request) {
	codeAgent := r.Context().Value(ContextKey("code_agent")).(*agents.CodeAgent)

	var req GenerateCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	response, err := codeAgent.GenerateFeature(r.Context(), req.Requirement, req.TechnicalStack)
	if err != nil {
		http.Error(w, `{"error": "failed to generate code: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"result":    response,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RefactorCodeRequest is the request body for refactoring
type RefactorCodeRequest struct {
	ExistingCode string `json:"existing_code"`
	Objective    string `json:"objective"`
}

// RefactorCodeHandler handles POST /api/v1/ai/code/refactor
func RefactorCodeHandler(w http.ResponseWriter, r *http.Request) {
	codeAgent := r.Context().Value(ContextKey("code_agent")).(*agents.CodeAgent)

	var req RefactorCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	response, err := codeAgent.RefactorCode(r.Context(), req.ExistingCode, req.Objective)
	if err != nil {
		http.Error(w, `{"error": "failed to refactor code: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"result":    response,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
