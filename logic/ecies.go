package logic

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
)

func ecdsaPublicToECDH(pub *ecdsa.PublicKey) (*ecdh.PublicKey, error) {
	curve := ecdh.P256()
	pubBytes := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	return curve.NewPublicKey(pubBytes)
}

func ecdsaPrivateToECDH(priv *ecdsa.PrivateKey) (*ecdh.PrivateKey, error) {
	curve := ecdh.P256()
	d := priv.D.Bytes()
	if len(d) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(d):], d)
		d = padded
	}
	return curve.NewPrivateKey(d)
}

func ecdhPublicToPEM(pub *ecdh.PublicKey) ([]byte, error) {
	pubBytes := pub.Bytes()
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	ecdsaPub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	der, err := x509.MarshalPKIXPublicKey(ecdsaPub)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), nil
}

func pemToECDHPublic(pemBytes []byte) (*ecdh.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM public key")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	ecdsaPub, ok := pubInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not ECDSA")
	}
	pubBytes := elliptic.Marshal(ecdsaPub.Curve, ecdsaPub.X, ecdsaPub.Y)
	return ecdh.P256().NewPublicKey(pubBytes)
}

// ECIESEncryptSessionKey encrypts a session key using recipient's public key.
// Returns ephemeral public key PEM, encrypted session key hex, IV hex.
func ECIESEncryptSessionKey(recipientPub *ecdsa.PublicKey, sessionKey []byte) (ephemeralPubPEM []byte, encryptedKeyHex string, ivHex string, err error) {
	recPubECDH, err := ecdsaPublicToECDH(recipientPub)
	if err != nil {
		return nil, "", "", fmt.Errorf("convert recipient public key: %w", err)
	}

	ephemeralPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", "", fmt.Errorf("generate ephemeral key: %w", err)
	}

	sharedSecret, err := ephemeralPriv.ECDH(recPubECDH)
	if err != nil {
		return nil, "", "", fmt.Errorf("derive shared secret: %w", err)
	}

	aesKey := sha256.Sum256(sharedSecret)

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, "", "", err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, "", "", fmt.Errorf("generate IV: %w", err)
	}

	padded := pkcs7Padding(sessionKey, aes.BlockSize)
	encryptedKey := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedKey, padded)

	pubPEM, err := ecdhPublicToPEM(ephemeralPriv.PublicKey())
	if err != nil {
		return nil, "", "", fmt.Errorf("encode ephemeral public key: %w", err)
	}

	return pubPEM, hex.EncodeToString(encryptedKey), hex.EncodeToString(iv), nil
}

// ECIESDecryptSessionKey decrypts a session key using recipient's private key.
func ECIESDecryptSessionKey(recipientPriv *ecdsa.PrivateKey, ephemeralPubPEM []byte, encryptedKeyHex string, ivHex string) ([]byte, error) {
	recPrivECDH, err := ecdsaPrivateToECDH(recipientPriv)
	if err != nil {
		return nil, fmt.Errorf("convert recipient private key: %w", err)
	}

	ephemeralPub, err := pemToECDHPublic(ephemeralPubPEM)
	if err != nil {
		return nil, err
	}

	sharedSecret, err := recPrivECDH.ECDH(ephemeralPub)
	if err != nil {
		return nil, fmt.Errorf("derive shared secret: %w", err)
	}

	aesKey := sha256.Sum256(sharedSecret)

	encryptedKey, err := hex.DecodeString(encryptedKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted key: %w", err)
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return nil, fmt.Errorf("decode IV: %w", err)
	}

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(encryptedKey))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, encryptedKey)
	return pkcs7Unpadding(decrypted)
}

func pkcs7Unpadding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding == 0 || unpadding > aes.BlockSize || unpadding > length {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := 0; i < unpadding; i++ {
		if data[length-1-i] != byte(unpadding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:length-unpadding], nil
}
