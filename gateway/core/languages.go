package core

import "strings"

// euLanguages is the set of 25 European language codes supported by both
// Parakeet v3 and Canary-1B-v2. These get the best-accuracy EU model;
// all other languages fall back to Whisper large-v3-turbo (99 languages).
var euLanguages = map[string]bool{
	"en": true, "bg": true, "hr": true, "cs": true, "da": true,
	"nl": true, "et": true, "fi": true, "fr": true, "de": true,
	"el": true, "hu": true, "it": true, "lv": true, "lt": true,
	"mt": true, "pl": true, "pt": true, "ro": true, "sk": true,
	"sl": true, "es": true, "sv": true, "ru": true, "uk": true,
}

// IsEULanguage reports whether the given language code is in the 25-language
// EU set supported by Parakeet v3 and Canary-1B-v2.
func IsEULanguage(lang string) bool {
	return euLanguages[strings.TrimSpace(strings.ToLower(lang))]
}

// ModelForLanguage returns the best model for the given language code.
//
// Three-tier routing (when all models are configured):
//
//  1. English (lang=="en" or empty): englishModel (canary-qwen-2.5b, best English accuracy)
//  2. EU languages (24 others): defaultModel (canary-1b-v2, multilingual EU)
//  3. Non-EU languages:         fallbackModel (large-v3-turbo, 99-language Whisper)
//
// If englishModel is not configured, tiers 1+2 collapse to defaultModel.
// If fallbackModel is not configured, all traffic goes to defaultModel.
// Health fallback: unhealthy preferred → try next tier. Both unhealthy → preferred anyway.
func (g *Gateway) ModelForLanguage(lang string) string {
	if g.fallbackModel == "" {
		return g.defaultModel
	}

	lang = strings.TrimSpace(strings.ToLower(lang))

	// Tier 1: English (or empty → assume English as most common case)
	if g.englishModel != "" && (lang == "en" || lang == "") {
		if g.health.get(g.englishModel) {
			return g.englishModel
		}
		// englishModel unhealthy — fall through to EU tier
	}

	// Tier 2: EU languages (including English when no englishModel, or as fallback)
	if lang == "" || euLanguages[lang] {
		if g.health.get(g.defaultModel) {
			return g.defaultModel
		}
		if g.health.get(g.fallbackModel) {
			return g.fallbackModel
		}
		return g.defaultModel
	}

	// Tier 3: Non-EU languages
	if g.health.get(g.fallbackModel) {
		return g.fallbackModel
	}
	if g.health.get(g.defaultModel) {
		return g.defaultModel
	}
	return g.fallbackModel
}
