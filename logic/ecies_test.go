package logic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
)

func TestECIESRoundTrip(t *testing.T) {
	recipientPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate recipient key: %v", err)
	}

	sessionKey := []byte("01234567890123456789012345678901")

	ephemeralPubPEM, encryptedKeyHex, ivHex, err := ECIESEncryptSessionKey(&recipientPriv.PublicKey, sessionKey)
	if err != nil {
		t.Fatalf("encrypt session key: %v", err)
	}

	decryptedKey, err := ECIESDecryptSessionKey(recipientPriv, ephemeralPubPEM, encryptedKeyHex, ivHex)
	if err != nil {
		t.Fatalf("decrypt session key: %v", err)
	}

	if !bytes.Equal(decryptedKey, sessionKey) {
		t.Fatalf("decrypted key mismatch: got %q, want %q", decryptedKey, sessionKey)
	}
}
