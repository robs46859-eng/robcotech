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

// DeploymentPlan represents an AI-generated deployment strategy
type DeploymentPlan struct {
	Platform       string            `json:"platform"`
	Steps          []string          `json:"steps"`
	ConfigFiles    map[string]string `json:"config_files,omitempty"`
	DockerCommands []string          `json:"docker_commands,omitempty"`
	WebhookSetup   string            `json:"webhook_setup,omitempty"`
	MobileChecks   []string          `json:"mobile_checks,omitempty"`
}

// AzureDeploymentBot is an autonomous agent for deploying the stack
type AzureDeploymentBot struct {
	gatewayURL string
	httpClient *http.Client
}

// NewAzureDeploymentBot creates a new deployment bot
func NewAzureDeploymentBot(gatewayURL string) *AzureDeploymentBot {
	return &AzureDeploymentBot{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// PlanAzureWebDeployment plans a web app deployment on Azure Container Apps and Static Web Apps
func (adb *AzureDeploymentBot) PlanAzureWebDeployment(ctx context.Context, repoContext string) (*DeploymentPlan, error) {
	prompt := fmt.Sprintf(`Plan an Azure deployment for the following repository context:

Context: %s

Skills to utilize:
1. Docker Container Apps (build, push, deploy)
2. Static Web Apps (frontend deployment)
3. Stripe Webhook configuration
4. PostgreSQL and Redis integration

Return JSON with keys: platform, steps, config_files, docker_commands, webhook_setup.`,
		repoContext)

	response, err := adb.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return adb.parseDeploymentPlan(response)
}

// PlanMobileDeployment plans mobile app deployment (iOS/Android)
func (adb *AzureDeploymentBot) PlanMobileDeployment(ctx context.Context, appType string) (*DeploymentPlan, error) {
	prompt := fmt.Sprintf(`Plan a mobile app deployment for:

App Type: %s (e.g., React Native, Flutter, Swift)

Skills to utilize:
1. App Store Connect / Google Play Console setup
2. Fastlane / CI/CD automation
3. OTA (Over The Air) updates configuration

Return JSON with keys: platform, steps, mobile_checks.`,
		appType)

	response, err := adb.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return adb.parseDeploymentPlan(response)
}

// ConfigureStripeWebhooks generates the setup instructions and code for Stripe Webhooks
func (adb *AzureDeploymentBot) ConfigureStripeWebhooks(ctx context.Context, environment string) (*DeploymentPlan, error) {
	prompt := fmt.Sprintf(`Provide the Stripe Webhook setup for environment: %s

Include:
1. CLI commands for local testing
2. Azure endpoint registration
3. Secret management strategy

Return JSON with keys: platform, webhook_setup, steps.`,
		environment)

	response, err := adb.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return adb.parseDeploymentPlan(response)
}

func (adb *AzureDeploymentBot) callGateway(ctx context.Context, prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":  "gemini-2.0-flash",
		"prompt": prompt,
		"system": "You are an expert Azure Deployment Bot. You specialize in Docker, Azure Container Apps, Static Web Apps, Stripe Webhooks, and Mobile App deployments. Always return valid JSON matching the requested structure.",
		"temperature": 0.2, // Low temperature for deployment tasks
	}
	
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", adb.gatewayURL+"/v1/ai", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer fsa_dev_key_12345")
	
	resp, err := adb.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gateway error: %s", string(respBody))
	}
	
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	return result.Response, nil
}

func (adb *AzureDeploymentBot) parseDeploymentPlan(response string) (*DeploymentPlan, error) {
	// Clean markdown formatting if present
	response = strings.TrimPrefix(response, "```json\n")
	response = strings.TrimSuffix(response, "\n```")
	response = strings.TrimSpace(response)
	
	var result DeploymentPlan
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse deployment plan: %w", err)
	}
	
	return &result, nil
}
