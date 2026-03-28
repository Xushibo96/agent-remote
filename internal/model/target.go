package model

import "time"

type TargetConfig struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Host             string          `json:"host"`
	Port             int             `json:"port"`
	User             string          `json:"user"`
	AuthMode         string          `json:"auth_mode"`
	PasswordEnvelope *SecretEnvelope `json:"password_envelope,omitempty"`
	PrivateKeyPath   string          `json:"private_key_path,omitempty"`
	KnownHostsPolicy string          `json:"known_hosts_policy"`
	DefaultBaseDir   string          `json:"default_base_dir,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type AddTargetRequest struct {
	ID               string
	Name             string
	Host             string
	Port             int
	User             string
	AuthMode         string
	Password         string
	PrivateKeyPath   string
	KnownHostsPolicy string
	DefaultBaseDir   string
	Tags             []string
}

type ResolvedTarget struct {
	TargetConfig
	Password string
}
