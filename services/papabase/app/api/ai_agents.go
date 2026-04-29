package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
)

// ============================================================================
// AI Lead Scoring Endpoints
// ============================================================================

// ScoreLeadRequest is the request to score a lead
type ScoreLeadRequest struct {
	LeadID string `json:"lead_id"`
}

// ScoreLeadHandler handles POST /api/v1/ai/leads/score
func ScoreLeadHandler(w http.ResponseWriter, r *http.Request) {
	leadScorer := r.Context().Value(ContextKey("lead_scorer")).(*agents.LeadScorer)
	store := r.Context().Value(CRMStoreKey).(CRMStore)

	var req ScoreLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	lead, err := store.GetLead(r.Context(), req.LeadID)
	if err != nil {
		http.Error(w, `{"error": "lead not found"}`, http.StatusNotFound)
		return
	}

	score, err := leadScorer.ScoreLead(r.Context(), lead)
	if err != nil {
		http.Error(w, `{"error": "scoring failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"lead_id":   req.LeadID,
		"score":     score,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ScoreLeadsBatchHandler handles POST /api/v1/ai/leads/score/batch
func ScoreLeadsBatchHandler(w http.ResponseWriter, r *http.Request) {
	leadScorer := r.Context().Value(ContextKey("lead_scorer")).(*agents.LeadScorer)
	store := r.Context().Value(CRMStoreKey).(CRMStore)

	leads, err := store.ListLeads(r.Context(), "")
	if err != nil {
		http.Error(w, `{"error": "failed to list leads"}`, http.StatusInternalServerError)
		return
	}

	scores, err := leadScorer.ScoreLeadsBatch(r.Context(), leads)
	if err != nil {
		http.Error(w, `{"error": "batch scoring failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total": len(scores),
		"scores": scores,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLeadInsightsHandler handles GET /api/v1/ai/leads/{id}/insights
func GetLeadInsightsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	leadID := vars["id"]

	leadScorer := r.Context().Value(ContextKey("lead_scorer")).(*agents.LeadScorer)
	store := r.Context().Value(CRMStoreKey).(CRMStore)

	lead, err := store.GetLead(r.Context(), leadID)
	if err != nil {
		http.Error(w, `{"error": "lead not found"}`, http.StatusNotFound)
		return
	}

	insights, err := leadScorer.GetLeadInsights(r.Context(), lead)
	if err != nil {
		http.Error(w, `{"error": "failed to get insights: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"lead_id":   leadID,
		"insights":  insights,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// AI Task Agent Endpoints
// ============================================================================

// GenerateTaskRequest is the request to generate a task from a note
type GenerateTaskRequest struct {
	Note    string            `json:"note"`
	Context map[string]string `json:"context"`
}

// GenerateTaskHandler handles POST /api/v1/ai/tasks/generate
func GenerateTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskAgent := r.Context().Value(ContextKey("task_agent")).(*agents.TaskAgent)

	var req GenerateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	suggestion, err := taskAgent.GenerateTaskFromNote(r.Context(), req.Note, req.Context)
	if err != nil {
		http.Error(w, `{"error": "task generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"suggestion": suggestion,
		"generated":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BreakdownTaskHandler handles POST /api/v1/ai/tasks/breakdown
func BreakdownTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskAgent := r.Context().Value(ContextKey("task_agent")).(*agents.TaskAgent)

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	subtasks, err := taskAgent.SuggestTaskBreakdown(r.Context(), req.Title, req.Description)
	if err != nil {
		http.Error(w, `{"error": "breakdown failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"title":    req.Title,
		"subtasks": subtasks,
		"count":    len(subtasks),
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// AI Content Generator Endpoints
// ============================================================================

// GenerateBusinessDescriptionHandler handles POST /api/v1/ai/content/business-description
func GenerateBusinessDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	contentAgent := r.Context().Value(ContextKey("content_agent")).(*agents.ContentAgent)

	var req struct {
		BusinessName string `json:"business_name"`
		Industry     string `json:"industry"`
		KeyServices  string `json:"key_services"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	content, err := contentAgent.GenerateBusinessDescription(r.Context(), req.BusinessName, req.Industry, req.KeyServices)
	if err != nil {
		http.Error(w, `{"error": "generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"content":  content,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateSEOMetaHandler handles POST /api/v1/ai/content/seo-meta
func GenerateSEOMetaHandler(w http.ResponseWriter, r *http.Request) {
	contentAgent := r.Context().Value(ContextKey("content_agent")).(*agents.ContentAgent)

	var req struct {
		PageContent string `json:"page_content"`
		Keywords    string `json:"keywords"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	metaTags, err := contentAgent.GenerateSEOMetaTags(r.Context(), req.PageContent, req.Keywords)
	if err != nil {
		http.Error(w, `{"error": "generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"meta_tags": metaTags,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateBlogPostHandler handles POST /api/v1/ai/content/blog
func GenerateBlogPostHandler(w http.ResponseWriter, r *http.Request) {
	contentAgent := r.Context().Value(ContextKey("content_agent")).(*agents.ContentAgent)

	var req struct {
		Topic   string `json:"topic"`
		Outline string `json:"outline"`
		Tone    string `json:"tone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	content, err := contentAgent.GenerateBlogPost(r.Context(), req.Topic, req.Outline, req.Tone)
	if err != nil {
		http.Error(w, `{"error": "generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"blog_post": content,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateSocialMediaHandler handles POST /api/v1/ai/content/social
func GenerateSocialMediaHandler(w http.ResponseWriter, r *http.Request) {
	contentAgent := r.Context().Value(ContextKey("content_agent")).(*agents.ContentAgent)

	var req struct {
		Topic     string `json:"topic"`
		Platforms string `json:"platforms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	posts, err := contentAgent.GenerateSocialMediaPosts(r.Context(), req.Topic, req.Platforms)
	if err != nil {
		http.Error(w, `{"error": "generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"posts":     posts,
		"count":     len(posts),
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateFAQHandler handles POST /api/v1/ai/content/faq
func GenerateFAQHandler(w http.ResponseWriter, r *http.Request) {
	contentAgent := r.Context().Value(ContextKey("content_agent")).(*agents.ContentAgent)

	var req struct {
		BusinessType     string `json:"business_type"`
		CommonQuestions  string `json:"common_questions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	faqs, err := contentAgent.GenerateFAQ(r.Context(), req.BusinessType, req.CommonQuestions)
	if err != nil {
		http.Error(w, `{"error": "generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"faqs":      faqs,
		"count":     len(faqs),
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// AI Proposal Builder Endpoints
// ============================================================================

// GenerateProposalHandler handles POST /api/v1/ai/proposals/generate
func GenerateProposalHandler(w http.ResponseWriter, r *http.Request) {
	proposalAgent := r.Context().Value(ContextKey("proposal_agent")).(*agents.ProposalAgent)
	store := r.Context().Value(CRMStoreKey).(CRMStore)

	var req struct {
		LeadID       string `json:"lead_id"`
		Requirements string `json:"requirements"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	lead, err := store.GetLead(r.Context(), req.LeadID)
	if err != nil {
		http.Error(w, `{"error": "lead not found"}`, http.StatusNotFound)
		return
	}

	proposal, err := proposalAgent.GenerateProposal(r.Context(), lead, req.Requirements)
	if err != nil {
		http.Error(w, `{"error": "proposal generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	proposal.ID = uuid.New().String()

	response := map[string]interface{}{
		"proposal":  proposal,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateQuoteHandler handles POST /api/v1/ai/proposals/quote
func GenerateQuoteHandler(w http.ResponseWriter, r *http.Request) {
	proposalAgent := r.Context().Value(ContextKey("proposal_agent")).(*agents.ProposalAgent)

	var req struct {
		Services  []string          `json:"services"`
		ClientInfo map[string]string `json:"client_info"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	quote, err := proposalAgent.GenerateQuote(r.Context(), req.Services, req.ClientInfo)
	if err != nil {
		http.Error(w, `{"error": "quote generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	quote.ID = uuid.New().String()

	response := map[string]interface{}{
		"quote":     quote,
		"generated": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateScopeHandler handles POST /api/v1/ai/proposals/scope
func GenerateScopeHandler(w http.ResponseWriter, r *http.Request) {
	proposalAgent := r.Context().Value(ContextKey("proposal_agent")).(*agents.ProposalAgent)

	var req struct {
		ProjectType  string `json:"project_type"`
		Requirements string `json:"requirements"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	scopeItems, err := proposalAgent.GenerateScopeOfWork(r.Context(), req.ProjectType, req.Requirements)
	if err != nil {
		http.Error(w, `{"error": "scope generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"scope_items": scopeItems,
		"count":       len(scopeItems),
		"generated":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GeneratePricingTiersHandler handles POST /api/v1/ai/proposals/pricing-tiers
func GeneratePricingTiersHandler(w http.ResponseWriter, r *http.Request) {
	proposalAgent := r.Context().Value(ContextKey("proposal_agent")).(*agents.ProposalAgent)

	var req struct {
		BaseScope string `json:"base_scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	pricing, err := proposalAgent.GenerateAlternativePricing(r.Context(), req.BaseScope)
	if err != nil {
		http.Error(w, `{"error": "pricing generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pricing_tiers": pricing,
		"generated":     time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
