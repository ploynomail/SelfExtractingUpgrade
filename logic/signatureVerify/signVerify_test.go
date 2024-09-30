package signatureverify

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"testing"
)

func TestSignFile(t *testing.T) {
	// Generate a private key for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some data to the temporary file
	data := []byte("test data")
	if _, err := tempFile.Write(data); err != nil {
		t.Fatalf("Failed to write data to temporary file: %v", err)
	}

	// Reset the file offset to the beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek to the beginning of the temporary file: %v", err)
	}

	// Call the SignFile function
	signature, err := SignFile(privateKey, tempFile)
	if err != nil {
		t.Fatalf("SignFile failed: %v", err)
	}

	// Verify the signature using the public key
	publicKey := &privateKey.PublicKey
	tempfile, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temporary file: %v", err)
	}
	defer tempfile.Close()
	isPass, err := VerifyFile(publicKey, tempfile, signature)
	if err != nil {
		t.Fatalf("VerifyFile failed: %v", err)
	}
	if !isPass {
		t.Fatalf("Signature verification failed")
	}
	t.Logf("Signature verification passed")

	data = []byte("test data 2")
	if _, err := tempFile.Write(data); err != nil {
		t.Fatalf("Failed to write data to temporary file: %v", err)
	}
	if _, err := tempFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek to the beginning of the temporary file: %v", err)
	}
	tempfile, err = os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temporary file: %v", err)
	}
	defer tempfile.Close()
	isPass, err = VerifyFile(publicKey, tempfile, signature)
	if err != nil {
		t.Fatalf("VerifyFile failed: %v", err)
	}
	if isPass {
		t.Fatalf("Signature verification failed")
	}
}
