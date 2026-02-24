package adapter

import "testing"

func TestNew_ValidProviders(t *testing.T) {
	tests := []struct {
		provider string
	}{
		{ProviderClaude},
		{ProviderOpenAI},
		{ProviderGemini},
		{ProviderOllama},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			a, err := New(tt.provider, "", "test-key", "")
			if err != nil {
				t.Fatalf("New(%q) error: %v", tt.provider, err)
			}
			if a == nil {
				t.Fatalf("New(%q) returned nil adapter", tt.provider)
			}
			info := a.Info()
			if info.Provider != tt.provider {
				t.Errorf("Info().Provider = %q, want %q", info.Provider, tt.provider)
			}
		})
	}
}

func TestNew_InvalidProvider(t *testing.T) {
	_, err := New("invalid", "", "key", "")
	if err == nil {
		t.Error("expected error for invalid provider")
	}
}

func TestNew_OllamaDefaults(t *testing.T) {
	a, err := New(ProviderOllama, "", "", "")
	if err != nil {
		t.Fatalf("New(ollama) error: %v", err)
	}
	// Should use default host and model.
	info := a.Info()
	if info.Provider != ProviderOllama {
		t.Errorf("provider: got %q", info.Provider)
	}
}

func TestProviderConstants(t *testing.T) {
	if ProviderClaude != "claude" {
		t.Errorf("ProviderClaude = %q", ProviderClaude)
	}
	if ProviderOpenAI != "openai" {
		t.Errorf("ProviderOpenAI = %q", ProviderOpenAI)
	}
	if ProviderGemini != "gemini" {
		t.Errorf("ProviderGemini = %q", ProviderGemini)
	}
	if ProviderOllama != "ollama" {
		t.Errorf("ProviderOllama = %q", ProviderOllama)
	}
}

func TestCompletionRequest_Defaults(t *testing.T) {
	req := CompletionRequest{}
	if req.MaxTokens != 0 {
		t.Error("expected zero MaxTokens by default")
	}
	if req.Stream {
		t.Error("expected Stream false by default")
	}
}
