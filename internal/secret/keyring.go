package secret

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	DefaultServiceName = "agent-remote"
	DefaultKeyID       = "master-key"
)

var ErrNotFound = errors.New("secret not found")

type Keyring interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

type OSKeyring struct{}

func (OSKeyring) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (OSKeyring) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

func (OSKeyring) Delete(service, user string) error {
	return keyring.Delete(service, user)
}

type KeyManager struct {
	keyring Keyring
	service string
	keyID   string
}

func NewKeyManager(k Keyring, service, keyID string) *KeyManager {
	if service == "" {
		service = DefaultServiceName
	}
	if keyID == "" {
		keyID = DefaultKeyID
	}
	return &KeyManager{keyring: k, service: service, keyID: keyID}
}

func (m *KeyManager) LoadOrCreateMasterKey() ([]byte, string, error) {
	if m == nil || m.keyring == nil {
		return nil, "", fmt.Errorf("key manager is not configured")
	}

	encoded, err := m.keyring.Get(m.service, m.keyID)
	if err == nil {
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, "", fmt.Errorf("decode master key: %w", err)
		}
		return key, m.keyID, nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, "", err
	}
	if err := m.keyring.Set(m.service, m.keyID, base64.StdEncoding.EncodeToString(key)); err != nil {
		return nil, "", err
	}
	return key, m.keyID, nil
}
