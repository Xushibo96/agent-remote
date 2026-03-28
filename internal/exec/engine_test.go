package exec

import (
	"context"
	"errors"
	"testing"
	"time"

	"agent-remote/internal/model"
	"agent-remote/internal/session"
	"golang.org/x/crypto/ssh"
)

type fakeConnections struct {
	err error
}

func (f fakeConnections) GetSSHClient(context.Context, model.ResolvedTarget) (*ssh.Client, error) {
	return nil, f.err
}

func (f fakeConnections) DetectCapabilities(context.Context, model.ResolvedTarget) (model.RemoteCapabilities, error) {
	return model.RemoteCapabilities{SSHAvailable: true}, nil
}

func (f fakeConnections) CloseIdle() error { return nil }

type fakeBudgeter struct{}

func (fakeBudgeter) BudgetSync(result model.SyncRunResult, _ model.BudgetPolicy) model.SyncRunResult {
	return result
}

func (fakeBudgeter) BudgetExec(events []model.ExecEvent, _ model.BudgetPolicy, cursor string, summary model.JobSummary) model.ExecReadResult {
	return model.ExecReadResult{
		ID:        summary.ID,
		Events:    events,
		Cursor:    cursor,
		Summary:   summary,
		Truncated: false,
	}
}

type fakeRunner struct {
	startFn func(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget, store *session.Store) (model.ExecStartResult, error)
}

func (f fakeRunner) Run(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget, store *session.Store) (model.ExecStartResult, error) {
	return f.startFn(ctx, req, target, store)
}

func TestEngineStartReadStopLifecycle(t *testing.T) {
	store := session.NewStore()
	engine := NewEngineWithRunner(fakeConnections{}, store, fakeBudgeter{}, fakeRunner{
		startFn: func(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget, store *session.Store) (model.ExecStartResult, error) {
			if _, err := store.AppendEvent(req.SessionID, model.ExecEvent{SessionID: req.SessionID, Type: "started", Stream: "system"}); err != nil {
				return model.ExecStartResult{}, err
			}
			if _, err := store.AppendEvent(req.SessionID, model.ExecEvent{SessionID: req.SessionID, Type: "stdout", Stream: "stdout", Payload: "hello"}); err != nil {
				return model.ExecStartResult{}, err
			}
			return model.ExecStartResult{ID: req.SessionID, Summary: model.JobSummary{ID: req.SessionID, Kind: "exec", State: "running", StartedAt: time.Now().UTC()}}, nil
		},
	})

	start, err := engine.Start(context.Background(), model.ExecRequest{
		SessionID: "sess-1",
		TargetID:  "target-1",
		Command:   "echo hello",
	}, model.ResolvedTarget{TargetConfig: model.TargetConfig{ID: "target-1"}})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if start.ID != "sess-1" {
		t.Fatalf("expected session id sess-1, got %q", start.ID)
	}

	read, err := engine.Read(context.Background(), model.ExecReadRequest{SessionID: "sess-1"})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(read.Events) == 0 {
		t.Fatal("expected events from read")
	}

	stop, err := engine.Stop(context.Background(), model.ExecStopRequest{SessionID: "sess-1"})
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if stop.Summary.State != "stopped" {
		t.Fatalf("expected stopped state, got %q", stop.Summary.State)
	}

	stopAgain, err := engine.Stop(context.Background(), model.ExecStopRequest{SessionID: "sess-1"})
	if err != nil {
		t.Fatalf("Stop() second call error = %v", err)
	}
	if stopAgain.Summary.State != "stopped" {
		t.Fatalf("expected idempotent stopped state, got %q", stopAgain.Summary.State)
	}
}

func TestEngineStartPropagatesConnectionError(t *testing.T) {
	store := session.NewStore()
	engine := NewEngineWithRunner(fakeConnections{err: errors.New("dial failed")}, store, fakeBudgeter{}, fakeRunner{
		startFn: func(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget, store *session.Store) (model.ExecStartResult, error) {
			return model.ExecStartResult{}, errors.New("should not be called")
		},
	})

	_, err := engine.Start(context.Background(), model.ExecRequest{
		SessionID: "sess-1",
		TargetID:  "target-1",
		Command:   "echo hello",
	}, model.ResolvedTarget{TargetConfig: model.TargetConfig{ID: "target-1"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
