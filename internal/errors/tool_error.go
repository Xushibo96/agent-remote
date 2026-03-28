package errors

import "fmt"

type ToolError struct {
	Code        string
	Category    string
	Stage       string
	Message     string
	Retryable   bool
	Remediation string
	Details     map[string]any
}

func (e *ToolError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code == "" {
		return e.Message
	}
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, category, stage, message string) *ToolError {
	return &ToolError{
		Code:     code,
		Category: category,
		Stage:    stage,
		Message:  message,
	}
}
