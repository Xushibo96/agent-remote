package sync

import (
	"fmt"
	"sort"

	"agent-remote/internal/model"
)

type Planner struct{}

func NewPlanner() *Planner { return &Planner{} }

func (p *Planner) PlanBidirectional(local Snapshot, remote Snapshot, policy string) ([]PlanAction, error) {
	keys := make(map[string]struct{}, len(local.Files)+len(remote.Files))
	for k := range local.Files {
		keys[k] = struct{}{}
	}
	for k := range remote.Files {
		keys[k] = struct{}{}
	}

	paths := make([]string, 0, len(keys))
	for path := range keys {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	actions := make([]PlanAction, 0, len(paths))
	for _, path := range paths {
		l, lok := local.Files[path]
		r, rok := remote.Files[path]
		switch {
		case lok && !rok:
			actions = append(actions, PlanAction{Type: PlanUpload, Path: path, Reason: "local-only"})
		case !lok && rok:
			actions = append(actions, PlanAction{Type: PlanDownload, Path: path, Reason: "remote-only"})
		case lok && rok:
			if l.Size == r.Size && l.ModTime.Equal(r.ModTime) {
				actions = append(actions, PlanAction{Type: PlanSkip, Path: path, Reason: "identical"})
				continue
			}
			kind, reason, err := ResolveConflict(l, r, policy)
			if err != nil && policy == model.ConflictFail {
				actions = append(actions, PlanAction{Type: PlanConflict, Path: path, Reason: err.Error()})
				continue
			}
			actions = append(actions, PlanAction{Type: kind, Path: path, Reason: reason})
		default:
			return nil, fmt.Errorf("unreachable path state for %s", path)
		}
	}
	return actions, nil
}
