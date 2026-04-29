package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
)

// ============================================================================
// Dad AI Website Generation Endpoints
// ============================================================================

// GenerateWebsiteRequest is the request body for website generation
type GenerateWebsiteRequest struct {
	Prompt       string   `json:"prompt"`
	BusinessType string   `json:"business_type"`
	OutputType   string   `json:"output_type"` // single_page, multi_page, dashboard
	Tier         string   `json:"tier"`        // starter, studio, agency
	ColorScheme  string   `json:"color_scheme"`
	Features     []string `json:"features"`
}

// GenerateWebsiteHandler handles POST /api/v1/ai/generate
func GenerateWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	dadAI := r.Context().Value(DadAIKey).(*agents.DadAIAgent)

	var req GenerateWebsiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate tier
	validTiers := map[string]bool{"starter": true, "studio": true, "agency": true}
	if !validTiers[req.Tier] {
		req.Tier = "starter" // Default to starter
	}

	// Generate the website
	genReq := agents.GenerationRequest{
		Prompt:         req.Prompt,
		BusinessType:   req.BusinessType,
		OutputType:     req.OutputType,
		Tier:           req.Tier,
		ColorScheme:    req.ColorScheme,
		Features:       req.Features,
	}

	response, err := dadAI.GenerateWebsite(r.Context(), genReq)
	if err != nil {
		http.Error(w, `{"error": "generation failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// WebsiteProject represents a stored website project
type WebsiteProject struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tier        string            `json:"tier"`
	Status      string            `json:"status"`
	Prompt      string            `json:"prompt"`
	OutputType  string            `json:"output_type"`
	HTML        string            `json:"html,omitempty"`
	CSS         string            `json:"css,omitempty"`
	JavaScript  string            `json:"javascript,omitempty"`
	ReactCode   string            `json:"react_code,omitempty"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// In-memory project store (replace with PostgreSQL in production)
var (
	projectStore = make(map[string]*WebsiteProject)
	projectMu    sync.RWMutex
)

// ListTemplatesHandler handles GET /api/v1/ai/templates
func ListTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	templates := []map[string]interface{}{
		{
			"id":          "landing-page",
			"name":        "Landing Page",
			"description": "Single page with hero, features, CTA",
			"tier":        "starter",
			"output_type": "single_page",
		},
		{
			"id":          "business-site",
			"name":        "Business Site",
			"description": "Multi-page with home, about, services, contact",
			"tier":        "studio",
			"output_type": "multi_page",
		},
		{
			"id":          "portfolio",
			"name":        "Portfolio",
			"description": "Gallery-focused site with project showcase",
			"tier":        "studio",
			"output_type": "multi_page",
		},
		{
			"id":          "saas-dashboard",
			"name":        "SaaS Dashboard",
			"description": "Full web app with auth, dashboard, analytics",
			"tier":        "agency",
			"output_type": "dashboard",
		},
		{
			"id":          "ecommerce",
			"name":        "E-commerce",
			"description": "Online store with products, cart, checkout",
			"tier":        "agency",
			"output_type": "dashboard",
		},
	}

	response := map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProjectHandler handles GET /api/v1/ai/projects/{id}
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	projectMu.RLock()
	project, ok := projectStore[projectID]
	projectMu.RUnlock()

	if !ok {
		http.Error(w, `{"error": "project not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// ListProjectsHandler handles GET /api/v1/ai/projects
func ListProjectsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")

	projectMu.RLock()
	var projects []*WebsiteProject
	for _, p := range projectStore {
		if p.TenantID == tenantID || tenantID == "" {
			projects = append(projects, p)
		}
	}
	projectMu.RUnlock()

	response := map[string]interface{}{
		"projects": projects,
		"total":    len(projects),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// Pricing and Billing Endpoints
// ============================================================================

// PricingPlan represents a subscription tier
type PricingPlan struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	PriceMonthly    float64  `json:"price_monthly"`
	PriceYearly     float64  `json:"price_yearly"`
	Seats           int      `json:"seats"`
	AIMonthly       int      `json:"ai_generations_monthly"`
	StorageGB       int      `json:"storage_gb"`
	Features        []string `json:"features"`
	OutputTypes     []string `json:"output_types"`
	SupportLevel    string   `json:"support_level"`
	CustomDomain    bool     `json:"custom_domain"`
	WhiteLabel      bool     `json:"white_label"`
	APIAccess       bool     `json:"api_access"`
	Slug            bool     `json:"sla"`
}

// ListPlansHandler handles GET /api/v1/pricing/plans
func ListPlansHandler(w http.ResponseWriter, r *http.Request) {
	plans := []PricingPlan{
		{
			ID:           "starter",
			Name:         "Starter",
			Description:  "For solo operators and side hustles",
			PriceMonthly: 29,
			PriceYearly:  290,
			Seats:        1,
			AIMonthly:    3,
			StorageGB:    1,
			Features: []string{
				"Single-page HTML websites",
				"Basic CRM (leads & tasks)",
				"Responsive design",
				"Contact forms",
				"SEO optimized",
			},
			OutputTypes:  []string{"single_page"},
			SupportLevel: "Community",
		},
		{
			ID:           "studio",
			Name:         "Studio",
			Description:  "For small studios (2-5 people)",
			PriceMonthly: 99,
			PriceYearly:  990,
			Seats:        5,
			AIMonthly:    15,
			StorageGB:    10,
			Features: []string{
				"Multi-page React websites",
				"Full CRM with workflows",
				"Email automation",
				"Custom domain",
				"Analytics dashboard",
				"Priority support",
				"Blog integration",
				"Booking system",
			},
			OutputTypes:  []string{"single_page", "multi_page"},
			SupportLevel: "Priority",
			CustomDomain: true,
		},
		{
			ID:           "agency",
			Name:         "Agency",
			Description:  "For growing agencies (6-20 people)",
			PriceMonthly: 299,
			PriceYearly:  2990,
			Seats:        20,
			AIMonthly:    -1, // Unlimited
			StorageGB:    100,
			Features: []string{
				"Full web applications",
				"Multi-user dashboards",
				"Client portals",
				"Payment integration",
				"API access",
				"White-label option",
				"Custom integrations",
				"SLA guarantee",
				"Dedicated support",
			},
			OutputTypes:  []string{"single_page", "multi_page", "dashboard"},
			SupportLevel: "Dedicated",
			CustomDomain: true,
			WhiteLabel:   true,
			APIAccess:    true,
			Slug:         true,
		},
		{
			ID:           "enterprise",
			Name:         "Enterprise",
			Description:  "For large teams with custom needs",
			PriceMonthly: 0, // Custom pricing
			PriceYearly:  0,
			Seats:        -1, // Unlimited
			AIMonthly:    -1,
			StorageGB:    -1,
			Features: []string{
				"Everything in Agency",
				"Unlimited seats",
				"Dedicated infrastructure",
				"SSO/SAML",
				"On-premise option",
				"Custom development",
				"24/7 phone support",
			},
			OutputTypes:  []string{"single_page", "multi_page", "dashboard"},
			SupportLevel: "24/7",
			CustomDomain: true,
			WhiteLabel:   true,
			APIAccess:    true,
			Slug:         true,
		},
	}

	response := map[string]interface{}{
		"plans": plans,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UsageRecord represents usage data
type UsageRecord struct {
	TenantID       string `json:"tenant_id"`
	TokensUsed     int64  `json:"tokens_used"`
	RequestsCount  int64  `json:"requests_count"`
	Generations    int64  `json:"generations"`
	StorageUsedMB  int64  `json:"storage_used_mb"`
	CurrentPeriod  string `json:"current_period"`
	PeriodStart    string `json:"period_start"`
	PeriodEnd      string `json:"period_end"`
}

// GetUsageHandler handles GET /api/v1/billing/usage
func GetUsageHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0)

	usage := UsageRecord{
		TenantID:      tenantID,
		TokensUsed:    0,
		RequestsCount: 0,
		Generations:   0,
		StorageUsedMB: 0,
		CurrentPeriod: now.Format("2006-01"),
		PeriodStart:   periodStart.Format("2006-01-02"),
		PeriodEnd:     periodEnd.Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// Invoice represents a billing invoice
type Invoice struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PeriodStart   string    `json:"period_start"`
	PeriodEnd     string    `json:"period_end"`
	LineItems     []string  `json:"line_items"`
	CreatedAt     time.Time `json:"created_at"`
	DueAt         time.Time `json:"due_at"`
	StripeInvoice string    `json:"stripe_invoice_id,omitempty"`
}

// ListInvoicesHandler handles GET /api/v1/billing/invoices
func ListInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")

	invoices := []*Invoice{
		{
			ID:          uuid.New().String(),
			TenantID:    tenantID,
			Amount:      99.00,
			Currency:    "USD",
			Status:      "paid",
			PeriodStart: time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
			PeriodEnd:   time.Now().Format("2006-01-02"),
			LineItems:   []string{"Studio Plan - Monthly"},
			CreatedAt:   time.Now().AddDate(0, -1, 0),
			DueAt:       time.Now(),
		},
	}

	response := map[string]interface{}{
		"invoices": invoices,
		"total":    len(invoices),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
