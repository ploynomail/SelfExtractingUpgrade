package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type GenerateEcdsaKeys struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewGenerateEcdsaKeys() *GenerateEcdsaKeys {
	return &GenerateEcdsaKeys{}
}

func (g *GenerateEcdsaKeys) GetPrivateKey() *ecdsa.PrivateKey {
	return g.PrivateKey
}

func (g *GenerateEcdsaKeys) GetPublicKey() *ecdsa.PublicKey {
	return &g.PrivateKey.PublicKey
}

func (g *GenerateEcdsaKeys) GenerateKeyPair() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	g.PrivateKey = privateKey
	return nil
}

func (g *GenerateEcdsaKeys) SavePrivateKey(filename string) error {
	if g.PrivateKey == nil {
		if err := g.GenerateKeyPair(); err != nil {
			return fmt.Errorf("generate key pair: %w", err)
		}
	}
	keyBytes, err := x509.MarshalECPrivateKey(g.PrivateKey)
	if err != nil {
		return err
	}
	pemPriv := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})
	pubBytes, err := x509.MarshalPKIXPublicKey(&g.PrivateKey.PublicKey)
	if err != nil {
		return err
	}
	pemPub := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	if err := os.WriteFile(filename+".key", pemPriv, 0600); err != nil {
		return fmt.Errorf("save private key: %w", err)
	}
	if err := os.WriteFile(filename+".pub", pemPub, 0644); err != nil {
		return fmt.Errorf("save public key: %w", err)
	}
	return nil
}

func (g *GenerateEcdsaKeys) LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read private key file %q: %w", filename, err)
	}
	p, _ := pem.Decode(keyBytes)
	if p == nil {
		return nil, fmt.Errorf("no valid PEM block found in %q", filename)
	}
	key, err := x509.ParseECPrivateKey(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC private key: %w", err)
	}
	g.PrivateKey = key
	return key, nil
}

func (g *GenerateEcdsaKeys) LoadPublicKey(filename string) (*ecdsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read public key file %q: %w", filename, err)
	}
	p, _ := pem.Decode(keyBytes)
	if p == nil {
		return nil, fmt.Errorf("no valid PEM block found in %q", filename)
	}
	pubInterface, err := x509.ParsePKIXPublicKey(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	pubKey, ok := pubInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%q is not an ECDSA public key", filename)
	}
	return pubKey, nil
}
