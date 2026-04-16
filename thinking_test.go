package acosmi

import (
	"testing"
)

// ============================================================================
// NewThinkingConfig
// ============================================================================

func TestNewThinkingConfig(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		wantType string
		wantLvl  string
	}{
		{"empty string → disabled", "", "disabled", ""},
		{"off → disabled", ThinkingOff, "disabled", ""},
		{"high → adaptive+high", ThinkingHigh, "adaptive", ThinkingHigh},
		{"max → adaptive+max", ThinkingMax, "adaptive", ThinkingMax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewThinkingConfig(tt.level)
			if cfg.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", cfg.Type, tt.wantType)
			}
			if cfg.Level != tt.wantLvl {
				t.Errorf("Level = %q, want %q", cfg.Level, tt.wantLvl)
			}
		})
	}
}

// ============================================================================
// resolveThinkingLevel
// ============================================================================

// helper: build body map with initial max_tokens, run resolveThinkingLevel
func runResolve(level string, inputMaxTokens int, caps ModelCapabilities) map[string]any {
	body := make(map[string]any)
	if inputMaxTokens > 0 {
		body["max_tokens"] = inputMaxTokens
	}
	// simulate temperature being set (to verify it gets deleted)
	body["temperature"] = 0.7
	req := &ChatRequest{
		MaxTokens: inputMaxTokens,
		Thinking:  &ThinkingConfig{Type: "adaptive", Level: level},
	}
	resolveThinkingLevel(body, req, caps)
	return body
}

var capsAdaptive = ModelCapabilities{
	SupportsThinking:         true,
	SupportsAdaptiveThinking: true,
	SupportsEffort:           true,
	SupportsMaxEffort:        true,
	MaxOutputTokens:          128000,
}

var capsOldModel = ModelCapabilities{
	SupportsThinking: true,
	// 不支持 adaptive、effort、maxEffort
}

func TestResolveThinkingLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		inputMax      int
		caps          ModelCapabilities
		wantThinkType string
		wantEffort    string // "" = should not exist
		wantMaxTokens int    // 0 = should not exist (off case)
	}{
		// ── off ──
		{"off", ThinkingOff, 8192, capsAdaptive, "disabled", "", 0},

		// ── high + adaptive model ──
		{"high-adaptive-8k", ThinkingHigh, 8192, capsAdaptive,
			"adaptive", "high", 32000},
		{"high-adaptive-64k", ThinkingHigh, 64000, capsAdaptive,
			"adaptive", "high", 64000}, // already >= 32K, no change
		{"high-adaptive-zero", ThinkingHigh, 0, capsAdaptive,
			"adaptive", "high", 32000}, // fallback to 32K

		// ── max + adaptive model ──
		{"max-adaptive-8k", ThinkingMax, 8192, capsAdaptive,
			"adaptive", "max", 128000},
		{"max-adaptive-128k", ThinkingMax, 128000, capsAdaptive,
			"adaptive", "max", 128000},

		// ── high + old model (no adaptive, no effort) ──
		{"high-old-8k", ThinkingHigh, 8192, capsOldModel,
			"enabled", "", 32000},

		// ── max + old model ──
		{"max-old-8k", ThinkingMax, 8192, capsOldModel,
			"enabled", "", 128000}, // fallback MaxOutputTokens=0 → 128K
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := runResolve(tt.level, tt.inputMax, tt.caps)

			// ── thinking type ──
			thinking, ok := body["thinking"].(map[string]any)
			if !ok {
				t.Fatalf("thinking should be map, got %T: %v", body["thinking"], body["thinking"])
			}
			if thinking["type"] != tt.wantThinkType {
				t.Errorf("thinking.type = %v, want %q", thinking["type"], tt.wantThinkType)
			}

			// ── effort ──
			if tt.wantEffort == "" {
				if _, has := body["effort"]; has {
					t.Errorf("effort should not exist, got %v", body["effort"])
				}
			} else {
				effort, ok := body["effort"].(map[string]any)
				if !ok {
					t.Fatalf("effort should be map, got %T", body["effort"])
				}
				if effort["level"] != tt.wantEffort {
					t.Errorf("effort.level = %v, want %q", effort["level"], tt.wantEffort)
				}
			}

			// ── max_tokens ──
			if tt.wantMaxTokens == 0 {
				// off: max_tokens should not be set by resolveThinkingLevel
				// (it may exist from our initial setup, but for "off" we return early)
			} else {
				gotMax, ok := body["max_tokens"].(int)
				if !ok {
					t.Fatalf("max_tokens should be int, got %T: %v", body["max_tokens"], body["max_tokens"])
				}
				if gotMax != tt.wantMaxTokens {
					t.Errorf("max_tokens = %d, want %d", gotMax, tt.wantMaxTokens)
				}
			}

			// ── temperature removed (except off) ──
			if tt.level != ThinkingOff {
				if _, has := body["temperature"]; has {
					t.Error("temperature should be deleted when thinking is active")
				}
			}
		})
	}
}

// ============================================================================
// Old model fallback: budget_tokens = maxTokens - 1
// ============================================================================

func TestResolveThinkingLevel_OldModelBudget(t *testing.T) {
	t.Run("high-old-budget", func(t *testing.T) {
		body := runResolve(ThinkingHigh, 8192, capsOldModel)
		thinking := body["thinking"].(map[string]any)
		budget, ok := thinking["budget_tokens"].(int)
		if !ok {
			t.Fatal("budget_tokens should be int")
		}
		// maxTokens was lifted to 32000; budget = 32000 - 1 = 31999
		if budget != 31999 {
			t.Errorf("budget = %d, want 31999", budget)
		}
	})

	t.Run("max-old-budget", func(t *testing.T) {
		body := runResolve(ThinkingMax, 8192, capsOldModel)
		thinking := body["thinking"].(map[string]any)
		budget := thinking["budget_tokens"].(int)
		// maxTokens was lifted to 128000 (fallback); budget = 128000 - 1 = 127999
		if budget != 127999 {
			t.Errorf("budget = %d, want 127999", budget)
		}
	})
}

// ============================================================================
// Display propagation
// ============================================================================

func TestResolveThinkingLevel_Display(t *testing.T) {
	body := make(map[string]any)
	req := &ChatRequest{
		MaxTokens: 32000,
		Thinking:  &ThinkingConfig{Type: "adaptive", Level: ThinkingHigh, Display: "summary"},
	}
	resolveThinkingLevel(body, req, capsAdaptive)
	thinking := body["thinking"].(map[string]any)
	if thinking["display"] != "summary" {
		t.Errorf("display = %v, want 'summary'", thinking["display"])
	}
}

// ============================================================================
// Off does NOT set max_tokens or effort
// ============================================================================

func TestResolveThinkingLevel_OffMinimalSideEffects(t *testing.T) {
	body := make(map[string]any)
	body["max_tokens"] = 8192
	body["temperature"] = 0.7
	req := &ChatRequest{
		MaxTokens: 8192,
		Thinking:  &ThinkingConfig{Type: "adaptive", Level: ThinkingOff},
	}
	resolveThinkingLevel(body, req, capsAdaptive)

	// off: thinking = disabled
	thinking := body["thinking"].(map[string]any)
	if thinking["type"] != "disabled" {
		t.Errorf("thinking.type = %v, want 'disabled'", thinking["type"])
	}

	// off should NOT touch max_tokens (still the original 8192)
	if body["max_tokens"] != 8192 {
		t.Errorf("max_tokens should remain 8192 for off, got %v", body["max_tokens"])
	}

	// off should NOT set effort
	if _, has := body["effort"]; has {
		t.Error("effort should not be set for off")
	}

	// off should NOT delete temperature
	if _, has := body["temperature"]; !has {
		t.Error("temperature should remain for off")
	}
}

// ============================================================================
// AnthropicAdapter.BuildRequestBody integration
// ============================================================================

func TestAnthropicAdapter_ThinkingLevelIntegration(t *testing.T) {
	adapter := &AnthropicAdapter{}

	t.Run("level=high removes temperature, sets effort", func(t *testing.T) {
		temp := 0.7
		req := &ChatRequest{
			Messages:    []ChatMessage{{Role: "user", Content: "hello"}},
			MaxTokens:   8192,
			Temperature: &temp,
			Thinking:    NewThinkingConfig(ThinkingHigh),
		}
		body, err := adapter.BuildRequestBody(capsAdaptive, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// temperature removed
		if _, has := body["temperature"]; has {
			t.Error("temperature should be removed when thinking is active")
		}

		// thinking = adaptive
		thinking := body["thinking"].(map[string]any)
		if thinking["type"] != "adaptive" {
			t.Errorf("thinking.type = %v, want 'adaptive'", thinking["type"])
		}

		// effort = high
		effort := body["effort"].(map[string]any)
		if effort["level"] != "high" {
			t.Errorf("effort.level = %v, want 'high'", effort["level"])
		}

		// max_tokens ≥ 32K
		if body["max_tokens"].(int) < ThinkingHighMinMaxTokens {
			t.Errorf("max_tokens = %v, want >= %d", body["max_tokens"], ThinkingHighMinMaxTokens)
		}
	})

	t.Run("level=off keeps temperature, no effort", func(t *testing.T) {
		temp := 0.7
		req := &ChatRequest{
			Messages:    []ChatMessage{{Role: "user", Content: "hello"}},
			MaxTokens:   8192,
			Temperature: &temp,
			Thinking:    NewThinkingConfig(ThinkingOff),
		}
		body, err := adapter.BuildRequestBody(capsAdaptive, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, has := body["temperature"]; !has {
			t.Error("temperature should remain for off")
		}
		if _, has := body["effort"]; has {
			t.Error("effort should not exist for off")
		}
	})

	t.Run("nil thinking → v0.8.0 passthrough", func(t *testing.T) {
		req := &ChatRequest{
			Messages:  []ChatMessage{{Role: "user", Content: "hello"}},
			MaxTokens: 8192,
			Effort:    &EffortConfig{Level: "medium"},
		}
		body, err := adapter.BuildRequestBody(capsAdaptive, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// no thinking key
		if _, has := body["thinking"]; has {
			t.Error("thinking should be absent when nil")
		}

		// effort passthrough
		if body["effort"] == nil {
			t.Error("effort should be passed through in compat mode")
		}
	})

	t.Run("level=max caps effort=max, maxTokens=modelMax", func(t *testing.T) {
		req := &ChatRequest{
			Messages:  []ChatMessage{{Role: "user", Content: "hello"}},
			MaxTokens: 8192,
			Thinking:  NewThinkingConfig(ThinkingMax),
		}
		body, err := adapter.BuildRequestBody(capsAdaptive, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		effort := body["effort"].(map[string]any)
		if effort["level"] != "max" {
			t.Errorf("effort.level = %v, want 'max'", effort["level"])
		}
		if body["max_tokens"].(int) != 128000 {
			t.Errorf("max_tokens = %v, want 128000", body["max_tokens"])
		}
	})
}

// ============================================================================
// Compat mode: Level="" → v0.8.0 passthrough
// ============================================================================

func TestAnthropicAdapter_CompatMode(t *testing.T) {
	adapter := &AnthropicAdapter{}

	// Explicit Type+BudgetTokens without Level → passthrough
	req := &ChatRequest{
		Messages:  []ChatMessage{{Role: "user", Content: "hello"}},
		MaxTokens: 8192,
		Thinking:  &ThinkingConfig{Type: "enabled", BudgetTokens: 4000},
		Effort:    &EffortConfig{Level: "high"},
	}
	body, err := adapter.BuildRequestBody(capsAdaptive, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// thinking is the raw struct (passthrough)
	thinking, ok := body["thinking"].(*ThinkingConfig)
	if !ok {
		t.Fatalf("compat mode should passthrough *ThinkingConfig, got %T", body["thinking"])
	}
	if thinking.BudgetTokens != 4000 {
		t.Errorf("BudgetTokens = %d, want 4000", thinking.BudgetTokens)
	}

	// effort is the raw struct (passthrough)
	effort, ok := body["effort"].(*EffortConfig)
	if !ok {
		t.Fatalf("compat mode should passthrough *EffortConfig, got %T", body["effort"])
	}
	if effort.Level != "high" {
		t.Errorf("effort.Level = %q, want 'high'", effort.Level)
	}

	// max_tokens unchanged
	if body["max_tokens"] != 8192 {
		t.Errorf("max_tokens = %v, want 8192 (unchanged)", body["max_tokens"])
	}
}

// ============================================================================
// betaEffort injection for Level mode
// ============================================================================

func TestBetaEffort_LevelMode(t *testing.T) {
	adapter := &AnthropicAdapter{}

	t.Run("level=high injects betaEffort", func(t *testing.T) {
		req := &ChatRequest{
			Messages:  []ChatMessage{{Role: "user", Content: "hello"}},
			MaxTokens: 8192,
			Thinking:  NewThinkingConfig(ThinkingHigh),
		}
		body, err := adapter.BuildRequestBody(capsAdaptive, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		betas, _ := body["betas"].([]string)
		found := false
		for _, b := range betas {
			if b == "effort-2025-11-24" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("betaEffort not found in betas: %v", betas)
		}
	})

	t.Run("level=off does NOT inject betaEffort", func(t *testing.T) {
		req := &ChatRequest{
			Messages:  []ChatMessage{{Role: "user", Content: "hello"}},
			MaxTokens: 8192,
			Thinking:  NewThinkingConfig(ThinkingOff),
		}
		body, err := adapter.BuildRequestBody(capsAdaptive, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		betas, _ := body["betas"].([]string)
		for _, b := range betas {
			if b == "effort-2025-11-24" {
				t.Error("betaEffort should NOT be injected for level=off")
				break
			}
		}
	})
}

// ============================================================================
// No thinking capability → resolveThinkingLevel is a no-op
// ============================================================================

func TestResolveThinkingLevel_NoCapability(t *testing.T) {
	capsNone := ModelCapabilities{} // no thinking, no adaptive, no effort

	body := make(map[string]any)
	body["max_tokens"] = 4096
	body["temperature"] = 0.5
	req := &ChatRequest{
		MaxTokens: 4096,
		Thinking:  &ThinkingConfig{Type: "adaptive", Level: ThinkingHigh},
	}
	resolveThinkingLevel(body, req, capsNone)

	// max_tokens should NOT be lifted
	if body["max_tokens"] != 4096 {
		t.Errorf("max_tokens = %v, want 4096 (unchanged for no-cap model)", body["max_tokens"])
	}

	// thinking should NOT be set
	if _, has := body["thinking"]; has {
		t.Error("thinking should not be set for no-cap model")
	}

	// effort should NOT be set
	if _, has := body["effort"]; has {
		t.Error("effort should not be set for no-cap model")
	}

	// temperature should NOT be deleted
	if _, has := body["temperature"]; !has {
		t.Error("temperature should remain for no-cap model")
	}
}
