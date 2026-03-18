package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// EncryptTranscript encrypts a transcript string for the given client ephemeral X25519 public key.
// Implements X25519 ECDH + HKDF-SHA256 key derivation + AES-256-GCM encryption.
// Both parties derive the same AES-256 key from the ECDH shared secret — no static server key needed.
// Returns base64url-encoded ciphertext and server ephemeral public key for the client to decrypt.
func EncryptTranscript(transcript string, clientPubB64 string) (ctB64, serverPubB64 string, err error) {
	// Decode client ephemeral public key
	clientPubBytes, err := base64.RawURLEncoding.DecodeString(clientPubB64)
	if err != nil {
		return "", "", fmt.Errorf("decode client pubkey: %w", err)
	}
	clientPub, err := ecdh.X25519().NewPublicKey(clientPubBytes)
	if err != nil {
		return "", "", fmt.Errorf("parse client pubkey: %w", err)
	}

	// Generate server ephemeral X25519 keypair
	serverPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("generate server key: %w", err)
	}

	// X25519 ECDH shared secret
	shared, err := serverPriv.ECDH(clientPub)
	if err != nil {
		return "", "", fmt.Errorf("ecdh: %w", err)
	}

	// HKDF-SHA256: derive 32-byte AES-256 key from shared secret
	key := hkdfSHA256(shared, nil, []byte("diction-transcript-v1"), 32)

	// AES-256-GCM encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := rand.Read(nonce); err != nil {
		return "", "", fmt.Errorf("nonce: %w", err)
	}
	// Wire format: nonce (12) + ciphertext + GCM tag (16)
	ciphertext := gcm.Seal(nonce, nonce, []byte(transcript), nil)

	return base64.RawURLEncoding.EncodeToString(ciphertext),
		base64.RawURLEncoding.EncodeToString(serverPriv.PublicKey().Bytes()),
		nil
}

// hkdfSHA256 derives keyLen bytes from ikm using HKDF-SHA256 (RFC 5869).
// Uses only stdlib crypto — no external dependencies required.
func hkdfSHA256(ikm, salt, info []byte, keyLen int) []byte {
	if salt == nil {
		salt = make([]byte, sha256.Size)
	}
	// Extract
	extractor := hmac.New(sha256.New, salt)
	extractor.Write(ikm)
	prk := extractor.Sum(nil)

	// Expand
	var okm []byte
	var prev []byte
	for i := 1; len(okm) < keyLen; i++ {
		h := hmac.New(sha256.New, prk)
		h.Write(prev)
		h.Write(info)
		h.Write([]byte{byte(i)})
		prev = h.Sum(nil)
		okm = append(okm, prev...)
	}
	return okm[:keyLen]
}
