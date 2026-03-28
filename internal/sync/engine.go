package sync

import (
	"context"
	"fmt"
	"path/filepath"

	"agent-remote/internal/interfaces"
	"agent-remote/internal/model"
	"agent-remote/internal/rsync"
)

type Engine struct {
	LocalFS     Filesystem
	RemoteFS    Filesystem
	Planner     BidirectionalPlanner
	Transferer  Transferer
	Runner      *rsync.Runner
	Caps        model.RemoteCapabilities
	Connections interfaces.ConnectionManager
}

func NewEngine(localFS, remoteFS Filesystem, runner *rsync.Runner) *Engine {
	return &Engine{
		LocalFS:    localFS,
		RemoteFS:   remoteFS,
		Planner:    NewPlanner(),
		Transferer: NewTransferer(),
		Runner:     runner,
	}
}

func NewRemoteEngine(connections interfaces.ConnectionManager, runner *rsync.Runner) *Engine {
	return &Engine{
		LocalFS:     OSFilesystem{},
		Planner:     NewPlanner(),
		Transferer:  NewTransferer(),
		Runner:      runner,
		Connections: connections,
	}
}

func (e *Engine) Run(ctx context.Context, req model.SyncRequest, target model.ResolvedTarget) (model.SyncRunResult, error) {
	normalized, err := model.NormalizeSyncRequest(req)
	if err != nil {
		return model.SyncRunResult{}, err
	}

	localFS := e.LocalFS
	if localFS == nil {
		localFS = OSFilesystem{}
	}

	remoteFS := e.RemoteFS
	caps := e.Caps
	var cleanup func() error
	if remoteFS == nil {
		if e.Connections == nil {
			return model.SyncRunResult{}, fmt.Errorf("remote filesystem is not configured")
		}
		if caps == (model.RemoteCapabilities{}) {
			detected, err := e.Connections.DetectCapabilities(ctx, target)
			if err != nil {
				return model.SyncRunResult{}, err
			}
			caps = detected
		}
		remoteFS, cleanup, err = NewSFTPFilesystem(ctx, e.Connections, target, target.DefaultBaseDir)
		if err != nil {
			return model.SyncRunResult{}, err
		}
	}
	if cleanup != nil {
		defer cleanup()
	}

	backend := ChooseBackend(normalized, caps)
	result := model.SyncRunResult{ID: normalized.JobID, EffectiveBackend: string(backend)}

	opts := TransferOptions{
		Overwrite:   normalized.Overwrite == nil || *normalized.Overwrite,
		CreateDirs:  normalized.CreateDirs,
		Conflict:    normalized.ConflictPolicy,
		MaxFileSize: normalized.MaxFileSizeBytes,
		Includes:    normalized.Includes,
		Excludes:    normalized.Excludes,
		Resume:      normalized.Resume,
	}

	switch normalized.Direction {
	case model.DirectionUpload:
		stats, err := e.Transferer.Upload(ctx, localFS, normalized.LocalPath, remoteFS, normalized.RemotePath, opts)
		if err != nil {
			return model.SyncRunResult{}, err
		}
		result.Summary.FilesTransferred = stats.FilesTransferred
		result.Summary.FilesSkipped = stats.FilesSkipped
		result.Summary.BytesTransferred = stats.BytesTransferred
	case model.DirectionDownload:
		stats, err := e.Transferer.Download(ctx, remoteFS, normalized.RemotePath, localFS, normalized.LocalPath, opts)
		if err != nil {
			return model.SyncRunResult{}, err
		}
		result.Summary.FilesTransferred = stats.FilesTransferred
		result.Summary.FilesSkipped = stats.FilesSkipped
		result.Summary.BytesTransferred = stats.BytesTransferred
	case model.DirectionBidir:
		localSnap, err := SnapshotFS(localFS, normalized.LocalPath)
		if err != nil {
			return model.SyncRunResult{}, err
		}
		remoteSnap, err := SnapshotFS(remoteFS, normalized.RemotePath)
		if err != nil {
			return model.SyncRunResult{}, err
		}
		actions, err := e.Planner.PlanBidirectional(localSnap, remoteSnap, normalized.ConflictPolicy)
		if err != nil {
			return model.SyncRunResult{}, err
		}
		for _, action := range actions {
			switch action.Type {
			case PlanUpload:
				stats, err := e.Transferer.Upload(ctx, localFS, filepath.Join(normalized.LocalPath, action.Path), remoteFS, filepath.Join(normalized.RemotePath, action.Path), opts)
				if err != nil {
					return model.SyncRunResult{}, err
				}
				result.Summary.FilesTransferred += stats.FilesTransferred
				result.Summary.BytesTransferred += stats.BytesTransferred
			case PlanDownload:
				stats, err := e.Transferer.Download(ctx, remoteFS, filepath.Join(normalized.RemotePath, action.Path), localFS, filepath.Join(normalized.LocalPath, action.Path), opts)
				if err != nil {
					return model.SyncRunResult{}, err
				}
				result.Summary.FilesTransferred += stats.FilesTransferred
				result.Summary.BytesTransferred += stats.BytesTransferred
			case PlanSkip:
				result.Summary.FilesSkipped++
			case PlanConflict:
				return model.SyncRunResult{}, fmt.Errorf("conflict at %s: %s", action.Path, action.Reason)
			}
		}
	default:
		return model.SyncRunResult{}, fmt.Errorf("unsupported direction %q", normalized.Direction)
	}

	return result, nil
}

func ChooseBackend(req model.SyncRequest, caps model.RemoteCapabilities) Backend {
	switch Backend(req.BackendPreference) {
	case BackendRsync:
		return BackendRsync
	case BackendSFTP:
		return BackendSFTP
	}
	if caps.RsyncAvailable {
		return BackendRsync
	}
	return BackendSFTP
}

func SnapshotFS(fsys Filesystem, root string) (Snapshot, error) {
	snap := Snapshot{Root: root, Files: map[string]FileMeta{}}
	info, err := fsys.Stat(root)
	if err != nil {
		return snap, err
	}
	if !info.IsDir() {
		snap.Files[filepath.Base(root)] = FileMeta{Path: filepath.Base(root), Size: info.Size(), ModTime: info.ModTime(), IsDir: false}
		return snap, nil
	}
	err = walkSnapshot(fsys, root, "", snap.Files)
	return snap, err
}

func walkSnapshot(fsys Filesystem, root, rel string, files map[string]FileMeta) error {
	current := filepath.Join(root, rel)
	entries, err := fsys.ReadDir(current)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		childRel := filepath.Join(rel, name)
		fi, err := fsys.Stat(filepath.Join(root, childRel))
		if err != nil {
			return err
		}
		files[childRel] = FileMeta{
			Path:    childRel,
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
			IsDir:   fi.IsDir(),
		}
		if fi.IsDir() {
			if err := walkSnapshot(fsys, root, childRel, files); err != nil {
				return err
			}
		}
	}
	return nil
}
