package config

import "testing"

func TestLoadCreatesConfigDir(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ConfigDir == "" || cfg.ConfigFile == "" {
		t.Fatal("expected config paths to be populated")
	}
}
