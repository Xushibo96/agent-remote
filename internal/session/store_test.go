package session

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"agent-remote/internal/model"
)

func TestStorePutGetJobAndReadEvents(t *testing.T) {
	store := NewStore()
	summary := model.JobSummary{
		ID:        "job-1",
		Kind:      "exec",
		State:     "running",
		StartedAt: time.Now().UTC(),
	}

	if err := store.PutJob(context.Background(), summary); err != nil {
		t.Fatalf("PutJob() error = %v", err)
	}
	if got, err := store.GetJob(context.Background(), "job-1"); err != nil || got.ID != "job-1" {
		t.Fatalf("GetJob() got=%v err=%v", got, err)
	}

	store.CreateSession(summary, 4)
	if _, err := store.AppendEvent("job-1", model.ExecEvent{Type: "started"}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}
	if _, err := store.AppendEvent("job-1", model.ExecEvent{Type: "stdout", Payload: "hello"}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	events, cursor, truncated, gotSummary, err := store.ReadEvents("job-1", "")
	if err != nil {
		t.Fatalf("ReadEvents() error = %v", err)
	}
	if truncated {
		t.Fatal("did not expect truncation")
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if cursor == "" {
		t.Fatal("expected cursor")
	}
	if gotSummary.ID != "job-1" {
		t.Fatalf("expected summary id job-1, got %q", gotSummary.ID)
	}
}

func TestStorePersistsAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")
	store, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatalf("NewStoreWithPath() error = %v", err)
	}
	summary := model.JobSummary{
		ID:        "job-persist",
		Kind:      "exec",
		State:     "running",
		StartedAt: time.Now().UTC(),
	}
	store.CreateSession(summary, 4)
	if _, err := store.AppendEvent("job-persist", model.ExecEvent{Type: "stdout", Payload: "hello"}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	reloaded, err := NewStoreWithPath(path)
	if err != nil {
		t.Fatalf("NewStoreWithPath(reload) error = %v", err)
	}
	events, _, _, gotSummary, err := reloaded.ReadEvents("job-persist", "")
	if err != nil {
		t.Fatalf("ReadEvents() error = %v", err)
	}
	if gotSummary.ID != "job-persist" || len(events) != 1 {
		t.Fatalf("unexpected reload state summary=%+v events=%d", gotSummary, len(events))
	}
}
