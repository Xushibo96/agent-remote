package mcp

import (
	"context"
	"fmt"

	"agent-remote/internal/model"
)

type Service interface {
	AddTarget(ctx context.Context, req model.AddTargetRequest) (model.TargetConfig, error)
	ListTargets(ctx context.Context) ([]model.TargetConfig, error)
	StartSync(ctx context.Context, req model.SyncRequest) (model.JobResult, error)
	StartExec(ctx context.Context, req model.ExecRequest) (model.ExecStartResult, error)
	ReadExec(ctx context.Context, req model.ExecReadRequest) (model.ExecReadResult, error)
	StopExec(ctx context.Context, req model.ExecStopRequest) (model.ExecStopResult, error)
	GetJob(ctx context.Context, req model.JobStatusRequest) (model.JobResult, error)
}

type Adapter struct {
	service Service
}

func NewAdapter(service Service) *Adapter {
	return &Adapter{service: service}
}

func (a *Adapter) AddTarget(ctx context.Context, req model.AddTargetRequest) (Response, error) {
	normalized, err := model.NormalizeAddTargetRequest(req)
	if err != nil {
		return Response{}, err
	}
	if a.service == nil {
		return Response{OK: false, State: "not_implemented", Error: errorInfo("not_implemented", "adapter", "add_target", "service is not configured")}, nil
	}
	target, err := a.service.AddTarget(ctx, normalized)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, ID: target.ID, State: "ready", Summary: map[string]any{"target": target}}, nil
}

func (a *Adapter) ListTargets(ctx context.Context) (Response, error) {
	if a.service == nil {
		return Response{OK: true, State: "not_implemented", Summary: map[string]any{"targets": []model.TargetConfig{}}}, nil
	}
	targets, err := a.service.ListTargets(ctx)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, State: "ready", Summary: map[string]any{"targets": targets}}, nil
}

func (a *Adapter) StartSync(ctx context.Context, req model.SyncRequest) (Response, error) {
	normalized, err := model.NormalizeSyncRequest(req)
	if err != nil {
		return Response{}, err
	}
	if a.service == nil {
		return Response{OK: false, ID: normalized.JobID, State: "not_implemented", Error: errorInfo("not_implemented", "adapter", "start_sync", "service is not configured")}, nil
	}
	result, err := a.service.StartSync(ctx, normalized)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, ID: result.ID, State: result.Summary.State, Summary: result.Summary}, nil
}

func (a *Adapter) StartExec(ctx context.Context, req model.ExecRequest) (Response, error) {
	normalized, err := model.NormalizeExecRequest(req)
	if err != nil {
		return Response{}, err
	}
	if a.service == nil {
		return Response{OK: false, ID: normalized.SessionID, State: "not_implemented", Error: errorInfo("not_implemented", "adapter", "start_exec", "service is not configured")}, nil
	}
	result, err := a.service.StartExec(ctx, normalized)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, ID: result.ID, State: result.Summary.State, Summary: result.Summary, Cursor: result.Cursor}, nil
}

func (a *Adapter) ReadExec(ctx context.Context, req model.ExecReadRequest) (Response, error) {
	if req.SessionID == "" {
		return Response{}, fmt.Errorf("session_id is required")
	}
	if a.service == nil {
		return Response{OK: false, ID: req.SessionID, State: "not_implemented", Error: errorInfo("not_implemented", "adapter", "read_exec", "service is not configured")}, nil
	}
	result, err := a.service.ReadExec(ctx, req)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, ID: result.ID, State: result.Summary.State, Summary: result.Summary, Events: result.Events, Cursor: result.Cursor, Truncated: result.Truncated}, nil
}

func (a *Adapter) StopExec(ctx context.Context, req model.ExecStopRequest) (Response, error) {
	if req.SessionID == "" {
		return Response{}, fmt.Errorf("session_id is required")
	}
	if a.service == nil {
		return Response{OK: false, ID: req.SessionID, State: "not_implemented", Error: errorInfo("not_implemented", "adapter", "stop_exec", "service is not configured")}, nil
	}
	result, err := a.service.StopExec(ctx, req)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, ID: result.ID, State: result.Summary.State, Summary: result.Summary}, nil
}

func (a *Adapter) GetJob(ctx context.Context, req model.JobStatusRequest) (Response, error) {
	if req.ID == "" {
		return Response{}, fmt.Errorf("id is required")
	}
	if a.service == nil {
		return Response{OK: false, ID: req.ID, State: "not_implemented", Error: errorInfo("not_implemented", "adapter", "get_job", "service is not configured")}, nil
	}
	result, err := a.service.GetJob(ctx, req)
	if err != nil {
		return Response{}, err
	}
	return Response{OK: true, ID: result.ID, State: result.Summary.State, Summary: result.Summary}, nil
}

func (a *Adapter) ErrorResponse(err error) Response {
	if err == nil {
		return Response{OK: true, State: "ok"}
	}
	return Response{OK: false, State: "error", Error: errorInfo("internal_error", "unknown", "adapter", err.Error())}
}
