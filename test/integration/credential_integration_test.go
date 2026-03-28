package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-remote/internal/credential"
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

func TestCredentialConfigDoesNotPersistPlaintext(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	vault := credential.NewVault(
		store.NewFileConfigStore(cfgPath),
		secret.NewKeyManager(&memoryKeyring{values: map[string]string{}}, "", ""),
	)

	_, err := vault.SaveTarget(context.Background(), model.AddTargetRequest{
		ID:       "prod",
		Host:     "example.com",
		User:     "root",
		Password: "super-secret",
	})
	if err != nil {
		t.Fatalf("SaveTarget() error = %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), "super-secret") {
		t.Fatal("config file should not contain plaintext password")
	}
}

