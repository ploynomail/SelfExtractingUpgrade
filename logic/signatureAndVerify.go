package logic

import (
	"crypto/ecdsa"

	"github.com/ploynomail/SelfExtractingUpgrade/logic/keys"
)

type GenerateKeys interface {
	GenerateKeyPair() error
	SavePrivateKey(filename string) error
	LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error)
	LoadPublicKey(filename string) (*ecdsa.PublicKey, error)
	GetPrivateKey() *ecdsa.PrivateKey
	GetPublicKey() *ecdsa.PublicKey
}

func NewGenerateKeys() GenerateKeys {
	return keys.NewGenerateEcdsaKeys()
}
