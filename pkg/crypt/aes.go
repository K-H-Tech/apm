package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// AESCrypt will implement Crypt interface for AES algorithm
type AESCrypt struct {
	secretKey []byte
}

// NewAESCrypt is the AESCrypt factory method
func NewAESCrypt(secretKey []byte) *AESCrypt {
	return &AESCrypt{secretKey: secretKey}
}

// Encrypt will encrypt the array of input bytes using aes algorithm
//
//	output string is base64 encoded
func (a *AESCrypt) Encrypt(msg []byte) (string, error) {
	block, err := aes.NewCipher(a.secretKey)
	if err != nil {
		return "", fmt.Errorf("AESCrypt: could not create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("AESCrypt: could not create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("AESCrypt: could not create nonce: %v", err)
	}

	cipherText := gcm.Seal(nonce, nonce, msg, nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt will decrypt the array of input bytes using aes algorithm
//
// input msg must be base64 encoded
func (a *AESCrypt) Decrypt(msg string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return "", fmt.Errorf("AESCrypt: could not base64 decode: %v", err)
	}

	block, err := aes.NewCipher(a.secretKey)
	if err != nil {
		return "", fmt.Errorf("AESCrypt: could not create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("AESCrypt: could not create GCM: %v", err)
	}

	if len(cipherText) < gcm.NonceSize() {
		return "", fmt.Errorf("AESCrypt: cipherText too short")
	}

	nonce := cipherText[:gcm.NonceSize()]
	cipherText = cipherText[gcm.NonceSize():]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("AESCrypt: decryption failed: %v", err)
	}

	return string(plainText), nil
}
