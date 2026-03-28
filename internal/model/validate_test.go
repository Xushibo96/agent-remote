package model

import "testing"

func TestNormalizeAddTargetRequestDefaults(t *testing.T) {
	req, err := NormalizeAddTargetRequest(AddTargetRequest{
		Name:     "Prod Host",
		Host:     "example.com",
		User:     "root",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("NormalizeAddTargetRequest() error = %v", err)
	}
	if req.ID != "prod-host" {
		t.Fatalf("expected generated id prod-host, got %q", req.ID)
	}
	if req.Port != 22 {
		t.Fatalf("expected default port 22, got %d", req.Port)
	}
	if req.AuthMode != AuthModePassword {
		t.Fatalf("expected default auth mode password, got %q", req.AuthMode)
	}
	if req.KnownHostsPolicy != KnownHostsStrict {
		t.Fatalf("expected strict known hosts, got %q", req.KnownHostsPolicy)
	}
}

func TestNormalizeAddTargetRequestKeyModeRequiresKey(t *testing.T) {
	_, err := NormalizeAddTargetRequest(AddTargetRequest{
		Host:     "example.com",
		User:     "root",
		AuthMode: AuthModeKey,
	})
	if err == nil {
		t.Fatal("expected error for missing private key path")
	}
}

func TestNormalizeSyncRequestDefaultsOverwrite(t *testing.T) {
	req, err := NormalizeSyncRequest(SyncRequest{
		TargetID:  "prod",
		Direction: DirectionUpload,
		LocalPath: "./local",
		RemotePath: "/srv/app",
	})
	if err != nil {
		t.Fatalf("NormalizeSyncRequest() error = %v", err)
	}
	if req.ConflictPolicy != ConflictOverwrite {
		t.Fatalf("expected overwrite by default, got %q", req.ConflictPolicy)
	}
	if req.BackendPreference != BackendAuto {
		t.Fatalf("expected auto backend by default, got %q", req.BackendPreference)
	}
}

func TestNormalizeSyncRequestBidirRequiresBothPaths(t *testing.T) {
	_, err := NormalizeSyncRequest(SyncRequest{
		TargetID:  "prod",
		Direction: DirectionBidir,
		LocalPath: "./local",
	})
	if err == nil {
		t.Fatal("expected error for missing remote path")
	}
}

func TestNormalizeExecRequestDefaultsTimeout(t *testing.T) {
	req, err := NormalizeExecRequest(ExecRequest{
		TargetID: "prod",
		Command:  "ls -la",
	})
	if err != nil {
		t.Fatalf("NormalizeExecRequest() error = %v", err)
	}
	if req.TimeoutSeconds != 600 {
		t.Fatalf("expected default timeout 600, got %d", req.TimeoutSeconds)
	}
}
