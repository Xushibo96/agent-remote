package sshclient

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildClientConfigRequiresUser(t *testing.T) {
	_, err := BuildClientConfig(AuthConfig{})
	if err == nil {
		t.Fatal("expected user validation error")
	}
}

func TestBuildClientConfigInsecurePolicy(t *testing.T) {
	cfg, err := BuildClientConfig(AuthConfig{
		User:             "root",
		Password:         "secret",
		KnownHostsPolicy: string(KnownHostsInsecure),
	})
	if err != nil {
		t.Fatalf("BuildClientConfig() error = %v", err)
	}
	if cfg.HostKeyCallback == nil {
		t.Fatal("expected host key callback")
	}
	if cfg.Timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %s", cfg.Timeout)
	}
}

func TestBuildClientConfigStrictPolicyRequiresKnownHosts(t *testing.T) {
	_, err := BuildClientConfig(AuthConfig{
		User:             "root",
		Password:         "secret",
		KnownHostsPolicy: string(KnownHostsStrict),
	})
	if err == nil {
		t.Fatal("expected known_hosts validation error")
	}
}

func TestBuildClientConfigPrivateKeyPathLoadsSigner(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_rsa")
	if err := os.WriteFile(keyPath, []byte("not-a-real-key"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := BuildClientConfig(AuthConfig{
		User:             "root",
		PrivateKeyPath:   keyPath,
		KnownHostsPolicy: string(KnownHostsInsecure),
	})
	if err == nil {
		t.Fatal("expected parse error for invalid private key")
	}
}
