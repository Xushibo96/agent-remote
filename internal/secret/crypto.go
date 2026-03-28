package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"agent-remote/internal/model"
)

const (
	encryptionAlgorithm = "aes-256-gcm"
)

func Encrypt(key []byte, plaintext string, keyID string) (*model.SecretEnvelope, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)
	return &model.SecretEnvelope{
		Version:    1,
		Algorithm:  encryptionAlgorithm,
		KeyID:      keyID,
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func Decrypt(key []byte, envelope *model.SecretEnvelope) (string, error) {
	if envelope == nil {
		return "", fmt.Errorf("secret envelope is required")
	}
	if envelope.Algorithm != encryptionAlgorithm {
		return "", fmt.Errorf("unsupported algorithm %q", envelope.Algorithm)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return "", err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
