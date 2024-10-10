package signatureverify

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"
)

// SignFile 签名文件内容
func SignFile(privateKey *ecdsa.PrivateKey, file *os.File) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	fileHash := hash.Sum(nil)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, fileHash)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// VerifyFile 验证文件签名
func VerifyFile(publicKey *ecdsa.PublicKey, file *os.File, signature []byte) (bool, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, err
	}
	fileHash := hash.Sum(nil)
	return ecdsa.VerifyASN1(publicKey, fileHash, signature), nil
}
