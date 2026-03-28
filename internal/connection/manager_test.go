package connection

import (
	"context"
	"testing"

	"agent-remote/internal/model"
	"golang.org/x/crypto/ssh"
)

func TestManagerReusesClient(t *testing.T) {
	client := &ssh.Client{}
	m := NewManagerWithDeps(func(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return client, nil
	}, func(ctx context.Context, client *ssh.Client) (model.RemoteCapabilities, error) {
		return model.RemoteCapabilities{SSHAvailable: true, SFTPAvailable: true}, nil
	})

	target := model.ResolvedTarget{
		TargetConfig: model.TargetConfig{
			ID:               "prod",
			Host:             "example.com",
			Port:             22,
			User:             "root",
			AuthMode:         model.AuthModePassword,
			KnownHostsPolicy: model.KnownHostsInsecure,
		},
		Password: "secret",
	}

	first, err := m.GetSSHClient(context.Background(), target)
	if err != nil {
		t.Fatalf("GetSSHClient() error = %v", err)
	}
	second, err := m.GetSSHClient(context.Background(), target)
	if err != nil {
		t.Fatalf("GetSSHClient() error = %v", err)
	}
	if first != second {
		t.Fatal("expected connection reuse")
	}
}

func TestManagerDetectCapabilities(t *testing.T) {
	m := NewManagerWithDeps(func(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return &ssh.Client{}, nil
	}, func(ctx context.Context, client *ssh.Client) (model.RemoteCapabilities, error) {
		return model.RemoteCapabilities{
			SSHAvailable:   true,
			SFTPAvailable:  true,
			RsyncAvailable: true,
			ShellPath:      "/bin/bash",
			OS:             "linux",
		}, nil
	})

	target := model.ResolvedTarget{
		TargetConfig: model.TargetConfig{
			ID:               "prod",
			Host:             "example.com",
			Port:             22,
			User:             "root",
			AuthMode:         model.AuthModePassword,
			KnownHostsPolicy: model.KnownHostsInsecure,
		},
		Password: "secret",
	}

	caps, err := m.DetectCapabilities(context.Background(), target)
	if err != nil {
		t.Fatalf("DetectCapabilities() error = %v", err)
	}
	if !caps.RsyncAvailable || !caps.SFTPAvailable || !caps.SSHAvailable {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
}

func TestCloseIdle(t *testing.T) {
	client := &ssh.Client{}
	m := NewManagerWithDeps(func(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return client, nil
	}, nil)

	target := model.ResolvedTarget{
		TargetConfig: model.TargetConfig{
			ID:               "prod",
			Host:             "example.com",
			Port:             22,
			User:             "root",
			AuthMode:         model.AuthModePassword,
			KnownHostsPolicy: model.KnownHostsInsecure,
		},
		Password: "secret",
	}

	if _, err := m.GetSSHClient(context.Background(), target); err != nil {
		t.Fatalf("GetSSHClient() error = %v", err)
	}
	if err := m.CloseIdle(); err != nil {
		t.Fatalf("CloseIdle() error = %v", err)
	}
}
