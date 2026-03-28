package main

import (
	"context"
	"encoding/json"
	"testing"

	"agent-remote/internal/app"
	"agent-remote/internal/budget"
	"agent-remote/internal/connection"
	"agent-remote/internal/credential"
	execengine "agent-remote/internal/exec"
	"agent-remote/internal/mcp"
	"agent-remote/internal/model"
	"agent-remote/internal/rsync"
	"agent-remote/internal/secret"
	"agent-remote/internal/session"
	"agent-remote/internal/store"
	syncengine "agent-remote/internal/sync"
)

func TestServerHandleToolDispatch(t *testing.T) {
	vault, jobs := testService(t)
	orchestrator := app.NewOrchestrator(
		vault,
		jobs,
		syncengine.NewEngine(syncengine.NewRootedFilesystem(t.TempDir(), nil), syncengine.NewRootedFilesystem(t.TempDir(), nil), rsync.NewRunner("")),
		execengine.NewEngine(connection.NewManagerWithDeps(nil, nil), jobs, budget.New()),
	)
	srv := mcp.NewServer(mcp.NewAdapter(orchestrator))

	addPayload := mustJSON(t, model.AddTargetRequest{
		ID:       "prod",
		Host:     "example.com",
		User:     "root",
		Password: "secret",
	})
	resp, err := srv.Handle(context.Background(), "target_add", addPayload)
	if err != nil {
		t.Fatalf("Handle(target_add) error = %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected target_add ok response, got %#v", resp)
	}

	resp, err = srv.Handle(context.Background(), "target_list", nil)
	if err != nil {
		t.Fatalf("Handle(target_list) error = %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected target_list ok response, got %#v", resp)
	}

	jobs.CreateSession(model.JobSummary{ID: "job-1", Kind: "exec", State: "running"}, 4)
	jobPayload := mustJSON(t, model.JobStatusRequest{ID: "job-1"})
	resp, err = srv.Handle(context.Background(), "job_status", jobPayload)
	if err != nil {
		t.Fatalf("Handle(job_status) error = %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected job_status ok response, got %#v", resp)
	}

}

func TestAdapterNotImplementedResponse(t *testing.T) {
	adapter := mcp.NewAdapter(nil)
	resp, err := adapter.StartExec(context.Background(), model.ExecRequest{
		TargetID: "prod",
		Command:  "echo hi",
	})
	if err != nil {
		t.Fatalf("StartExec() returned unexpected error: %v", err)
	}
	if resp.OK {
		t.Fatal("expected not implemented response to be non-ok")
	}
	if resp.State != "not_implemented" {
		t.Fatalf("expected not_implemented state, got %q", resp.State)
	}
}

func testService(t *testing.T) (*credential.Vault, *session.Store) {
	t.Helper()
	keyring := &memoryKeyring{values: map[string]string{}}
	km := secret.NewKeyManager(keyring, "", "")
	cfgStore := store.NewFileConfigStore(t.TempDir() + "/config.json")
	vault := credential.NewVault(cfgStore, km)
	jobs := session.NewStore()
	return vault, jobs
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}

type memoryKeyring struct {
	values map[string]string
}

func (m *memoryKeyring) Get(service, user string) (string, error) {
	if value, ok := m.values[service+":"+user]; ok {
		return value, nil
	}
	return "", secret.ErrNotFound
}

func (m *memoryKeyring) Set(service, user, password string) error {
	m.values[service+":"+user] = password
	return nil
}

func (m *memoryKeyring) Delete(service, user string) error {
	delete(m.values, service+":"+user)
	return nil
}

func TestJSONResponseEncodes(t *testing.T) {
	data, err := json.Marshal(mcp.Response{OK: true, State: "ready"})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty json")
	}
}
