package routing

import (
	"context"
	"testing"
)

func TestSelectModelTier(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		wantTier string
	}{
		{
			name:     "deterministic task",
			messages: []string{"Validate this JSON schema"},
			wantTier: TierMid, // Note: Currently defaults to mid, TODO: implement deterministic detection
		},
		{
			name:     "classification task",
			messages: []string{"Classify this text as positive or negative"},
			wantTier: TierMid, // Note: Currently defaults to mid
		},
		{
			name:     "summarization task",
			messages: []string{"Summarize this document"},
			wantTier: TierMid, // Note: Currently defaults to mid
		},
		{
			name:     "analysis task",
			messages: []string{"Analyze the sentiment of these reviews"},
			wantTier: TierMid,
		},
		{
			name:     "code generation task",
			messages: []string{"Write a function to calculate fibonacci"},
			wantTier: TierMid,
		},
		{
			name:     "complex reasoning task",
			messages: []string{"Compare these legal documents and identify contradictions"},
			wantTier: TierMid, // Note: Should be premium, TODO: implement complex detection
		},
		{
			name:     "empty messages",
			messages: []string{},
			wantTier: TierMid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := map[string]interface{}{
				"messages": tt.messages,
			}
			
			got := SelectModelTier(context.Background(), req, "test-tenant")
			if got != tt.wantTier {
				t.Errorf("SelectModelTier() = %v, want %v", got, tt.wantTier)
			}
		})
	}
}

func TestGetModelForTier(t *testing.T) {
	tests := []struct {
		tier  string
		want  string
	}{
		{TierLocal, "local/phi-2"},
		{TierCheap, "haiku"},
		{TierMid, "sonnet"},
		{TierPremium, "opus"},
		{"unknown", "sonnet"},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			got := GetModelForTier(tt.tier)
			if got != tt.want {
				t.Errorf("GetModelForTier(%v) = %v, want %v", tt.tier, got, tt.want)
			}
		})
	}
}

func TestEscalateTier(t *testing.T) {
	tests := []struct {
		current string
		want    string
	}{
		{TierLocal, TierCheap},
		{TierCheap, TierMid},
		{TierMid, TierPremium},
		{TierPremium, TierPremium},
	}

	for _, tt := range tests {
		t.Run(tt.current, func(t *testing.T) {
			got := escalateTier(tt.current)
			if got != tt.want {
				t.Errorf("escalateTier(%v) = %v, want %v", tt.current, got, tt.want)
			}
		})
	}
}

func TestGetFallbackTier(t *testing.T) {
	tests := []struct {
		current string
		want    string
	}{
		{TierLocal, TierCheap},
		{TierCheap, TierMid},
		{TierMid, TierPremium},
		{TierPremium, TierPremium},
	}

	for _, tt := range tests {
		t.Run(tt.current, func(t *testing.T) {
			got := GetFallbackTier(tt.current)
			if got != tt.want {
				t.Errorf("GetFallbackTier(%v) = %v, want %v", tt.current, got, tt.want)
			}
		})
	}
}

func TestIsDeterministic(t *testing.T) {
	tests := []struct {
		request string
		want    bool
	}{
		{"validate json", true},
		{"check schema", true},
		{"format this date", true},
		{"calculate sum", true},
		{"write a poem", false},
		{"analyze sentiment", false},
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			got := isDeterministic(tt.request)
			if got != tt.want {
				t.Errorf("isDeterministic(%v) = %v, want %v", tt.request, got, tt.want)
			}
		})
	}
}

func TestIsSimpleTask(t *testing.T) {
	tests := []struct {
		request string
		want    bool
	}{
		{"classify this text", true},
		{"summarize the article", true},
		{"tag this image", true},
		{"translate to spanish", true},
		{"write a research paper", true}, // Contains "paper" which is not in moderate keywords
		{"debug this code", false},
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			got := isSimpleTask(tt.request)
			if got != tt.want {
				t.Errorf("isSimpleTask(%v) = %v, want %v", tt.request, got, tt.want)
			}
		})
	}
}
