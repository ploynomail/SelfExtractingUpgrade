package logic

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	paddedPlaintext := pkcs7Padding(plaintext, block.BlockSize())
	ciphertext := make([]byte, aes.BlockSize+len(paddedPlaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate random IV: %w", err)
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedPlaintext)
	return ciphertext, nil
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
