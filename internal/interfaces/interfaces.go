package interfaces

import (
	"context"

	"agent-remote/internal/model"
	"golang.org/x/crypto/ssh"
)

type Orchestrator interface {
	AddTarget(ctx context.Context, req model.AddTargetRequest) (model.TargetConfig, error)
	StartSync(ctx context.Context, req model.SyncRequest) (model.JobResult, error)
	StartExec(ctx context.Context, req model.ExecRequest) (model.ExecStartResult, error)
	ReadExec(ctx context.Context, req model.ExecReadRequest) (model.ExecReadResult, error)
	StopExec(ctx context.Context, req model.ExecStopRequest) (model.ExecStopResult, error)
	GetJob(ctx context.Context, req model.JobStatusRequest) (model.JobResult, error)
}

type CredentialVault interface {
	SaveTarget(ctx context.Context, req model.AddTargetRequest) (model.TargetConfig, error)
	LoadTarget(ctx context.Context, targetID string) (model.ResolvedTarget, error)
	ListTargets(ctx context.Context) ([]model.TargetConfig, error)
	RotateMasterKey(ctx context.Context) error
}

type ConnectionManager interface {
	GetSSHClient(ctx context.Context, target model.ResolvedTarget) (*ssh.Client, error)
	DetectCapabilities(ctx context.Context, target model.ResolvedTarget) (model.RemoteCapabilities, error)
	CloseIdle() error
}

type SyncEngine interface {
	Run(ctx context.Context, req model.SyncRequest, target model.ResolvedTarget) (model.SyncRunResult, error)
}

type ExecEngine interface {
	Start(ctx context.Context, req model.ExecRequest, target model.ResolvedTarget) (model.ExecStartResult, error)
	Read(ctx context.Context, req model.ExecReadRequest) (model.ExecReadResult, error)
	Stop(ctx context.Context, req model.ExecStopRequest) (model.ExecStopResult, error)
}

type OutputBudgeter interface {
	BudgetSync(result model.SyncRunResult, budget model.BudgetPolicy) model.SyncRunResult
	BudgetExec(events []model.ExecEvent, budget model.BudgetPolicy, cursor string, summary model.JobSummary) model.ExecReadResult
}

type SessionStore interface {
	PutJob(ctx context.Context, summary model.JobSummary) error
	GetJob(ctx context.Context, id string) (model.JobSummary, error)
}
