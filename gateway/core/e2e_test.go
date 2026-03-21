package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestLoadStaticKey(t *testing.T) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	b64 := base64.RawURLEncoding.EncodeToString(priv.Bytes())

	loaded, err := LoadStaticKey(b64)
	if err != nil {
		t.Fatalf("LoadStaticKey: %v", err)
	}
	if string(loaded.Bytes()) != string(priv.Bytes()) {
		t.Fatal("loaded key does not match original")
	}
}

func TestLoadStaticKey_Invalid(t *testing.T) {
	_, err := LoadStaticKey("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecryptRequest_RoundTrip(t *testing.T) {
	serverStaticPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serverStaticPubB64 := base64.RawURLEncoding.EncodeToString(serverStaticPriv.PublicKey().Bytes())

	clientPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())

	plaintext := `{"text":"hello world","context":"test"}`
	ctB64 := testClientEncrypt(t, clientPriv, serverStaticPubB64, []byte(plaintext))

	decrypted, err := DecryptRequest(ctB64, clientPubB64, serverStaticPriv)
	if err != nil {
		t.Fatalf("DecryptRequest: %v", err)
	}
	if string(decrypted) != plaintext {
		t.Fatalf("decrypted %q, want %q", string(decrypted), plaintext)
	}
}

func TestDecryptRequest_BadCiphertext(t *testing.T) {
	serverPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())

	_, err := DecryptRequest("dGhpcyBpcyBub3QgdmFsaWQ", clientPubB64, serverPriv)
	if err == nil {
		t.Fatal("expected error for bad ciphertext")
	}
}

func TestDecryptRequest_WrongKey(t *testing.T) {
	serverPriv1, _ := ecdh.X25519().GenerateKey(rand.Reader)
	serverPriv2, _ := ecdh.X25519().GenerateKey(rand.Reader)

	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())
	serverPub1B64 := base64.RawURLEncoding.EncodeToString(serverPriv1.PublicKey().Bytes())

	ctB64 := testClientEncrypt(t, clientPriv, serverPub1B64, []byte("secret"))

	_, err := DecryptRequest(ctB64, clientPubB64, serverPriv2)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestEncryptTranscript_DecryptRoundTrip(t *testing.T) {
	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())

	ct, pk, err := EncryptTranscript("hello from server", clientPubB64)
	if err != nil {
		t.Fatalf("EncryptTranscript: %v", err)
	}
	if ct == "" || pk == "" {
		t.Fatal("empty ct or pk")
	}

	serverPubBytes, _ := base64.RawURLEncoding.DecodeString(pk)
	serverPub, _ := ecdh.X25519().NewPublicKey(serverPubBytes)
	shared, _ := clientPriv.ECDH(serverPub)
	key := hkdfSHA256(shared, nil, []byte("diction-transcript-v1"), 32)

	ctBytes, _ := base64.RawURLEncoding.DecodeString(ct)
	decrypted := testAESGCMDecrypt(t, key, ctBytes)
	if string(decrypted) != "hello from server" {
		t.Fatalf("got %q, want %q", string(decrypted), "hello from server")
	}
}

func TestFullCleanupE2ERoundTrip(t *testing.T) {
	serverStaticPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	serverStaticPubB64 := base64.RawURLEncoding.EncodeToString(serverStaticPriv.PublicKey().Bytes())

	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())

	// iOS: encrypt request
	reqBody, _ := json.Marshal(map[string]string{"text": "raw text", "context": "email"})
	encReq := testClientEncrypt(t, clientPriv, serverStaticPubB64, reqBody)

	// Gateway: decrypt request
	decryptedReq, err := DecryptRequest(encReq, clientPubB64, serverStaticPriv)
	if err != nil {
		t.Fatalf("gateway decrypt: %v", err)
	}
	var parsed struct {
		Text    string `json:"text"`
		Context string `json:"context"`
	}
	if err := json.Unmarshal(decryptedReq, &parsed); err != nil {
		t.Fatalf("parse decrypted: %v", err)
	}
	if parsed.Text != "raw text" || parsed.Context != "email" {
		t.Fatalf("unexpected: %+v", parsed)
	}

	// Gateway: encrypt response
	respCT, respPK, err := EncryptTranscript("polished text", clientPubB64)
	if err != nil {
		t.Fatalf("encrypt response: %v", err)
	}

	// iOS: decrypt response
	serverRespPubBytes, _ := base64.RawURLEncoding.DecodeString(respPK)
	serverRespPub, _ := ecdh.X25519().NewPublicKey(serverRespPubBytes)
	respShared, _ := clientPriv.ECDH(serverRespPub)
	respKey := hkdfSHA256(respShared, nil, []byte("diction-transcript-v1"), 32)
	respCTBytes, _ := base64.RawURLEncoding.DecodeString(respCT)
	decryptedResp := testAESGCMDecrypt(t, respKey, respCTBytes)

	if string(decryptedResp) != "polished text" {
		t.Fatalf("response: got %q, want %q", string(decryptedResp), "polished text")
	}
}

// --- EncryptTranscript error paths ---

func TestEncryptTranscript_InvalidBase64(t *testing.T) {
	_, _, err := EncryptTranscript("hello", "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 client pubkey")
	}
}

func TestEncryptTranscript_WrongKeySize(t *testing.T) {
	// Valid base64 but only 16 bytes — X25519 requires exactly 32
	shortKey := base64.RawURLEncoding.EncodeToString(make([]byte, 16))
	_, _, err := EncryptTranscript("hello", shortKey)
	if err == nil {
		t.Fatal("expected error for wrong-size X25519 key")
	}
}

// --- DecryptRequest error paths ---

func TestDecryptRequest_InvalidClientPubBase64(t *testing.T) {
	serverPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	_, err := DecryptRequest("dGVzdA", "not-valid-base64!!!", serverPriv)
	if err == nil {
		t.Fatal("expected error for invalid client pubkey base64")
	}
}

func TestDecryptRequest_WrongSizeClientPub(t *testing.T) {
	serverPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	shortKey := base64.RawURLEncoding.EncodeToString(make([]byte, 16))
	_, err := DecryptRequest("dGVzdA", shortKey, serverPriv)
	if err == nil {
		t.Fatal("expected error for wrong-size client pubkey")
	}
}

func TestDecryptRequest_InvalidCiphertextBase64(t *testing.T) {
	serverPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())
	_, err := DecryptRequest("not-valid!!!", clientPubB64, serverPriv)
	if err == nil {
		t.Fatal("expected error for invalid ciphertext base64")
	}
}

func TestDecryptRequest_CiphertextTooShort(t *testing.T) {
	serverPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	clientPubB64 := base64.RawURLEncoding.EncodeToString(clientPriv.PublicKey().Bytes())
	// Only 4 bytes — shorter than GCM nonce (12 bytes)
	shortCT := base64.RawURLEncoding.EncodeToString([]byte("tiny"))
	_, err := DecryptRequest(shortCT, clientPubB64, serverPriv)
	if err == nil {
		t.Fatal("expected error for ciphertext shorter than nonce")
	}
}

// --- test helpers ---

// testClientEncrypt simulates iOS-side encryption:
// ECDH(client_priv, server_static_pub) + HKDF("diction-cleanup-req-v1") + AES-256-GCM
func testClientEncrypt(t *testing.T, clientPriv *ecdh.PrivateKey, serverPubB64 string, plaintext []byte) string {
	t.Helper()
	serverPubBytes, _ := base64.RawURLEncoding.DecodeString(serverPubB64)
	serverPub, _ := ecdh.X25519().NewPublicKey(serverPubBytes)
	shared, _ := clientPriv.ECDH(serverPub)
	key := hkdfSHA256(shared, nil, []byte("diction-cleanup-req-v1"), 32)
	return testAESGCMEncrypt(t, key, plaintext)
}

func testAESGCMEncrypt(t *testing.T, key, plaintext []byte) string {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	ct := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ct)
}

func testAESGCMDecrypt(t *testing.T, key, combined []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := combined[:gcm.NonceSize()]
	ct := combined[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	return plaintext
}
