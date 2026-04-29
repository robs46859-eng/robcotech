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

// ContentGeneration represents AI-generated business content
type ContentGeneration struct {
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	HTML        string            `json:"html,omitempty"`
	MetaTags    map[string]string `json:"meta_tags,omitempty"`
	WordCount   int               `json:"word_count"`
	ReadingTime int               `json:"reading_time_minutes"`
	SEOScore    int               `json:"seo_score"`
}

// ContentAgent generates business copy, SEO content, and marketing materials
type ContentAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// NewContentAgent creates a new content generation agent
func NewContentAgent(gatewayURL string) *ContentAgent {
	return &ContentAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// GenerateBusinessDescription creates a compelling business description
func (ca *ContentAgent) GenerateBusinessDescription(ctx context.Context, businessName, industry, keyServices string) (*ContentGeneration, error) {
	prompt := fmt.Sprintf(`Write a compelling business description for:

Business Name: %s
Industry: %s
Key Services: %s

Requirements:
- 2-3 paragraphs
- Professional yet approachable tone
- Highlight unique value proposition
- Include relevant keywords for SEO
- Call-to-action at the end

Return JSON with keys: title, content, word_count, seo_score`,
		businessName, industry, keyServices)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ca.parseContentGeneration(response)
}

// GenerateServiceDescription creates detailed service descriptions
func (ca *ContentAgent) GenerateServiceDescription(ctx context.Context, serviceName, targetAudience, benefits string) (*ContentGeneration, error) {
	prompt := fmt.Sprintf(`Write a detailed service description:

Service: %s
Target Audience: %s
Key Benefits: %s

Requirements:
- Clear headline
- Problem/solution framework
- Feature list with benefits
- Social proof section placeholder
- Strong call-to-action
- 300-500 words

Return JSON with keys: title, content, html, word_count`,
		serviceName, targetAudience, benefits)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ca.parseContentGeneration(response)
}

// GenerateSEOMetaTags creates SEO-optimized meta tags
func (ca *ContentAgent) GenerateSEOMetaTags(ctx context.Context, pageContent, keywords string) (map[string]string, error) {
	prompt := fmt.Sprintf(`Generate SEO meta tags for this content:

Page Content: %s

Target Keywords: %s

Create:
- Meta title (50-60 characters)
- Meta description (150-160 characters)
- Open Graph title
- Open Graph description
- 5-8 meta keywords

Return JSON with keys: meta_title, meta_description, og_title, og_description, keywords`,
		pageContent[:min(1000, len(pageContent))], keywords)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var metaTags map[string]string
	if err := json.Unmarshal([]byte(response), &metaTags); err != nil {
		return map[string]string{"error": "Failed to parse meta tags"}, nil
	}

	return metaTags, nil
}

// GenerateBlogPost creates a full blog post
func (ca *ContentAgent) GenerateBlogPost(ctx context.Context, topic, outline, tone string) (*ContentGeneration, error) {
	prompt := fmt.Sprintf(`Write a comprehensive blog post:

Topic: %s
Outline: %s
Tone: %s

Requirements:
- Engaging headline
- Introduction with hook
- Well-structured sections with H2/H3 headings
- Examples and actionable advice
- Conclusion with key takeaways
- 800-1500 words
- Include HTML formatting

Return JSON with keys: title, content, html, word_count, reading_time_minutes`,
		topic, outline, tone)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ca.parseContentGeneration(response)
}

// GenerateEmailCampaign creates email marketing content
func (ca *ContentAgent) GenerateEmailCampaign(ctx context.Context, campaignType, audience, goal string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Create an email campaign:

Type: %s
Audience: %s
Goal: %s

Generate:
1. Subject line options (3 variations)
2. Email body with personalization
3. Call-to-action
4. Follow-up email draft

Return JSON with keys: subject_lines, email_body, cta, follow_up_email`,
		campaignType, audience, goal)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var campaign map[string]interface{}
	if err := json.Unmarshal([]byte(response), &campaign); err != nil {
		return map[string]interface{}{"email_content": response}, nil
	}

	return campaign, nil
}

// GenerateSocialMediaPosts creates social media content
func (ca *ContentAgent) GenerateSocialMediaPosts(ctx context.Context, topic, platforms string) ([]map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Create social media posts for:

Topic: %s
Platforms: %s

For each platform, create:
- Platform-appropriate length
- Relevant hashtags
- Engagement-focused copy
- Emoji usage where appropriate

Return JSON array of objects with keys: platform, content, hashtags, character_count`,
		topic, platforms)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var posts []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &posts); err != nil {
		return []map[string]interface{}{{"content": response}}, nil
	}

	return posts, nil
}

// GenerateAboutUs creates an About Us page
func (ca *ContentAgent) GenerateAboutUs(ctx context.Context, companyInfo, mission, teamInfo string) (*ContentGeneration, error) {
	prompt := fmt.Sprintf(`Write an About Us page:

Company Info: %s
Mission: %s
Team Info: %s

Requirements:
- Compelling origin story
- Mission and values
- Team highlights
- Company culture
- Trust signals
- Professional yet personable tone
- 400-600 words

Return JSON with keys: title, content, html, word_count`,
		companyInfo, mission, teamInfo)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ca.parseContentGeneration(response)
}

// GenerateFAQ creates frequently asked questions
func (ca *ContentAgent) GenerateFAQ(ctx context.Context, businessType, commonQuestions string) ([]map[string]string, error) {
	prompt := fmt.Sprintf(`Generate FAQ content:

Business Type: %s
Common Questions/Topics: %s

Create 8-12 Q&A pairs that:
- Address common customer concerns
- Showcase expertise
- Include pricing/process questions
- End with contact CTA

Return JSON array of objects with keys: question, answer, category`,
		businessType, commonQuestions)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var faqs []map[string]string
	if err := json.Unmarshal([]byte(response), &faqs); err != nil {
		return []map[string]string{{"question": "FAQ generation error", "answer": response}}, nil
	}

	return faqs, nil
}

// GenerateTestimonialRequest creates testimonial request emails
func (ca *ContentAgent) GenerateTestimonialRequest(ctx context.Context, clientName, projectType string) (string, error) {
	prompt := fmt.Sprintf(`Write a testimonial request email:

Client: %s
Project Type: %s

Tone: Friendly, professional, not pushy
Include:
- Appreciation for working together
- Why their feedback matters
- Easy way to provide testimonial
- Optional: incentive mention

Return the email body only.`,
		clientName, projectType)

	return ca.callGateway(ctx, prompt)
}

// RewriteContent improves existing content
func (ca *ContentAgent) RewriteContent(ctx context.Context, originalContent, improvementGoal string) (*ContentGeneration, error) {
	prompt := fmt.Sprintf(`Rewrite and improve this content:

Goal: %s

Original Content:
%s

Requirements:
- Maintain core message
- Improve clarity and flow
- Enhance persuasiveness
- Fix any issues
- Return improved version

Return JSON with keys: title, content, word_count`,
		improvementGoal, originalContent)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ca.parseContentGeneration(response)
}

// GenerateLandingPageCopy creates landing page content
func (ca *ContentAgent) GenerateLandingPageCopy(ctx context.Context, product, targetAudience, uniqueValue string) (*ContentGeneration, error) {
	prompt := fmt.Sprintf(`Write landing page copy:

Product/Service: %s
Target Audience: %s
Unique Value Proposition: %s

Include:
- Attention-grabbing headline
- Subheadline with benefit
- Problem/agitation section
- Solution presentation
- Feature/benefit bullets
- Social proof section
- Strong CTA
- FAQ section

Return JSON with keys: title, content, html, word_count`,
		product, targetAudience, uniqueValue)

	response, err := ca.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ca.parseContentGeneration(response)
}

func (ca *ContentAgent) parseContentGeneration(response string) (*ContentGeneration, error) {
	var gen ContentGeneration

	if err := json.Unmarshal([]byte(response), &gen); err != nil {
		// Return as plain content
		gen.Content = response
		gen.WordCount = len(strings.Fields(response))
		gen.ReadingTime = gen.WordCount / 200
		if gen.ReadingTime < 1 {
			gen.ReadingTime = 1
		}
	}

	// Calculate reading time if not provided
	if gen.ReadingTime == 0 && gen.WordCount > 0 {
		gen.ReadingTime = gen.WordCount / 200
		if gen.ReadingTime < 1 {
			gen.ReadingTime = 1
		}
	}

	return &gen, nil
}

func (ca *ContentAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a professional copywriter and content strategist. Return responses in valid JSON format when requested."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/ai", ca.gatewayURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "papabase-service-key")

	resp, err := ca.httpClient.Do(req)
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
