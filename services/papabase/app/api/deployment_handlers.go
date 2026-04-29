package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
)

// PlanAzureDeploymentRequest is the request body for deployment planning
type PlanAzureDeploymentRequest struct {
	RepoContext string `json:"repo_context"`
}

// PlanAzureDeploymentHandler handles POST /api/v1/ai/deploy/azure
func PlanAzureDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	deploymentBot := r.Context().Value(ContextKey("azure_deployment_bot")).(*agents.AzureDeploymentBot)

	var req PlanAzureDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	response, err := deploymentBot.PlanAzureWebDeployment(r.Context(), req.RepoContext)
	if err != nil {
		http.Error(w, `{"error": "failed to plan azure deployment: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"result":    response,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DebugErrorRequest is the request body for debugging
type DebugErrorRequest struct {
	ErrorMessage string `json:"error_message"`
	CodeContext  string `json:"code_context"`
}

// DebugErrorHandler handles POST /api/v1/ai/debug
func DebugErrorHandler(w http.ResponseWriter, r *http.Request) {
	debugAgent := r.Context().Value(ContextKey("debug_agent")).(*agents.DebugAgent)

	var req DebugErrorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	response, err := debugAgent.DebugError(r.Context(), req.ErrorMessage, req.CodeContext)
	if err != nil {
		http.Error(w, `{"error": "failed to debug: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"result":    response,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
