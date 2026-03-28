package model

type BudgetPolicy struct {
	MaxBytes     int  `json:"max_bytes,omitempty"`
	MaxLines     int  `json:"max_lines,omitempty"`
	WindowBytes  int  `json:"window_bytes,omitempty"`
	KeepErrors   bool `json:"keep_errors,omitempty"`
	KeepLifecycle bool `json:"keep_lifecycle,omitempty"`
}

type SyncRequest struct {
	JobID             string
	TargetID          string
	Direction         string
	LocalPath         string
	RemotePath        string
	Includes          []string
	Excludes          []string
	MaxFileSizeBytes  int64
	Overwrite         *bool
	ConflictPolicy    string
	BackendPreference string
	CreateDirs        bool
	Resume            bool
	Budget            BudgetPolicy
}
