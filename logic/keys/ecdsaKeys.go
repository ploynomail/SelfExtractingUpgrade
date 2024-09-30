package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
)

type GenerateEcdsaKeys struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewGenerateEcdsaKeys() *GenerateEcdsaKeys {
	return &GenerateEcdsaKeys{
		PrivateKey: nil,
	}
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
		g.GenerateKeyPair()
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
	os.WriteFile(filename+".key", pemPriv, 0644)
	os.WriteFile(filename+".pub", pemPub, 0644)
	return nil
}

func (g *GenerateEcdsaKeys) LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if os.IsNotExist(err) {
		g.GenerateKeyPair()
		g.SavePrivateKey(filename)
		return g.PrivateKey, nil
	}
	p, _ := pem.Decode(keyBytes)
	key, err := x509.ParseECPrivateKey(p.Bytes)
	if err != nil {
		return nil, err
	}
	g.PrivateKey = key
	return key, nil
}
