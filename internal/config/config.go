package config

import (
	"os"
	"path/filepath"
)

const (
	defaultAppDirName = "agent-remote"
	configFileName    = "config.json"
)

type Config struct {
	ConfigDir  string
	ConfigFile string
}

func Load(baseDir string) (Config, error) {
	if baseDir == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return Config{}, err
		}
		baseDir = filepath.Join(userConfigDir, defaultAppDirName)
	}

	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return Config{}, err
	}

	return Config{
		ConfigDir:  baseDir,
		ConfigFile: filepath.Join(baseDir, configFileName),
	}, nil
}
