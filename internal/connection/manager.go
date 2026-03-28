package connection

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"agent-remote/internal/model"
	"agent-remote/internal/sshclient"
	"golang.org/x/crypto/ssh"
)

type DialFunc func(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

type probeFunc func(ctx context.Context, client *ssh.Client) (model.RemoteCapabilities, error)

type managerEntry struct {
	client *ssh.Client
	usedAt time.Time
}

type Manager struct {
	mu      sync.Mutex
	clients map[string]*managerEntry
	dial    DialFunc
	probe   probeFunc
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*managerEntry),
		dial:    defaultDial,
		probe:   defaultProbe,
	}
}

func NewManagerWithDeps(dial DialFunc, probe probeFunc) *Manager {
	m := NewManager()
	if dial != nil {
		m.dial = dial
	}
	if probe != nil {
		m.probe = probe
	}
	return m
}

func (m *Manager) GetSSHClient(ctx context.Context, target model.ResolvedTarget) (*ssh.Client, error) {
	key := target.ID
	m.mu.Lock()
	if entry, ok := m.clients[key]; ok && entry.client != nil {
		entry.usedAt = time.Now().UTC()
		client := entry.client
		m.mu.Unlock()
		return client, nil
	}
	m.mu.Unlock()

	clientConfig, err := sshclient.BuildClientConfig(sshclient.AuthConfig{
		User:             target.User,
		Password:         target.Password,
		PrivateKeyPath:   target.PrivateKeyPath,
		KnownHostsPolicy: target.KnownHostsPolicy,
	})
	if err != nil {
		return nil, err
	}

	addr := net.JoinHostPort(target.Host, fmt.Sprintf("%d", target.Port))
	client, err := m.dial(ctx, "tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.clients[key] = &managerEntry{client: client, usedAt: time.Now().UTC()}
	m.mu.Unlock()
	return client, nil
}

func (m *Manager) DetectCapabilities(ctx context.Context, target model.ResolvedTarget) (model.RemoteCapabilities, error) {
	client, err := m.GetSSHClient(ctx, target)
	if err != nil {
		return model.RemoteCapabilities{}, err
	}
	return m.probe(ctx, client)
}

func (m *Manager) CloseIdle() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, entry := range m.clients {
		if entry.client != nil {
			func() {
				defer func() { _ = recover() }()
				_ = entry.client.Close()
			}()
		}
		delete(m.clients, key)
	}
	return nil
}

func defaultDial(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func defaultProbe(ctx context.Context, client *ssh.Client) (model.RemoteCapabilities, error) {
	_ = ctx
	_ = client
	return model.RemoteCapabilities{
		SSHAvailable:   true,
		SFTPAvailable:  true,
		RsyncAvailable: false,
		ShellPath:      "/bin/sh",
		OS:             "unknown",
	}, nil
}
