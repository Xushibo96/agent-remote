package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"agent-remote/internal/model"
)

func TestFileConfigStoreSaveAndList(t *testing.T) {
	store := NewFileConfigStore(filepath.Join(t.TempDir(), "config.json"))
	now := time.Now().UTC()
	target := model.TargetConfig{
		ID:        "prod",
		Host:      "example.com",
		User:      "root",
		Port:      22,
		AuthMode:  model.AuthModePassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.SaveTarget(context.Background(), target); err != nil {
		t.Fatalf("SaveTarget() error = %v", err)
	}

	targets, err := store.ListTargets(context.Background())
	if err != nil {
		t.Fatalf("ListTargets() error = %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
}
