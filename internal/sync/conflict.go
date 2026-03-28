package sync

import (
	"fmt"
	"time"

	"agent-remote/internal/model"
)

func ResolveConflict(local, remote FileMeta, policy string) (PlanActionType, string, error) {
	switch policy {
	case "", model.ConflictOverwrite:
		return PlanUpload, "overwrite", nil
	case model.ConflictSkip:
		return PlanSkip, "skip", nil
	case model.ConflictFail:
		return PlanConflict, "conflict", fmt.Errorf("conflict for %s", local.Path)
	case model.ConflictNewerWins:
		if newer(local.ModTime, remote.ModTime) {
			return PlanUpload, "local newer", nil
		}
		return PlanDownload, "remote newer", nil
	default:
		return PlanConflict, "unsupported policy", fmt.Errorf("unsupported conflict policy %q", policy)
	}
}

func newer(a, b time.Time) bool {
	if a.IsZero() {
		return false
	}
	if b.IsZero() {
		return true
	}
	return a.After(b) || a.Equal(b)
}
