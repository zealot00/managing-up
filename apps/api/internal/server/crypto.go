package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var encryptionKey []byte

func init() {
	keyStr := os.Getenv("GATEWAY_ENCRYPTION_KEY")
	if keyStr == "" {
		return
	}
	encryptionKey, _ = base64.StdEncoding.DecodeString(keyStr)
	if len(encryptionKey) != 32 {
		encryptionKey = nil
	}
}

func getEncryptionKey() ([]byte, error) {
	if encryptionKey == nil {
		return nil, errors.New("GATEWAY_ENCRYPTION_KEY not set or invalid (must be 32-byte base64)")
	}
	return encryptionKey, nil
}

func EncryptAPIKey(plaintext string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return plaintext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptAPIKey(encrypted string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return encrypted, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
