package exec

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"agent-remote/internal/interfaces"
	"agent-remote/internal/model"
	"agent-remote/internal/session"
	"golang.org/x/crypto/ssh"
)

const defaultBufferCapacity = 256

type Runner interface {
	Run(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget, store *session.Store) (model.ExecStartResult, error)
}

type Engine struct {
	connections interfaces.ConnectionManager
	sessions    *session.Store
	budgeter    interfaces.OutputBudgeter
	runner      Runner

	mu     sync.Mutex
	active map[string]context.CancelFunc
}

func NewEngine(connections interfaces.ConnectionManager, sessions *session.Store, budgeter interfaces.OutputBudgeter) *Engine {
	e := &Engine{
		connections: connections,
		sessions:    sessions,
		budgeter:    budgeter,
		active:      make(map[string]context.CancelFunc),
	}
	e.runner = &sshRunner{connections: connections}
	return e
}

func NewEngineWithRunner(connections interfaces.ConnectionManager, sessions *session.Store, budgeter interfaces.OutputBudgeter, runner Runner) *Engine {
	e := NewEngine(connections, sessions, budgeter)
	if runner != nil {
		e.runner = runner
	}
	return e
}

func (e *Engine) Start(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget) (model.ExecStartResult, error) {
	normalized, err := model.NormalizeExecRequest(req)
	if err != nil {
		return model.ExecStartResult{}, err
	}
	if normalized.SessionID == "" {
		normalized.SessionID = newSessionID()
	}

	summary := model.JobSummary{
		ID:        normalized.SessionID,
		Kind:      "exec",
		State:     "running",
		StartedAt: time.Now().UTC(),
	}
	e.sessions.CreateSession(summary, defaultBufferCapacity)

	runCtx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.active[normalized.SessionID] = cancel
	e.mu.Unlock()

	startResult, err := e.runner.Run(runCtx, normalized, target, e.sessions)
	e.mu.Lock()
	delete(e.active, normalized.SessionID)
	e.mu.Unlock()
	cancel()
	if err != nil {
		e.finishSession(normalized.SessionID, "failed", nil, "exec_start", true)
		return model.ExecStartResult{}, err
	}

	if startResult.ID == "" {
		startResult.ID = normalized.SessionID
	}
	if startResult.Summary.ID == "" {
		startResult.Summary = summary
	}
	if startResult.Cursor == "" {
		startResult.Cursor = currentCursor(e.sessions, normalized.SessionID)
	}
	return startResult, nil
}

func (e *Engine) Read(ctx context.Context, req model.ExecReadRequest) (model.ExecReadResult, error) {
	_ = ctx
	if req.SessionID == "" {
		return model.ExecReadResult{}, fmt.Errorf("session_id is required")
	}
	events, cursor, truncated, summary, err := e.sessions.ReadEvents(req.SessionID, req.Cursor)
	if err != nil {
		return model.ExecReadResult{}, err
	}
	if e.budgeter == nil {
		return model.ExecReadResult{
			ID:        req.SessionID,
			Events:    events,
			Cursor:    cursor,
			Truncated: truncated,
			Summary:   summary,
		}, nil
	}
	return e.budgeter.BudgetExec(events, req.Budget, cursor, summary), nil
}

func (e *Engine) Stop(ctx context.Context, req model.ExecStopRequest) (model.ExecStopResult, error) {
	_ = ctx
	if req.SessionID == "" {
		return model.ExecStopResult{}, fmt.Errorf("session_id is required")
	}

	summary, err := e.sessions.GetJob(context.Background(), req.SessionID)
	if err != nil {
		return model.ExecStopResult{}, err
	}
	if isTerminal(summary.State) {
		return model.ExecStopResult{ID: req.SessionID, Summary: summary}, nil
	}

	e.mu.Lock()
	cancel, ok := e.active[req.SessionID]
	if ok {
		delete(e.active, req.SessionID)
	}
	e.mu.Unlock()
	if ok && cancel != nil {
		cancel()
	}

	finished := time.Now().UTC()
	summary.State = "stopped"
	summary.FinishedAt = &finished
	if err := e.sessions.UpdateSummary(summary); err != nil {
		return model.ExecStopResult{}, err
	}
	return model.ExecStopResult{ID: req.SessionID, Summary: summary}, nil
}

func (e *Engine) finishSession(sessionID, state string, exitCode *int, failureStage string, retryable bool) {
	summary, err := e.sessions.GetJob(context.Background(), sessionID)
	if err != nil {
		return
	}
	finished := time.Now().UTC()
	summary.State = state
	summary.FinishedAt = &finished
	summary.ExitCode = exitCode
	summary.FailureStage = failureStage
	summary.Retryable = retryable
	_ = e.sessions.UpdateSummary(summary)
}

func isTerminal(state string) bool {
	switch strings.ToLower(state) {
	case "completed", "failed", "stopped", "interrupted":
		return true
	default:
		return false
	}
}

func newSessionID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("exec-%d", time.Now().UnixNano())
	}
	return "exec-" + hex.EncodeToString(buf[:])
}

type sshRunner struct {
	connections interfaces.ConnectionManager
}

func (r *sshRunner) Run(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget, store *session.Store) (model.ExecStartResult, error) {
	if r.connections == nil {
		return model.ExecStartResult{}, fmt.Errorf("connection manager is required")
	}
	client, err := r.connections.GetSSHClient(ctx, target)
	if err != nil {
		return model.ExecStartResult{}, err
	}

	sshSession, err := client.NewSession()
	if err != nil {
		return model.ExecStartResult{}, err
	}
	defer func() { _ = sshSession.Close() }()

	if req.WorkDir != "" {
		_ = sshSession.Setenv("PWD", req.WorkDir)
	}
	for key, value := range req.Env {
		_ = sshSession.Setenv(key, value)
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		return model.ExecStartResult{}, err
	}
	stderr, err := sshSession.StderrPipe()
	if err != nil {
		return model.ExecStartResult{}, err
	}

	if _, err := store.AppendEvent(req.SessionID, model.ExecEvent{
		SessionID: req.SessionID,
		Type:      "started",
		Stream:    "system",
		Payload:   req.Command,
	}); err != nil {
		return model.ExecStartResult{}, err
	}

	if req.PTY {
		_ = sshSession.RequestPty("xterm", 80, 40, ssh.TerminalModes{})
	}

	if err := sshSession.Start(req.Command); err != nil {
		return model.ExecStartResult{}, err
	}

	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})
	go func() {
		defer close(stdoutDone)
		pumpStream(ctx, store, req.SessionID, "stdout", stdout)
	}()
	go func() {
		defer close(stderrDone)
		pumpStream(ctx, store, req.SessionID, "stderr", stderr)
	}()

	waitErr := sshSession.Wait()
	<-stdoutDone
	<-stderrDone

	state := "completed"
	var exitCode *int
	failureStage := ""
	retryable := false
	if waitErr != nil {
		state = "failed"
		failureStage = "exec_wait"
		retryable = true
		if exitErr, ok := waitErr.(*ssh.ExitError); ok {
			code := exitErr.ExitStatus()
			exitCode = &code
		}
	}
	finished := time.Now().UTC()
	summary, err := store.GetJob(context.Background(), req.SessionID)
	if err != nil {
		return model.ExecStartResult{}, err
	}
	summary.State = state
	summary.FinishedAt = &finished
	summary.ExitCode = exitCode
	summary.FailureStage = failureStage
	summary.Retryable = retryable
	if err := store.UpdateSummary(summary); err != nil {
		return model.ExecStartResult{}, err
	}

	return model.ExecStartResult{
		ID:      req.SessionID,
		Summary: summary,
		Cursor:  currentCursor(store, req.SessionID),
	}, nil
}

func pumpStream(ctx context.Context, store *session.Store, sessionID, stream string, r io.Reader) {
	if r == nil {
		return
	}
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			_, _ = store.AppendEvent(sessionID, model.ExecEvent{
				SessionID: sessionID,
				Type:      stream,
				Stream:    stream,
				Payload:   string(buf[:n]),
			})
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				_, _ = store.AppendEvent(sessionID, model.ExecEvent{
					SessionID: sessionID,
					Type:      "failed",
					Stream:    "system",
					Payload:   err.Error(),
				})
			}
			return
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func currentCursor(store *session.Store, sessionID string) string {
	if store == nil || sessionID == "" {
		return ""
	}
	_, cursor, _, _, err := store.ReadEvents(sessionID, "")
	if err != nil {
		return ""
	}
	return cursor
}
