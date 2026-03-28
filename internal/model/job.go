package model

import "time"

type JobStatusRequest struct {
	ID string
}

type RemoteCapabilities struct {
	SSHAvailable   bool   `json:"ssh_available"`
	SFTPAvailable  bool   `json:"sftp_available"`
	RsyncAvailable bool   `json:"rsync_available"`
	ShellPath      string `json:"shell_path,omitempty"`
	OS             string `json:"os,omitempty"`
}

type JobSummary struct {
	ID               string     `json:"id"`
	Kind             string     `json:"kind"`
	State            string     `json:"state"`
	StartedAt        time.Time  `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
	FilesTransferred int64      `json:"files_transferred,omitempty"`
	FilesSkipped     int64      `json:"files_skipped,omitempty"`
	BytesTransferred int64      `json:"bytes_transferred,omitempty"`
	EventCount       int64      `json:"event_count,omitempty"`
	TruncatedEvents  int64      `json:"truncated_events,omitempty"`
	ExitCode         *int       `json:"exit_code,omitempty"`
	FailureStage     string     `json:"failure_stage,omitempty"`
	Retryable        bool       `json:"retryable,omitempty"`
}

type JobResult struct {
	ID      string     `json:"id"`
	Summary JobSummary `json:"summary"`
}

type ExecStartResult struct {
	ID      string     `json:"id"`
	Summary JobSummary `json:"summary"`
	Cursor  string     `json:"cursor,omitempty"`
}

type ExecReadResult struct {
	ID        string      `json:"id"`
	Events    []ExecEvent `json:"events"`
	Cursor    string      `json:"cursor,omitempty"`
	Truncated bool        `json:"truncated,omitempty"`
	Summary   JobSummary  `json:"summary"`
}

type ExecStopResult struct {
	ID      string     `json:"id"`
	Summary JobSummary `json:"summary"`
}

type SyncRunResult struct {
	ID               string     `json:"id"`
	EffectiveBackend string     `json:"effective_backend"`
	Summary          JobSummary `json:"summary"`
}
