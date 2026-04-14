package core

import "testing"

func TestIsEULanguage(t *testing.T) {
	tests := []struct {
		lang string
		want bool
	}{
		{"en", true},
		{"cs", true},
		{"fr", true},
		{"ja", false},
		{"zh", false},
		{"ar", false},
		{"", false},
		{"EN", true},   // case-insensitive
		{" en ", true}, // trimmed
	}
	for _, tt := range tests {
		if got := IsEULanguage(tt.lang); got != tt.want {
			t.Errorf("IsEULanguage(%q) = %v, want %v", tt.lang, got, tt.want)
		}
	}
}

func TestModelForLanguage_NoFallback(t *testing.T) {
	// Self-hosted: no fallback model configured — always returns default.
	g := &Gateway{
		defaultModel:  "small",
		fallbackModel: "",
		health:        newHealthState(),
	}
	if got := g.ModelForLanguage("en"); got != "small" {
		t.Errorf("got %q, want small", got)
	}
	if got := g.ModelForLanguage("ja"); got != "small" {
		t.Errorf("got %q, want small", got)
	}
}

func TestModelForLanguage_EULanguage(t *testing.T) {
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", true)
	g.health.set("large-v3-turbo", true)

	if got := g.ModelForLanguage("en"); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2", got)
	}
	if got := g.ModelForLanguage("cs"); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2", got)
	}
}

func TestModelForLanguage_NonEULanguage(t *testing.T) {
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", true)
	g.health.set("large-v3-turbo", true)

	if got := g.ModelForLanguage("ja"); got != "large-v3-turbo" {
		t.Errorf("got %q, want large-v3-turbo", got)
	}
	if got := g.ModelForLanguage("zh"); got != "large-v3-turbo" {
		t.Errorf("got %q, want large-v3-turbo", got)
	}
}

func TestModelForLanguage_EmptyLanguage(t *testing.T) {
	// Empty language defaults to the EU model (default model).
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", true)
	g.health.set("large-v3-turbo", true)

	if got := g.ModelForLanguage(""); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2", got)
	}
}

func TestModelForLanguage_HealthFallback_DefaultDown(t *testing.T) {
	// Default model is down, EU language → fall back to large-v3-turbo.
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", false)
	g.health.set("large-v3-turbo", true)

	if got := g.ModelForLanguage("en"); got != "large-v3-turbo" {
		t.Errorf("got %q, want large-v3-turbo (health fallback)", got)
	}
}

func TestModelForLanguage_HealthFallback_TurboDown(t *testing.T) {
	// large-v3-turbo is down, non-EU language → fall back to canary (better than nothing).
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", true)
	g.health.set("large-v3-turbo", false)

	if got := g.ModelForLanguage("ja"); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2 (health fallback)", got)
	}
}

func TestModelForLanguage_BothUnhealthy(t *testing.T) {
	// Both down — returns preferred anyway (might recover).
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", false)
	g.health.set("large-v3-turbo", false)

	if got := g.ModelForLanguage("en"); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2 (preferred even when unhealthy)", got)
	}
	if got := g.ModelForLanguage("ja"); got != "large-v3-turbo" {
		t.Errorf("got %q, want large-v3-turbo (preferred even when unhealthy)", got)
	}
}

func TestEULanguageSet_Exact25(t *testing.T) {
	// Guard against drift — the set must contain exactly 25 languages.
	if got := len(euLanguages); got != 25 {
		t.Errorf("euLanguages has %d entries, want 25", got)
	}
}

// --- Three-tier routing: englishModel ---

func testGatewayThreeTier() *Gateway {
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		englishModel:  "canary-qwen",
		health:        newHealthState(),
	}
	g.health.set("canary-qwen", true)
	g.health.set("canary-v2", true)
	g.health.set("large-v3-turbo", true)
	return g
}

func TestModelForLanguage_English_UsesEnglishModel(t *testing.T) {
	g := testGatewayThreeTier()
	if got := g.ModelForLanguage("en"); got != "canary-qwen" {
		t.Errorf("got %q, want canary-qwen", got)
	}
}

func TestModelForLanguage_Empty_UsesEnglishModel(t *testing.T) {
	// Empty language → assume English (most common case).
	g := testGatewayThreeTier()
	if got := g.ModelForLanguage(""); got != "canary-qwen" {
		t.Errorf("got %q, want canary-qwen", got)
	}
}

func TestModelForLanguage_OtherEU_UsesDefaultModel(t *testing.T) {
	g := testGatewayThreeTier()
	for _, lang := range []string{"cs", "de", "fr", "pl", "uk"} {
		if got := g.ModelForLanguage(lang); got != "canary-v2" {
			t.Errorf("lang=%q: got %q, want canary-v2", lang, got)
		}
	}
}

func TestModelForLanguage_NonEU_UsesFallback(t *testing.T) {
	g := testGatewayThreeTier()
	for _, lang := range []string{"ja", "zh", "ar", "ko"} {
		if got := g.ModelForLanguage(lang); got != "large-v3-turbo" {
			t.Errorf("lang=%q: got %q, want large-v3-turbo", lang, got)
		}
	}
}

func TestModelForLanguage_EnglishModelDown_FallsToCanary(t *testing.T) {
	g := testGatewayThreeTier()
	g.health.set("canary-qwen", false)
	if got := g.ModelForLanguage("en"); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2 (english model down)", got)
	}
}

func TestModelForLanguage_EnglishModelDown_EmptyLang_FallsToCanary(t *testing.T) {
	g := testGatewayThreeTier()
	g.health.set("canary-qwen", false)
	if got := g.ModelForLanguage(""); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2 (english model down)", got)
	}
}

func TestModelForLanguage_NoEnglishModel_EnUsesDefault(t *testing.T) {
	// Without englishModel, English routes to defaultModel like any other EU language.
	g := &Gateway{
		defaultModel:  "canary-v2",
		fallbackModel: "large-v3-turbo",
		englishModel:  "",
		health:        newHealthState(),
	}
	g.health.set("canary-v2", true)
	g.health.set("large-v3-turbo", true)
	if got := g.ModelForLanguage("en"); got != "canary-v2" {
		t.Errorf("got %q, want canary-v2", got)
	}
}
