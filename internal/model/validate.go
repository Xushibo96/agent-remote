package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	AuthModePassword = "password"
	AuthModeKey      = "key"

	KnownHostsStrict    = "strict"
	KnownHostsAcceptNew = "accept-new"
	KnownHostsInsecure  = "insecure"

	DirectionUpload   = "upload"
	DirectionDownload = "download"
	DirectionBidir    = "bidir"

	ConflictOverwrite = "overwrite"
	ConflictSkip      = "skip"
	ConflictFail      = "fail"
	ConflictNewerWins = "newer-wins"

	BackendAuto  = "auto"
	BackendRsync = "rsync"
	BackendSFTP  = "sftp"
)

func NormalizeAddTargetRequest(req AddTargetRequest) (AddTargetRequest, error) {
	req.ID = strings.TrimSpace(req.ID)
	req.Name = strings.TrimSpace(req.Name)
	req.Host = strings.TrimSpace(req.Host)
	req.User = strings.TrimSpace(req.User)
	req.AuthMode = strings.TrimSpace(req.AuthMode)
	req.PrivateKeyPath = strings.TrimSpace(req.PrivateKeyPath)
	req.KnownHostsPolicy = strings.TrimSpace(req.KnownHostsPolicy)
	req.DefaultBaseDir = strings.TrimSpace(req.DefaultBaseDir)

	if req.Port == 0 {
		req.Port = 22
	}
	if req.AuthMode == "" {
		req.AuthMode = AuthModePassword
	}
	if req.KnownHostsPolicy == "" {
		req.KnownHostsPolicy = KnownHostsStrict
	}

	switch req.AuthMode {
	case AuthModePassword:
		if req.Password == "" {
			return req, fmt.Errorf("password is required for password auth")
		}
	case AuthModeKey:
		if req.PrivateKeyPath == "" {
			return req, fmt.Errorf("private_key_path is required for key auth")
		}
	default:
		return req, fmt.Errorf("unsupported auth_mode %q", req.AuthMode)
	}

	switch req.KnownHostsPolicy {
	case KnownHostsStrict, KnownHostsAcceptNew, KnownHostsInsecure:
	default:
		return req, fmt.Errorf("unsupported known_hosts_policy %q", req.KnownHostsPolicy)
	}

	if req.Host == "" {
		return req, fmt.Errorf("host is required")
	}
	if req.User == "" {
		return req, fmt.Errorf("user is required")
	}
	if req.ID == "" {
		req.ID = slug(strings.ToLower(firstNonEmpty(req.Name, req.Host)))
	}
	return req, nil
}

func NormalizeSyncRequest(req SyncRequest) (SyncRequest, error) {
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Direction = strings.TrimSpace(req.Direction)
	req.LocalPath = strings.TrimSpace(req.LocalPath)
	req.RemotePath = strings.TrimSpace(req.RemotePath)
	req.ConflictPolicy = strings.TrimSpace(req.ConflictPolicy)
	req.BackendPreference = strings.TrimSpace(req.BackendPreference)

	if req.ConflictPolicy == "" {
		req.ConflictPolicy = ConflictOverwrite
	}
	if req.BackendPreference == "" {
		req.BackendPreference = BackendAuto
	}

	switch req.Direction {
	case DirectionUpload:
		if req.LocalPath == "" || req.RemotePath == "" {
			return req, fmt.Errorf("upload requires local_path and remote_path")
		}
	case DirectionDownload:
		if req.LocalPath == "" || req.RemotePath == "" {
			return req, fmt.Errorf("download requires local_path and remote_path")
		}
	case DirectionBidir:
		if req.LocalPath == "" || req.RemotePath == "" {
			return req, fmt.Errorf("bidir requires local_path and remote_path")
		}
	default:
		return req, fmt.Errorf("unsupported direction %q", req.Direction)
	}

	switch req.ConflictPolicy {
	case ConflictOverwrite, ConflictSkip, ConflictFail, ConflictNewerWins:
	default:
		return req, fmt.Errorf("unsupported conflict_policy %q", req.ConflictPolicy)
	}

	switch req.BackendPreference {
	case BackendAuto, BackendRsync, BackendSFTP:
	default:
		return req, fmt.Errorf("unsupported backend_preference %q", req.BackendPreference)
	}

	if req.TargetID == "" {
		return req, fmt.Errorf("target_id is required")
	}

	req.LocalPath = filepath.Clean(req.LocalPath)
	return req, nil
}

func NormalizeExecRequest(req ExecRequest) (ExecRequest, error) {
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Command = strings.TrimSpace(req.Command)
	req.WorkDir = strings.TrimSpace(req.WorkDir)
	if req.TargetID == "" {
		return req, fmt.Errorf("target_id is required")
	}
	if req.Command == "" {
		return req, fmt.Errorf("command is required")
	}
	if req.TimeoutSeconds < 0 {
		return req, fmt.Errorf("timeout_seconds must be >= 0")
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 600
	}
	return req, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func slug(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "_", "-")
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case r == '-':
			if !lastDash {
				b.WriteRune(r)
				lastDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "target"
	}
	return result
}
