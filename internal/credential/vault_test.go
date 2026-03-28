package credential

import (
	"context"
	"encoding/base64"
	"path/filepath"
	"testing"

	"agent-remote/internal/model"
	"agent-remote/internal/secret"
	"agent-remote/internal/store"
)

type memoryKeyring struct {
	values map[string]string
}

func (m *memoryKeyring) Get(service, user string) (string, error) {
	if value, ok := m.values[service+":"+user]; ok {
		return value, nil
	}
	return "", secret.ErrNotFound
}

func (m *memoryKeyring) Set(service, user, password string) error {
	m.values[service+":"+user] = password
	return nil
}

func (m *memoryKeyring) Delete(service, user string) error {
	delete(m.values, service+":"+user)
	return nil
}

func TestVaultSaveTargetEncryptsPassword(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}

	keyring := &memoryKeyring{values: map[string]string{
		"agent-remote:master-key": base64.StdEncoding.EncodeToString(key),
	}}
	km := secret.NewKeyManager(keyring, "", "")
	cfgStore := store.NewFileConfigStore(filepath.Join(t.TempDir(), "config.json"))
	vault := NewVault(cfgStore, km)

	target, err := vault.SaveTarget(context.Background(), model.AddTargetRequest{
		ID:       "prod",
		Host:     "example.com",
		User:     "root",
		Password: "super-secret",
	})
	if err != nil {
		t.Fatalf("SaveTarget() error = %v", err)
	}
	if target.PasswordEnvelope == nil {
		t.Fatal("expected password envelope to be set")
	}

	resolved, err := vault.LoadTarget(context.Background(), "prod")
	if err != nil {
		t.Fatalf("LoadTarget() error = %v", err)
	}
	if resolved.Password != "super-secret" {
		t.Fatalf("expected decrypted password, got %q", resolved.Password)
	}
}
