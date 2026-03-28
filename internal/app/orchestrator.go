package app

import (
	"context"
	"fmt"

	toolerrors "agent-remote/internal/errors"
	"agent-remote/internal/interfaces"
	"agent-remote/internal/model"
)

type Orchestrator struct {
	vault interfaces.CredentialVault
	jobs  interfaces.SessionStore
	sync  interfaces.SyncEngine
	exec  interfaces.ExecEngine
}

func NewOrchestrator(vault interfaces.CredentialVault, jobs interfaces.SessionStore, syncEngine interfaces.SyncEngine, execEngine interfaces.ExecEngine) *Orchestrator {
	return &Orchestrator{
		vault: vault,
		jobs:  jobs,
		sync:  syncEngine,
		exec:  execEngine,
	}
}

func (o *Orchestrator) AddTarget(ctx context.Context, req model.AddTargetRequest) (model.TargetConfig, error) {
	if o.vault == nil {
		return model.TargetConfig{}, toolerrors.New("vault_missing", "validation", "add_target", "credential vault is not configured")
	}
	return o.vault.SaveTarget(ctx, req)
}

func (o *Orchestrator) ListTargets(ctx context.Context) ([]model.TargetConfig, error) {
	if o.vault == nil {
		return nil, toolerrors.New("vault_missing", "validation", "list_targets", "credential vault is not configured")
	}
	return o.vault.ListTargets(ctx)
}

func (o *Orchestrator) StartSync(ctx context.Context, req model.SyncRequest) (model.JobResult, error) {
	if o.sync == nil {
		return model.JobResult{}, toolerrors.New("sync_missing", "validation", "start_sync", "sync engine is not configured")
	}
	target, err := o.vault.LoadTarget(ctx, req.TargetID)
	if err != nil {
		return model.JobResult{}, err
	}
	result, err := o.sync.Run(ctx, req, target)
	if err != nil {
		return model.JobResult{}, err
	}
	if o.jobs != nil {
		_ = o.jobs.PutJob(ctx, result.Summary)
	}
	return model.JobResult{ID: result.ID, Summary: result.Summary}, nil
}

func (o *Orchestrator) StartExec(ctx context.Context, req model.ExecRequest) (model.ExecStartResult, error) {
	if o.exec == nil {
		return model.ExecStartResult{}, toolerrors.New("exec_missing", "validation", "start_exec", "exec engine is not configured")
	}
	target, err := o.vault.LoadTarget(ctx, req.TargetID)
	if err != nil {
		return model.ExecStartResult{}, err
	}
	return o.exec.Start(ctx, req, target)
}

func (o *Orchestrator) ReadExec(ctx context.Context, req model.ExecReadRequest) (model.ExecReadResult, error) {
	if o.exec == nil {
		return model.ExecReadResult{}, toolerrors.New("exec_missing", "validation", "read_exec", "exec engine is not configured")
	}
	return o.exec.Read(ctx, req)
}

func (o *Orchestrator) StopExec(ctx context.Context, req model.ExecStopRequest) (model.ExecStopResult, error) {
	if o.exec == nil {
		return model.ExecStopResult{}, toolerrors.New("exec_missing", "validation", "stop_exec", "exec engine is not configured")
	}
	return o.exec.Stop(ctx, req)
}

func (o *Orchestrator) GetJob(ctx context.Context, req model.JobStatusRequest) (model.JobResult, error) {
	if req.ID == "" {
		return model.JobResult{}, fmt.Errorf("id is required")
	}
	if o.jobs == nil {
		return model.JobResult{}, toolerrors.New("jobs_missing", "validation", "get_job", "session store is not configured")
	}
	summary, err := o.jobs.GetJob(ctx, req.ID)
	if err != nil {
		return model.JobResult{}, err
	}
	return model.JobResult{
		ID:      summary.ID,
		Summary: summary,
	}, nil
}
