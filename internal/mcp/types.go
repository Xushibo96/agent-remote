package mcp

import "agent-remote/internal/model"

type ErrorInfo struct {
	Code        string `json:"code"`
	Category    string `json:"category"`
	Stage       string `json:"stage"`
	Message     string `json:"message"`
	Retryable   bool   `json:"retryable,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

type Response struct {
	OK               bool              `json:"ok"`
	ID               string            `json:"id,omitempty"`
	State            string            `json:"state,omitempty"`
	Summary          any               `json:"summary,omitempty"`
	Events           []model.ExecEvent `json:"events,omitempty"`
	Cursor           string            `json:"cursor,omitempty"`
	Truncated        bool              `json:"truncated,omitempty"`
	EffectiveBackend string            `json:"effective_backend,omitempty"`
	Error            *ErrorInfo        `json:"error,omitempty"`
}

func errorInfo(code, category, stage, message string) *ErrorInfo {
	return &ErrorInfo{
		Code:     code,
		Category: category,
		Stage:    stage,
		Message:  message,
	}
}
