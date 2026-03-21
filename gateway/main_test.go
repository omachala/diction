package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// testSecret is a 32-byte key used across trial token tests.
var testSecret = []byte("0123456789abcdef0123456789abcdef")

// validDeviceID is a UUID that passes the handleTrial validation.
const validDeviceID = "550E8400-E29B-41D4-A716-446655440000"

// --- trialStore ---

func newTempTrialStore(t *testing.T) *trialStore {
	t.Helper()
	dir := t.TempDir()
	return newTrialStore(dir + "/trials.json")
}

func TestTrialStore_InitiallyEmpty(t *testing.T) {
	s := newTempTrialStore(t)
	_, exists := s.getTrial("any-device")
	if exists {
		t.Error("expected empty store")
	}
}

func TestTrialStore_GrantAndGet(t *testing.T) {
	s := newTempTrialStore(t)
	exp := time.Now().Add(24 * time.Hour)
	s.grantTrial(validDeviceID, exp, "trial")

	r, exists := s.getTrial(validDeviceID)
	if !exists {
		t.Fatal("expected record after grant")
	}
	if r.DeviceID != validDeviceID {
		t.Errorf("device_id: want %s, got %s", validDeviceID, r.DeviceID)
	}
	if r.TokenType != "trial" {
		t.Errorf("token_type: want trial, got %s", r.TokenType)
	}
	if r.ExpiresAt.Unix() != exp.Unix() {
		t.Errorf("expires_at mismatch")
	}
}

func TestTrialStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/trials.json"

	s1 := newTrialStore(path)
	exp := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	s1.grantTrial(validDeviceID, exp, "trial")

	// Reload from same path
	s2 := newTrialStore(path)
	r, exists := s2.getTrial(validDeviceID)
	if !exists {
		t.Fatal("expected record to persist across reload")
	}
	if r.DeviceID != validDeviceID {
		t.Errorf("device_id mismatch after reload")
	}
}

func TestTrialStore_OverwriteExisting(t *testing.T) {
	s := newTempTrialStore(t)
	exp1 := time.Now().Add(1 * time.Hour)
	exp2 := time.Now().Add(48 * time.Hour)

	s.grantTrial(validDeviceID, exp1, "trial")
	s.grantTrial(validDeviceID, exp2, "trial")

	r, _ := s.getTrial(validDeviceID)
	if r.ExpiresAt.Unix() != exp2.Unix() {
		t.Error("expected second grant to overwrite first")
	}
}

// --- generateTrialToken + verifyTrialToken ---

func TestTrialToken_RoundTrip(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	token := generateTrialToken(validDeviceID, exp, "trial", testSecret)
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if err := verifyTrialToken(token, testSecret); err != nil {
		t.Errorf("expected valid token to pass: %v", err)
	}
}

func TestTrialToken_ExpiredFails(t *testing.T) {
	// Token expired 1 second ago
	exp := time.Now().Add(-1 * time.Second)
	token := generateTrialToken(validDeviceID, exp, "trial", testSecret)

	err := verifyTrialToken(token, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	ae, ok := err.(*authError)
	if !ok {
		t.Fatalf("expected authError, got %T: %v", err, err)
	}
	if ae.reason != "expired_trial" {
		t.Errorf("reason: want expired_trial, got %s", ae.reason)
	}
}

func TestTrialToken_TamperedSignatureFails(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	token := generateTrialToken(validDeviceID, exp, "trial", testSecret)

	// Flip a character in the signature portion
	parts := strings.Split(token, ".")
	sig := []byte(parts[1])
	sig[0] ^= 0x01
	tampered := parts[0] + "." + string(sig)

	err := verifyTrialToken(tampered, testSecret)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestTrialToken_WrongSecretFails(t *testing.T) {
	// Use a unique device ID to avoid cache collision with other token tests.
	exp := time.Now().Add(24 * time.Hour)
	token := generateTrialToken("WRONG-SECRET-TEST-0000-00000000001", exp, "trial", testSecret)
	tokenCache.Delete(token) // ensure no stale cache entry

	wrongSecret := []byte("different-secret-key-0123456789ab")
	err := verifyTrialToken(token, wrongSecret)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestTrialToken_InvalidFormat(t *testing.T) {
	err := verifyTrialToken("notavalidtoken", testSecret)
	if err == nil {
		t.Fatal("expected error for token with no dot separator")
	}
}

func TestTrialToken_CachesValidToken(t *testing.T) {
	// Use a unique expiry to avoid colliding with other test cache entries
	exp := time.Now().Add(25 * time.Hour)
	token := generateTrialToken("CACHE-TEST-0000-0000-000000000001", exp, "trial", testSecret)

	// First call verifies from scratch
	if err := verifyTrialToken(token, testSecret); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	// Second call should hit cache (with wrong secret - still passes due to cache)
	if err := verifyTrialToken(token, []byte("wrong")); err != nil {
		t.Errorf("second verify should hit cache and pass, got: %v", err)
	}

	// Clean up cache entry
	tokenCache.Delete(token)
}

// --- handleTrial ---

func trialRequest(t *testing.T, method, body string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, "/v1/trial", bodyReader)
	req.Header.Set("Content-Type", "application/json")
	return httptest.NewRecorder(), req
}

func TestHandleTrial_MethodNotAllowed(t *testing.T) {
	s := newTempTrialStore(t)
	rr, req := trialRequest(t, http.MethodGet, "")
	handleTrial(rr, req, s, testSecret, 24*time.Hour)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: want 405, got %d", rr.Code)
	}
}

func TestHandleTrial_NoSecret(t *testing.T) {
	s := newTempTrialStore(t)
	rr, req := trialRequest(t, http.MethodPost, `{"device_id":"`+validDeviceID+`"}`)
	handleTrial(rr, req, s, nil, 24*time.Hour) // nil secret
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status: want 503, got %d", rr.Code)
	}
}

func TestHandleTrial_InvalidBody(t *testing.T) {
	s := newTempTrialStore(t)
	rr, req := trialRequest(t, http.MethodPost, `not-json`)
	handleTrial(rr, req, s, testSecret, 24*time.Hour)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", rr.Code)
	}
}

func TestHandleTrial_InvalidDeviceID_TooShort(t *testing.T) {
	s := newTempTrialStore(t)
	rr, req := trialRequest(t, http.MethodPost, `{"device_id":"short"}`)
	handleTrial(rr, req, s, testSecret, 24*time.Hour)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", rr.Code)
	}
}

func TestHandleTrial_InvalidDeviceID_NoHyphens(t *testing.T) {
	s := newTempTrialStore(t)
	// 36 chars but no hyphens
	rr, req := trialRequest(t, http.MethodPost, `{"device_id":"550E8400E29B41D4A716446655440000xxxx"}`)
	handleTrial(rr, req, s, testSecret, 24*time.Hour)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", rr.Code)
	}
}

func TestHandleTrial_NewGrant(t *testing.T) {
	s := newTempTrialStore(t)
	rr, req := trialRequest(t, http.MethodPost, fmt.Sprintf(`{"device_id":"%s"}`, validDeviceID))
	handleTrial(rr, req, s, testSecret, 24*time.Hour)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
	if resp["expires_at"] == "" {
		t.Error("expected expires_at in response")
	}
}

func TestHandleTrial_ReissueBeforeExpiry(t *testing.T) {
	s := newTempTrialStore(t)

	// First grant
	body := fmt.Sprintf(`{"device_id":"%s"}`, validDeviceID)
	rr1, req1 := trialRequest(t, http.MethodPost, body)
	handleTrial(rr1, req1, s, testSecret, 24*time.Hour)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first grant failed: %d", rr1.Code)
	}
	var resp1 map[string]string
	json.NewDecoder(rr1.Body).Decode(&resp1)

	// Second call (re-issue)
	rr2, req2 := trialRequest(t, http.MethodPost, body)
	handleTrial(rr2, req2, s, testSecret, 24*time.Hour)
	if rr2.Code != http.StatusOK {
		t.Fatalf("re-issue failed: %d (body: %s)", rr2.Code, rr2.Body.String())
	}
	var resp2 map[string]string
	json.NewDecoder(rr2.Body).Decode(&resp2)

	// Same expiry → same token (deterministic HMAC)
	if resp1["token"] != resp2["token"] {
		t.Error("expected same token on re-issue (same expiry)")
	}
}

func TestHandleTrial_ExpiredTrialConflict(t *testing.T) {
	s := newTempTrialStore(t)
	// Pre-insert an already-expired trial
	expired := time.Now().Add(-1 * time.Hour)
	s.grantTrial(validDeviceID, expired, "trial")

	rr, req := trialRequest(t, http.MethodPost, fmt.Sprintf(`{"device_id":"%s"}`, validDeviceID))
	handleTrial(rr, req, s, testSecret, 24*time.Hour)

	if rr.Code != http.StatusConflict {
		t.Errorf("status: want 409, got %d (body: %s)", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "trial_already_used" {
		t.Errorf("error: want trial_already_used, got %s", resp["error"])
	}
}

// --- authMiddleware ---

// okHandler is a trivial handler that returns 200.
var okHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAuthMiddleware_Disabled(t *testing.T) {
	handler := authMiddleware(okHandler, false, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil) // no Authorization
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("disabled auth: want 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
	assertReason(t, rr, "missing_token")
}

func TestAuthMiddleware_EmptyBearer(t *testing.T) {
	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
	assertReason(t, rr, "missing_token")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// 3+ dots → not a trial (1 dot) or JWS (2 dots) → invalid_token
	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer a.b.c.d")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
	assertReason(t, rr, "invalid_token")
}

func TestAuthMiddleware_ValidTrialToken(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	token := generateTrialToken(validDeviceID, exp, "trial", testSecret)

	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("valid trial token: want 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ExpiredTrialToken(t *testing.T) {
	exp := time.Now().Add(-1 * time.Second)
	token := generateTrialToken(validDeviceID, exp, "trial", testSecret)
	// Ensure this token isn't cached from another test
	tokenCache.Delete(token)

	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
	assertReason(t, rr, "expired_trial")
}

func TestAuthMiddleware_TamperedTrialToken(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	token := generateTrialToken(validDeviceID, exp, "trial", testSecret)
	// Tamper the payload byte
	parts := strings.Split(token, ".")
	payload := []byte(parts[0])
	payload[0] ^= 0x01
	tampered := string(payload) + "." + parts[1]
	tokenCache.Delete(tampered)

	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_NoSecretRejectsTrialToken(t *testing.T) {
	// Trial token format (1 dot) but gateway has no secret configured
	handler := authMiddleware(okHandler, true, "one.diction", nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer payload.signature")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
	assertReason(t, rr, "invalid_token")
}

func TestAuthMiddleware_InvalidJWSToken(t *testing.T) {
	// 2-dot format → tries Apple JWS verification → should fail
	handler := authMiddleware(okHandler, true, "one.diction", testSecret)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.realJWS")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
}

// --- base64URLDecode ---

func TestBase64URLDecode_Standard(t *testing.T) {
	// "hello" in standard base64 is "aGVsbG8="
	// URL-safe variant uses - and _ instead of + and /
	data, err := base64URLDecode("aGVsbG8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("want 'hello', got %q", string(data))
	}
}

func TestBase64URLDecode_URLSafeChars(t *testing.T) {
	// Encode some bytes that produce + and / in standard base64
	// 0xFB 0xFF → +/8= in standard, -_8= in URL-safe
	input := "-_8"
	_, err := base64URLDecode(input)
	if err != nil {
		t.Fatalf("URL-safe decode failed: %v", err)
	}
}

// --- helpers ---

func assertReason(t *testing.T, rr *httptest.ResponseRecorder, wantReason string) {
	t.Helper()
	var resp authErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode auth error response: %v", err)
	}
	if resp.Reason != wantReason {
		t.Errorf("reason: want %s, got %s", wantReason, resp.Reason)
	}
}

// --- writeAuthError ---

func TestWriteAuthError_UnknownReason(t *testing.T) {
	// An unrecognised reason falls back to the "invalid_token" message.
	rr := httptest.NewRecorder()
	writeAuthError(rr, "totally_unknown_reason")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", rr.Code)
	}
	var resp authErrorResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Reason != "totally_unknown_reason" {
		t.Errorf("reason: want totally_unknown_reason, got %s", resp.Reason)
	}
	// Message should fall back to the invalid_token message, not empty.
	if resp.Message == "" {
		t.Error("expected fallback message, got empty string")
	}
}

// --- base64URLDecode padding ---

func TestBase64URLDecode_NoPadding(t *testing.T) {
	// len%4 == 0: no padding needed - "dGVzdA" (=base64("test") without "==")
	// "dGVzdA==" → "test"
	data, err := base64URLDecode("dGVzdA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "test" {
		t.Errorf("want 'test', got %q", string(data))
	}
}

func TestBase64URLDecode_OnePadByte(t *testing.T) {
	// len%4 == 3: adds one "=" → "dGVzdA8" → decode 5 bytes
	data, err := base64URLDecode("dGVzdA8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty decoded output")
	}
}

func TestBase64URLDecode_TwoPadBytes(t *testing.T) {
	// len%4 == 2: adds "==" → "dA" → decode 1 byte (0x74 = 't')
	data, err := base64URLDecode("dA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "t" {
		t.Errorf("want 't', got %q", string(data))
	}
}

// --- verifyTrialToken cache eviction ---

func TestVerifyTrialToken_ShortExpiryClampsCache(t *testing.T) {
	// Token expires in 2 minutes - less than tokenCacheTTL (5 min).
	// This exercises the cacheExpiry = tokenExp branch.
	exp := time.Now().Add(2 * time.Minute)
	token := generateTrialToken("CLAMP-TEST-0000-0000-000000000001", exp, "trial", testSecret)
	tokenCache.Delete(token)
	t.Cleanup(func() { tokenCache.Delete(token) })

	if err := verifyTrialToken(token, testSecret); err != nil {
		t.Errorf("expected short-expiry token to be valid: %v", err)
	}
}

func TestVerifyTrialToken_InvalidPayloadJSON(t *testing.T) {
	// Construct a token where payload is valid base64 but not JSON.
	// Signature must match so we can get past the HMAC check.
	payload := []byte("not-json-at-all")
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, testSecret)
	mac.Write(payload)
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := payloadB64 + "." + sigB64
	tokenCache.Delete(token)

	err := verifyTrialToken(token, testSecret)
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestVerifyTrialToken_ExpiredCacheEntryEvicted(t *testing.T) {
	// Seed the cache with an already-expired entry for a unique token string.
	fakeToken := "fakePayload.fakeSig_expired_eviction"
	tokenCache.Store(fakeToken, cachedToken{
		expiresAt: time.Now().Add(-1 * time.Second), // already expired
		claimExp:  time.Now().Add(24 * time.Hour),
	})
	// verifyTrialToken should detect the stale cache entry, delete it, then
	// attempt real verification (which fails on the fake format).
	err := verifyTrialToken(fakeToken, testSecret)
	if err == nil {
		t.Fatal("expected error after stale cache eviction")
	}
	// Confirm entry was removed from cache.
	if _, ok := tokenCache.Load(fakeToken); ok {
		t.Error("expected stale cache entry to be deleted")
	}
}

// --- trialStore.save error paths ---

func TestTrialStore_SaveIgnoresWriteError(t *testing.T) {
	// Point the store at a path inside a nonexistent, unwritable directory.
	// save() logs the error but does not panic.
	s := &trialStore{
		records: make(map[string]trialRecord),
		path:    "/nonexistent_dir_xyz/trials.json",
	}
	s.grantTrial(validDeviceID, time.Now().Add(time.Hour), "trial")
	// If we get here without panic, save() handled the error gracefully.
}

// --- TRIAL_SECRET env parsing (smoke test for main() hex decoding) ---

func TestTrialSecretHexDecoding(t *testing.T) {
	secretHex := hex.EncodeToString(testSecret)
	decoded, err := hex.DecodeString(secretHex)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(decoded, testSecret) {
		t.Error("decoded secret does not match original")
	}
}

// --- buildMux (covers env-var wiring extracted from main) ---

func TestBuildMux_Defaults(t *testing.T) {
	for _, k := range []string{"GATEWAY_PORT", "DEFAULT_MODEL", "MAX_BODY_SIZE",
		"AUTH_ENABLED", "BUNDLE_ID", "TRIAL_SECRET", "TRIAL_DB_PATH", "TRIAL_DURATION"} {
		os.Unsetenv(k)
	}
	os.Setenv("TRIAL_DB_PATH", t.TempDir()+"/trials.json")
	defer os.Unsetenv("TRIAL_DB_PATH")

	mux, port, err := buildMux()
	if err != nil {
		t.Fatalf("buildMux: %v", err)
	}
	if mux == nil {
		t.Fatal("expected non-nil mux")
	}
	if port != "8080" {
		t.Errorf("port: want 8080, got %s", port)
	}
}

func TestBuildMux_CustomPort(t *testing.T) {
	os.Setenv("GATEWAY_PORT", "9999")
	os.Setenv("TRIAL_DB_PATH", t.TempDir()+"/trials.json")
	defer func() { os.Unsetenv("GATEWAY_PORT"); os.Unsetenv("TRIAL_DB_PATH") }()

	_, port, err := buildMux()
	if err != nil {
		t.Fatalf("buildMux: %v", err)
	}
	if port != "9999" {
		t.Errorf("port: want 9999, got %s", port)
	}
}

func TestBuildMux_InvalidTrialSecret(t *testing.T) {
	os.Setenv("TRIAL_SECRET", "not-valid-hex!!")
	os.Setenv("TRIAL_DB_PATH", t.TempDir()+"/trials.json")
	defer func() { os.Unsetenv("TRIAL_SECRET"); os.Unsetenv("TRIAL_DB_PATH") }()

	_, _, err := buildMux()
	if err == nil {
		t.Fatal("expected error for invalid TRIAL_SECRET hex")
	}
}

func TestBuildMux_InvalidTrialDuration(t *testing.T) {
	os.Setenv("TRIAL_DURATION", "notaduration")
	os.Setenv("TRIAL_DB_PATH", t.TempDir()+"/trials.json")
	defer func() { os.Unsetenv("TRIAL_DURATION"); os.Unsetenv("TRIAL_DB_PATH") }()

	_, _, err := buildMux()
	if err == nil {
		t.Fatal("expected error for invalid TRIAL_DURATION")
	}
}

func TestBuildMux_ValidTrialSecret(t *testing.T) {
	os.Setenv("TRIAL_SECRET", hex.EncodeToString(testSecret))
	os.Setenv("TRIAL_DB_PATH", t.TempDir()+"/trials.json")
	defer func() { os.Unsetenv("TRIAL_SECRET"); os.Unsetenv("TRIAL_DB_PATH") }()

	mux, _, err := buildMux()
	if err != nil || mux == nil {
		t.Fatalf("buildMux with valid secret: err=%v, mux=%v", err, mux)
	}
}

func TestBuildMux_RoutesRegistered(t *testing.T) {
	os.Setenv("TRIAL_DB_PATH", t.TempDir()+"/trials.json")
	defer os.Unsetenv("TRIAL_DB_PATH")

	mux, _, err := buildMux()
	if err != nil {
		t.Fatalf("buildMux: %v", err)
	}

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/health: want 200, got %d", resp.StatusCode)
	}
}
