// Package arkham provides client integration with Arkham security service
package arkham

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client for Arkham security service
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// ThreatClassification is the response from Arkham classify endpoint
type ThreatClassification struct {
	RequestID         string                 `json:"request_id"`
	Classification    string                 `json:"classification"` // benign, probe, attack, scanner
	ThreatScore       float64                `json:"threat_score"`
	FingerprintHash   string                 `json:"fingerprint_hash,omitempty"`
	RecommendedAction string                 `json:"recommended_action"` // pass, deceive, block
	Metadata          map[string]interface{} `json:"metadata"`
}

// DeceptionResponse is the response from Arkham deceive endpoint
type DeceptionResponse struct {
	RequestID        string                 `json:"request_id"`
	TrapType         string                 `json:"trap_type"`
	DeceptionPayload map[string]interface{} `json:"deception_payload"`
	EngagementID     string                 `json:"engagement_id"`
}

// BlockResponse is the response from Arkham block endpoint
type BlockResponse struct {
	HTTPStatus int                    `json:"http_status"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
}

// NewClient creates a new Arkham client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ClassifyRequest sends a request to Arkham for threat classification
func (c *Client) ClassifyRequest(ctx context.Context, tenantID string, requestInfo map[string]interface{}) (*ThreatClassification, error) {
	url := fmt.Sprintf("%s/api/v1/classify?tenant_id=%s", c.baseURL, tenantID)
	
	payload, err := json.Marshal(requestInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arkham classify failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arkham returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var classification ThreatClassification
	if err := json.Unmarshal(body, &classification); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &classification, nil
}

// GenerateDeception requests a deception trap from Arkham
func (c *Client) GenerateDeception(ctx context.Context, tenantID, fingerprintHash string, requestInfo map[string]interface{}) (*DeceptionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/deceive?tenant_id=%s&fingerprint_hash=%s", c.baseURL, tenantID, fingerprintHash)
	
	payload, err := json.Marshal(requestInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arkham deceive failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arkham returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var deception DeceptionResponse
	if err := json.Unmarshal(body, &deception); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &deception, nil
}

// ApplyBlock requests a creative block from Arkham
func (c *Client) ApplyBlock(ctx context.Context, tenantID, fingerprintHash, engagementID string) (*BlockResponse, error) {
	url := fmt.Sprintf("%s/api/v1/block?tenant_id=%s&fingerprint_hash=%s", c.baseURL, tenantID, fingerprintHash)
	
	if engagementID != "" {
		url += "&engagement_id=" + engagementID
	}
	
	payload := map[string]interface{}{}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arkham block failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arkham returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var block BlockResponse
	if err := json.Unmarshal(body, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &block, nil
}
