package model

import "time"

type SecretEnvelope struct {
	Version    int       `json:"version"`
	Algorithm  string    `json:"algorithm"`
	KeyID      string    `json:"key_id"`
	Ciphertext string    `json:"ciphertext"`
	Nonce      string    `json:"nonce"`
	CreatedAt  time.Time `json:"created_at"`
}
