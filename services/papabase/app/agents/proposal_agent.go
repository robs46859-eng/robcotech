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

// Proposal represents a complete business proposal
type Proposal struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	ExecutiveSummary string        `json:"executive_summary"`
	ScopeOfWork   []ScopeItem      `json:"scope_of_work"`
	Timeline      Timeline         `json:"timeline"`
	Pricing       PricingBreakdown `json:"pricing"`
	Terms         []string         `json:"terms"`
	NextSteps     []string         `json:"next_steps"`
	ValidUntil    string           `json:"valid_until"`
}

// ScopeItem represents a line item in the scope of work
type ScopeItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Deliverables []string `json:"deliverables"`
	EstimatedHours float64 `json:"estimated_hours"`
	Price       float64 `json:"price"`
}

// Timeline represents project timeline
type Timeline struct {
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
	TotalDays     int      `json:"total_days"`
	Milestones    []Milestone `json:"milestones"`
}

// Milestone represents a project milestone
type Milestone struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	PaymentPercent float64 `json:"payment_percent"`
}

// PricingBreakdown represents the pricing structure
type PricingBreakdown struct {
	Subtotal    float64 `json:"subtotal"`
	TaxRate     float64 `json:"tax_rate"`
	Tax         float64 `json:"tax"`
	Total       float64 `json:"total"`
	Currency    string  `json:"currency"`
	PaymentTerms string `json:"payment_terms"`
}

// ProposalAgent generates professional business proposals and quotes
type ProposalAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// NewProposalAgent creates a new proposal generation agent
func NewProposalAgent(gatewayURL string) *ProposalAgent {
	return &ProposalAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// GenerateProposal creates a complete proposal from lead requirements
func (pa *ProposalAgent) GenerateProposal(ctx context.Context, lead *Lead, requirements string) (*Proposal, error) {
	prompt := pa.buildProposalPrompt(lead, requirements)

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return pa.parseProposal(response)
}

// GenerateQuote creates a quick quote/estimate
func (pa *ProposalAgent) GenerateQuote(ctx context.Context, services []string, clientInfo map[string]string) (*Proposal, error) {
	prompt := pa.buildQuotePrompt(services, clientInfo)

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return pa.parseProposal(response)
}

// GenerateScopeOfWork creates detailed scope of work document
func (pa *ProposalAgent) GenerateScopeOfWork(ctx context.Context, projectType, requirements string) ([]ScopeItem, error) {
	prompt := fmt.Sprintf(`Create a detailed scope of work for:

Project Type: %s
Requirements: %s

Break down into 5-10 line items. For each item include:
- Clear title
- Detailed description
- Specific deliverables (list)
- Estimated hours
- Fair market price

Return JSON array of objects with keys: id, title, description, deliverables, estimated_hours, price`,
		projectType, requirements)

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var scopeItems []ScopeItem
	if err := json.Unmarshal([]byte(response), &scopeItems); err != nil {
		// Return single item if parsing fails
		item, err := pa.parseScopeItem(response)
		if err != nil {
			return nil, err
		}
		return []ScopeItem{*item}, nil
	}

	return scopeItems, nil
}

// EstimateProjectTimeline creates a project timeline
func (pa *ProposalAgent) EstimateProjectTimeline(ctx context.Context, scopeItems []ScopeItem) (*Timeline, error) {
	prompt := fmt.Sprintf(`Create a project timeline based on this scope:

Scope Items: %d items
Total Estimated Hours: %.1f

Create:
- Start date (next Monday)
- End date (based on hours + buffer)
- 3-5 milestones with payment percentages
- Total duration in days

Return JSON with keys: start_date, end_date, total_days, milestones`,
		len(scopeItems), sumHours(scopeItems))

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var timeline Timeline
	if err := json.Unmarshal([]byte(response), &timeline); err != nil {
		// Default timeline
		timeline = Timeline{
			StartDate:  time.Now().Format("2006-01-02"),
			EndDate:    time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			TotalDays:  30,
			Milestones: pa.getDefaultMilestones(),
		}
	}

	return &timeline, nil
}

// GeneratePricingBreakdown creates pricing with tax calculation
func (pa *ProposalAgent) GeneratePricingBreakdown(ctx context.Context, scopeItems []ScopeItem, taxRate float64, currency string) (*PricingBreakdown, error) {
	subtotal := 0.0
	for _, item := range scopeItems {
		subtotal += item.Price
	}

	tax := subtotal * (taxRate / 100)
	total := subtotal + tax

	return &PricingBreakdown{
		Subtotal:     subtotal,
		TaxRate:      taxRate,
		Tax:          tax,
		Total:        total,
		Currency:     currency,
		PaymentTerms: "50% upfront, 50% on completion",
	}, nil
}

// GenerateContractTerms creates standard contract terms
func (pa *ProposalAgent) GenerateContractTerms(ctx context.Context, projectType string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate standard contract terms for a %s project.

Include 8-12 terms covering:
- Payment terms
- Revision policy
- Timeline expectations
- Intellectual property
- Confidentiality
- Termination conditions
- Liability limitations
- Dispute resolution

Return JSON array of strings.`, projectType)

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return pa.getDefaultTerms(), nil
	}

	var terms []string
	if err := json.Unmarshal([]byte(response), &terms); err != nil {
		return pa.getDefaultTerms(), nil
	}

	return terms, nil
}

// GenerateNextSteps creates action items for proposal acceptance
func (pa *ProposalAgent) GenerateNextSteps(ctx context.Context, proposalType string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate next steps for accepting a %s proposal.

Include 4-6 clear action items for the client.

Return JSON array of strings.`, proposalType)

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return []string{
			"Review this proposal",
			"Sign and return the agreement",
			"Pay the initial deposit",
			"Schedule kickoff meeting",
		}, nil
	}

	var steps []string
	if err := json.Unmarshal([]byte(response), &steps); err != nil {
		return []string{response}, nil
	}

	return steps, nil
}

// GenerateProposalEmail creates a cover email for sending proposals
func (pa *ProposalAgent) GenerateProposalEmail(ctx context.Context, clientName, proposalTitle string) (string, error) {
	prompt := fmt.Sprintf(`Write a professional email to accompany a proposal:

Client: %s
Proposal: %s

Tone: Professional, enthusiastic, not pushy
Include:
- Personalized greeting
- Reference to previous discussions
- Proposal attached/linked
- Key highlights
- Call to action
- Availability for questions
- Professional sign-off

Return the email body only.`, clientName, proposalTitle)

	return pa.callGateway(ctx, prompt)
}

// GenerateAlternativePricing creates tiered pricing options
func (pa *ProposalAgent) GenerateAlternativePricing(ctx context.Context, baseScope string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Create tiered pricing options based on this scope:

Base Scope: %s

Generate 3 tiers:
1. Basic/Economy - stripped down version
2. Standard/Professional - full scope
3. Premium/Enterprise - with extras

For each tier include:
- Tier name
- What's included
- Price
- Best for (target customer)

Return JSON with keys: basic, standard, premium`, baseScope)

	response, err := pa.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var pricing map[string]interface{}
	if err := json.Unmarshal([]byte(response), &pricing); err != nil {
		return map[string]interface{}{"pricing_options": response}, nil
	}

	return pricing, nil
}

func (pa *ProposalAgent) buildProposalPrompt(lead *Lead, requirements string) string {
	return fmt.Sprintf(`Create a comprehensive business proposal.

Client Information:
- Name: %s
- Company: %s
- Email: %s
- Phone: %s

Project Requirements:
%s

Generate a complete proposal with:
1. Executive Summary (2-3 paragraphs)
2. Scope of Work (5-8 line items with descriptions, deliverables, hours, prices)
3. Timeline (start date, end date, milestones with payment percentages)
4. Pricing Breakdown (subtotal, tax, total in USD)
5. Terms and Conditions (8-10 standard terms)
6. Next Steps (4-6 action items)

Return as JSON with keys: title, executive_summary, scope_of_work, timeline, pricing, terms, next_steps, valid_until

Use realistic pricing for professional services.`,
		lead.Name, lead.Company, lead.Email, lead.Phone, requirements)
}

func (pa *ProposalAgent) buildQuotePrompt(services []string, clientInfo map[string]string) string {
	servicesStr := strings.Join(services, ", ")

	return fmt.Sprintf(`Create a quick quote for these services:

Services: %s
Client: %s
Company: %s

Generate:
1. Brief description for each service
2. Individual pricing
3. Total estimate
4. Valid until date (30 days from now)
5. Simple terms (3-5 items)

Return JSON with keys: title, scope_of_work, pricing, terms, valid_until`,
		servicesStr, clientInfo["name"], clientInfo["company"])
}

func (pa *ProposalAgent) parseProposal(response string) (*Proposal, error) {
	var proposal Proposal

	if err := json.Unmarshal([]byte(response), &proposal); err != nil {
		// Create minimal proposal from text
		proposal = Proposal{
			Title:         "Service Proposal",
			ExecutiveSummary: response[:min(500, len(response))],
			ScopeOfWork:   []ScopeItem{{Title: "Services", Description: response}},
			Terms:         pa.getDefaultTerms(),
			NextSteps:     []string{"Contact us to discuss", "Sign agreement", "Pay deposit"},
			ValidUntil:    time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		}
	}

	// Set defaults if missing
	if proposal.ValidUntil == "" {
		proposal.ValidUntil = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	}
	if proposal.Terms == nil || len(proposal.Terms) == 0 {
		proposal.Terms = pa.getDefaultTerms()
	}
	if proposal.NextSteps == nil || len(proposal.NextSteps) == 0 {
		proposal.NextSteps = []string{"Review proposal", "Sign agreement", "Pay deposit", "Schedule kickoff"}
	}

	return &proposal, nil
}

func (pa *ProposalAgent) parseScopeItem(response string) (*ScopeItem, error) {
	var item ScopeItem

	if err := json.Unmarshal([]byte(response), &item); err != nil {
		item = ScopeItem{
			Title:       "Service Item",
			Description: response,
			Price:       1000,
		}
	}

	return &item, nil
}

func (pa *ProposalAgent) getDefaultMilestones() []Milestone {
	return []Milestone{
		{Name: "Project Kickoff", Description: "Initial meeting and requirements gathering", DueDate: time.Now().AddDate(0, 0, 7).Format("2006-01-02"), PaymentPercent: 0},
		{Name: "Design Approval", Description: "Review and approve designs", DueDate: time.Now().AddDate(0, 0, 21).Format("2006-01-02"), PaymentPercent: 25},
		{Name: "Development Complete", Description: "All features implemented", DueDate: time.Now().AddDate(0, 0, 45).Format("2006-01-02"), PaymentPercent: 50},
		{Name: "Final Delivery", Description: "Project launch and handoff", DueDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"), PaymentPercent: 25},
	}
}

func (pa *ProposalAgent) getDefaultTerms() []string {
	return []string{
		"50% deposit required to begin work, 50% upon completion",
		"Proposal valid for 30 days from date of issue",
		"Additional revisions beyond scope billed at $150/hour",
		"Client to provide all necessary content and materials",
		"Project timeline dependent on timely client feedback",
		"Intellectual property transfers upon final payment",
		"Either party may terminate with 14 days written notice",
		"Confidentiality maintained for all project information",
	}
}

func (pa *ProposalAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a professional proposal writer. Create detailed, persuasive business proposals. Return responses in valid JSON format."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.5,
		"max_tokens":  3000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/ai", pa.gatewayURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "papabase-service-key")

	resp, err := pa.httpClient.Do(req)
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

func sumHours(items []ScopeItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.EstimatedHours
	}
	return total
}
