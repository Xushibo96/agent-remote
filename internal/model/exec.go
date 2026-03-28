package model

import "time"

type ExecRequest struct {
	SessionID      string
	TargetID       string
	Command        string
	WorkDir        string
	Env            map[string]string
	PTY            bool
	TimeoutSeconds int
	Budget         BudgetPolicy
}

type ExecReadRequest struct {
	SessionID string
	Cursor    string
	Budget    BudgetPolicy
}

type ExecStopRequest struct {
	SessionID string
}

type ExecEvent struct {
	Seq        int64     `json:"seq"`
	SessionID  string    `json:"session_id"`
	Type       string    `json:"type"`
	Stream     string    `json:"stream"`
	Payload    string    `json:"payload,omitempty"`
	ByteOffset int64     `json:"byte_offset,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Truncated  bool      `json:"truncated,omitempty"`
}
