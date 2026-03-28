package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"

	"agent-remote/internal/model"
)

var ErrTargetNotFound = errors.New("target not found")

type FileConfigStore struct {
	path string
}

type fileConfig struct {
	Targets map[string]model.TargetConfig `json:"targets"`
}

func NewFileConfigStore(path string) *FileConfigStore {
	return &FileConfigStore{path: path}
}

func (s *FileConfigStore) SaveTarget(_ context.Context, target model.TargetConfig) error {
	cfg, err := s.read()
	if err != nil {
		return err
	}
	cfg.Targets[target.ID] = target
	return s.write(cfg)
}

func (s *FileConfigStore) GetTarget(_ context.Context, targetID string) (model.TargetConfig, error) {
	cfg, err := s.read()
	if err != nil {
		return model.TargetConfig{}, err
	}
	target, ok := cfg.Targets[targetID]
	if !ok {
		return model.TargetConfig{}, ErrTargetNotFound
	}
	return target, nil
}

func (s *FileConfigStore) ListTargets(_ context.Context) ([]model.TargetConfig, error) {
	cfg, err := s.read()
	if err != nil {
		return nil, err
	}
	targets := make([]model.TargetConfig, 0, len(cfg.Targets))
	for _, target := range cfg.Targets {
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].ID < targets[j].ID
	})
	return targets, nil
}

func (s *FileConfigStore) read() (fileConfig, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return fileConfig{Targets: map[string]model.TargetConfig{}}, nil
	}
	if err != nil {
		return fileConfig{}, err
	}

	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fileConfig{}, err
	}
	if cfg.Targets == nil {
		cfg.Targets = map[string]model.TargetConfig{}
	}
	return cfg, nil
}

func (s *FileConfigStore) write(cfg fileConfig) error {
	if cfg.Targets == nil {
		cfg.Targets = map[string]model.TargetConfig{}
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
