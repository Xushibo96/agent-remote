package budget

import (
	"testing"
	"time"

	"agent-remote/internal/model"
)

func TestBudgetExecKeepsRecentWindowAndErrors(t *testing.T) {
	events := []model.ExecEvent{
		{Seq: 1, Type: "started", Stream: "system", Payload: "start", Timestamp: time.Now()},
		{Seq: 2, Type: "stdout", Stream: "stdout", Payload: "line-1", Timestamp: time.Now()},
		{Seq: 3, Type: "stderr", Stream: "stderr", Payload: "err-1", Timestamp: time.Now()},
		{Seq: 4, Type: "stdout", Stream: "stdout", Payload: "line-2", Timestamp: time.Now()},
	}

	result := New().BudgetExec(events, model.BudgetPolicy{
		MaxLines:      2,
		MaxBytes:      64,
		WindowBytes:   32,
		KeepErrors:    true,
		KeepLifecycle: true,
	}, "", model.JobSummary{ID: "sess-1"})

	if len(result.Events) == 0 {
		t.Fatal("expected some events")
	}
	if !result.Truncated {
		t.Fatal("expected truncation")
	}
	if result.Cursor == "" {
		t.Fatal("expected cursor")
	}
}

func TestBudgetSyncPassThrough(t *testing.T) {
	summary := model.JobSummary{ID: "job-1"}
	result := New().BudgetSync(model.SyncRunResult{
		ID:              "job-1",
		EffectiveBackend: "rsync",
		Summary:         summary,
	}, model.BudgetPolicy{})

	if result.ID != "job-1" || result.EffectiveBackend != "rsync" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

