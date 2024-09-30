package logic

import (
	"crypto/ecdsa"

	"github.com/ploynomail/SelfExtractingUpgrade/logic/keys"
)

type GenerateKeys interface {
	GenerateKeyPair() error
	SavePrivateKey(filename string) error
	LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error)
	GetPrivateKey() *ecdsa.PrivateKey
	GetPublicKey() *ecdsa.PublicKey
}

type SignatureAndVerify interface {
	Sign() error
	Verify() error
}

func NewGenerateKeys() GenerateKeys {
	return keys.NewGenerateEcdsaKeys()
}
