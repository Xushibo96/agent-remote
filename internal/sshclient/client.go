package sshclient

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type KnownHostsPolicy string

const (
	KnownHostsStrict    KnownHostsPolicy = "strict"
	KnownHostsAcceptNew  KnownHostsPolicy = "accept-new"
	KnownHostsInsecure   KnownHostsPolicy = "insecure"
)

type AuthConfig struct {
	User             string
	Password         string
	PrivateKeyPath   string
	KnownHostsPolicy string
	KnownHostsPath   string
	Timeout          time.Duration
}

func BuildClientConfig(cfg AuthConfig) (*ssh.ClientConfig, error) {
	if cfg.User == "" {
		return nil, fmt.Errorf("user is required")
	}
	authMethods, err := authMethods(cfg.Password, cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	hostKeyCallback, err := hostKeyCallback(cfg.KnownHostsPolicy, cfg.KnownHostsPath)
	if err != nil {
		return nil, err
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}, nil
}

func authMethods(password, privateKeyPath string) ([]ssh.AuthMethod, error) {
	methods := make([]ssh.AuthMethod, 0, 2)
	if password != "" {
		methods = append(methods, ssh.Password(password))
	}
	if privateKeyPath != "" {
		key, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("at least one auth method is required")
	}
	return methods, nil
}

func hostKeyCallback(policy string, knownHostsPath string) (ssh.HostKeyCallback, error) {
	switch KnownHostsPolicy(policy) {
	case "", KnownHostsStrict:
		if knownHostsPath == "" {
			return nil, fmt.Errorf("known_hosts path is required for strict policy")
		}
		return knownhosts.New(knownHostsPath)
	case KnownHostsAcceptNew:
		if knownHostsPath == "" {
			return ssh.InsecureIgnoreHostKey(), nil
		}
		callback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, err
		}
		return callback, nil
	case KnownHostsInsecure:
		return ssh.InsecureIgnoreHostKey(), nil
	default:
		return nil, fmt.Errorf("unsupported known_hosts policy %q", policy)
	}
}
