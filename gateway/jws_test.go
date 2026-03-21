package main

// Tests for verifyAppleJWS.
//
// Strategy: override the package-level appleRootCA with a test-controlled
// P-384 CA (matching the real Apple Root CA G3 curve). We then sign a
// P-256 leaf cert (ES256) through the test CA and build synthetic JWS
// tokens to exercise every branch of the function.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

// buildTestPKI creates a self-signed test CA (P-384) and a leaf cert (P-256)
// signed by that CA. Returns (caKey, caCert, leafKey, leafCert).
func buildTestPKI(t *testing.T) (*ecdsa.PrivateKey, *x509.Certificate, *ecdsa.PrivateKey, *x509.Certificate) {
	t.Helper()

	// CA: P-384 to match Apple Root CA G3
	caKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}

	// Leaf: P-256 (ES256)
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Leaf"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create leaf cert: %v", err)
	}
	leafCert, err := x509.ParseCertificate(leafDER)
	if err != nil {
		t.Fatalf("parse leaf cert: %v", err)
	}

	return caKey, caCert, leafKey, leafCert
}

// buildJWS constructs a JWS string (header.payload.signature) signed with
// leafKey, with the x5c chain [leafCert, caCert].
func buildJWS(t *testing.T, leafKey *ecdsa.PrivateKey, leafCert, caCert *x509.Certificate, payload jwsPayload) string {
	t.Helper()

	// Header
	header := jwsHeader{
		Alg: "ES256",
		X5c: []string{
			base64.StdEncoding.EncodeToString(leafCert.Raw),
			base64.StdEncoding.EncodeToString(caCert.Raw),
		},
	}
	headerBytes, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Payload
	payloadBytes, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Sign header.payload
	signingInput := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, leafKey, hash[:])
	if err != nil {
		t.Fatalf("sign JWS: %v", err)
	}

	// ES256 signature: 32-byte r || 32-byte s
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	sig := make([]byte, 64)
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return headerB64 + "." + payloadB64 + "." + sigB64
}

// withTestCA temporarily replaces the package-level appleRootCA with the
// provided cert and restores the original on test cleanup.
func withTestCA(t *testing.T, ca *x509.Certificate) {
	t.Helper()
	original := appleRootCA
	appleRootCA = ca
	t.Cleanup(func() { appleRootCA = original })
}

// --- Tests ---

func TestVerifyAppleJWS_CacheHit(t *testing.T) {
	// Seed a still-valid cache entry and confirm verifyAppleJWS returns nil
	// without doing any real crypto work.
	fakeToken := "fake.jws.cache_hit_test"
	tokenCache.Store(fakeToken, cachedToken{
		expiresAt: time.Now().Add(5 * time.Minute),
		claimExp:  time.Now().Add(24 * time.Hour),
	})
	t.Cleanup(func() { tokenCache.Delete(fakeToken) })

	if err := verifyAppleJWS(fakeToken, "one.diction"); err != nil {
		t.Errorf("expected cache hit to return nil, got: %v", err)
	}
}

func TestVerifyAppleJWS_CacheExpiredEvicted(t *testing.T) {
	// Seed an expired cache entry - should be deleted and re-verified (fails
	// because the token itself is invalid).
	fakeToken := "fake.jws.cache_evict_test"
	tokenCache.Store(fakeToken, cachedToken{
		expiresAt: time.Now().Add(-1 * time.Second), // expired
		claimExp:  time.Now().Add(24 * time.Hour),
	})

	err := verifyAppleJWS(fakeToken, "one.diction")
	if err == nil {
		t.Fatal("expected error after evicted cache entry")
	}
	if _, ok := tokenCache.Load(fakeToken); ok {
		t.Error("stale entry should have been deleted")
	}
}

func TestVerifyAppleJWS_WrongPartCount(t *testing.T) {
	err := verifyAppleJWS("onlytwoparts.here", "one.diction")
	if err == nil {
		t.Fatal("expected error for 2-part token")
	}
}

func TestVerifyAppleJWS_BadHeaderBase64(t *testing.T) {
	err := verifyAppleJWS("!!!.payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error for bad header base64")
	}
}

func TestVerifyAppleJWS_BadHeaderJSON(t *testing.T) {
	// Valid base64 but decodes to non-JSON
	badHeader := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	err := verifyAppleJWS(badHeader+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error for bad header JSON")
	}
}

func TestVerifyAppleJWS_WrongAlg(t *testing.T) {
	header := jwsHeader{Alg: "RS256", X5c: []string{"cert"}}
	hBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(hBytes)
	err := verifyAppleJWS(hB64+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error for wrong algorithm")
	}
}

func TestVerifyAppleJWS_EmptyX5c(t *testing.T) {
	header := jwsHeader{Alg: "ES256", X5c: []string{}}
	hBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(hBytes)
	err := verifyAppleJWS(hB64+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error for empty x5c")
	}
}

func TestVerifyAppleJWS_BadX5cBase64(t *testing.T) {
	header := jwsHeader{Alg: "ES256", X5c: []string{"!!not-base64!!"}}
	hBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(hBytes)
	err := verifyAppleJWS(hB64+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error for bad x5c base64")
	}
}

func TestVerifyAppleJWS_BadX5cDER(t *testing.T) {
	// Valid base64 but not a valid DER certificate.
	invalidDER := base64.StdEncoding.EncodeToString([]byte("not-a-cert"))
	header := jwsHeader{Alg: "ES256", X5c: []string{invalidDER}}
	hBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(hBytes)
	err := verifyAppleJWS(hB64+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error for invalid DER in x5c")
	}
}

func TestVerifyAppleJWS_ChainVerifyFails(t *testing.T) {
	// Two unrelated self-signed certs - x5c[0].CheckSignatureFrom(x5c[1]) fails.
	key1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	key2, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := func(cn string) *x509.Certificate {
		return &x509.Certificate{
			SerialNumber: big.NewInt(99),
			Subject:      pkix.Name{CommonName: cn},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(time.Hour),
			IsCA:         true, BasicConstraintsValid: true,
		}
	}
	der1, _ := x509.CreateCertificate(rand.Reader, tmpl("C1"), tmpl("C1"), &key1.PublicKey, key1)
	der2, _ := x509.CreateCertificate(rand.Reader, tmpl("C2"), tmpl("C2"), &key2.PublicKey, key2)

	header := jwsHeader{
		Alg: "ES256",
		X5c: []string{
			base64.StdEncoding.EncodeToString(der1),
			base64.StdEncoding.EncodeToString(der2),
		},
	}
	hBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(hBytes)
	err := verifyAppleJWS(hB64+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected chain verification failure")
	}
}

func TestVerifyAppleJWS_NotSignedByAppleRoot(t *testing.T) {
	// Single self-signed cert not signed by the real Apple Root CA.
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Rogue CA"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true, BasicConstraintsValid: true,
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	header := jwsHeader{
		Alg: "ES256",
		X5c: []string{base64.StdEncoding.EncodeToString(der)},
	}
	hBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(hBytes)
	err := verifyAppleJWS(hB64+".payload.sig", "one.diction")
	if err == nil {
		t.Fatal("expected error: cert not signed by Apple Root CA")
	}
}

// --- Tests using the synthetic PKI (full chain passes verification) ---

func TestVerifyAppleJWS_HappyPath(t *testing.T) {
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	payload := jwsPayload{
		BundleID:    "one.diction",
		ExpiresDate: time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	tokenCache.Delete(token)

	if err := verifyAppleJWS(token, "one.diction"); err != nil {
		t.Errorf("expected valid JWS to pass, got: %v", err)
	}

	// Should now be cached - second call with a garbage secret still passes.
	if err := verifyAppleJWS(token, "one.diction"); err != nil {
		t.Errorf("expected cached token to pass on second call: %v", err)
	}
	tokenCache.Delete(token)
}

func TestVerifyAppleJWS_BundleIDMismatch(t *testing.T) {
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	payload := jwsPayload{
		BundleID:    "com.other.app",
		ExpiresDate: time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	tokenCache.Delete(token)

	err := verifyAppleJWS(token, "one.diction")
	if err == nil {
		t.Fatal("expected error for bundle ID mismatch")
	}
}

func TestVerifyAppleJWS_ExpiredSubscription(t *testing.T) {
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	payload := jwsPayload{
		BundleID:    "one.diction",
		ExpiresDate: time.Now().Add(-1 * time.Hour).UnixMilli(), // expired
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	tokenCache.Delete(token)

	err := verifyAppleJWS(token, "one.diction")
	if err == nil {
		t.Fatal("expected error for expired subscription")
	}
	ae, ok := err.(*authError)
	if !ok || ae.reason != "expired_subscription" {
		t.Errorf("want expired_subscription authError, got: %v", err)
	}
}

func TestVerifyAppleJWS_RevokedSubscription(t *testing.T) {
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	revocationDate := time.Now().Add(-1 * time.Hour).UnixMilli()
	payload := jwsPayload{
		BundleID:       "one.diction",
		ExpiresDate:    time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
		RevocationDate: &revocationDate,
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	tokenCache.Delete(token)

	err := verifyAppleJWS(token, "one.diction")
	if err == nil {
		t.Fatal("expected error for revoked subscription")
	}
	ae, ok := err.(*authError)
	if !ok || ae.reason != "revoked" {
		t.Errorf("want revoked authError, got: %v", err)
	}
}

func TestVerifyAppleJWS_BadSigBase64(t *testing.T) {
	// Build a valid header+payload, but put garbage in the signature slot.
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	payload := jwsPayload{
		BundleID:    "one.diction",
		ExpiresDate: time.Now().Add(24 * time.Hour).UnixMilli(),
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	parts := splitDots(token)
	badToken := parts[0] + "." + parts[1] + ".!!!bad-sig!!!"
	tokenCache.Delete(badToken)

	err := verifyAppleJWS(badToken, "one.diction")
	if err == nil {
		t.Fatal("expected error for bad signature base64")
	}
}

func TestVerifyAppleJWS_WrongSigLength(t *testing.T) {
	// Valid base64 sig but wrong number of bytes (not 64).
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	payload := jwsPayload{
		BundleID:    "one.diction",
		ExpiresDate: time.Now().Add(24 * time.Hour).UnixMilli(),
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	parts := splitDots(token)
	shortSig := base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x02}) // only 2 bytes
	badToken := parts[0] + "." + parts[1] + "." + shortSig
	tokenCache.Delete(badToken)

	err := verifyAppleJWS(badToken, "one.diction")
	if err == nil {
		t.Fatal("expected error for wrong signature length")
	}
}

func TestVerifyAppleJWS_SignatureVerificationFails(t *testing.T) {
	// Flip a byte in an otherwise-valid signature so ECDSA verify fails.
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	payload := jwsPayload{
		BundleID:    "one.diction",
		ExpiresDate: time.Now().Add(24 * time.Hour).UnixMilli(),
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	parts := splitDots(token)
	sigBytes, _ := base64.RawURLEncoding.DecodeString(parts[2])
	sigBytes[0] ^= 0xFF // corrupt first byte
	corruptedSig := base64.RawURLEncoding.EncodeToString(sigBytes)
	badToken := parts[0] + "." + parts[1] + "." + corruptedSig
	tokenCache.Delete(badToken)

	err := verifyAppleJWS(badToken, "one.diction")
	if err == nil {
		t.Fatal("expected signature verification failure")
	}
}

func TestVerifyAppleJWS_CacheExpiryClampedToSubscription(t *testing.T) {
	// When the subscription expires before tokenCacheTTL, cacheExpiry should
	// be clamped to the subscription expiry (exercises the if branch).
	_, caCert, leafKey, leafCert := buildTestPKI(t)
	withTestCA(t, caCert)

	// Expires in 1 minute - less than the 5-minute tokenCacheTTL
	payload := jwsPayload{
		BundleID:    "one.diction",
		ExpiresDate: time.Now().Add(1 * time.Minute).UnixMilli(),
	}
	token := buildJWS(t, leafKey, leafCert, caCert, payload)
	tokenCache.Delete(token)
	t.Cleanup(func() { tokenCache.Delete(token) })

	if err := verifyAppleJWS(token, "one.diction"); err != nil {
		t.Errorf("expected valid token to pass: %v", err)
	}
}

// splitDots splits a JWS token into its 3 dot-separated parts.
func splitDots(token string) [3]string {
	var out [3]string
	idx := [2]int{}
	n := 0
	for i, c := range token {
		if c == '.' {
			idx[n] = i
			n++
			if n == 2 {
				break
			}
		}
	}
	out[0] = token[:idx[0]]
	out[1] = token[idx[0]+1 : idx[1]]
	out[2] = token[idx[1]+1:]
	return out
}
