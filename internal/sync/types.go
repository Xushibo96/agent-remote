package sync

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"agent-remote/internal/model"
)

var ErrNotImplemented = errors.New("not implemented")

type Filesystem interface {
	Stat(name string) (os.FileInfo, error)
	ReadDir(name string) ([]os.DirEntry, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
}

type OSFilesystem struct{}

func (OSFilesystem) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (OSFilesystem) ReadDir(name string) ([]os.DirEntry, error) { return os.ReadDir(name) }
func (OSFilesystem) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (OSFilesystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (OSFilesystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (OSFilesystem) RemoveAll(path string) error { return os.RemoveAll(path) }

type RootedFilesystem struct {
	Root string
	FS   Filesystem
}

func NewRootedFilesystem(root string, fs Filesystem) RootedFilesystem {
	if fs == nil {
		fs = OSFilesystem{}
	}
	return RootedFilesystem{Root: root, FS: fs}
}

func (r RootedFilesystem) join(name string) string {
	if filepath.IsAbs(name) {
		name = name[1:]
	}
	return filepath.Join(r.Root, filepath.Clean(name))
}

func (r RootedFilesystem) Stat(name string) (os.FileInfo, error) { return r.FS.Stat(r.join(name)) }
func (r RootedFilesystem) ReadDir(name string) ([]os.DirEntry, error) { return r.FS.ReadDir(r.join(name)) }
func (r RootedFilesystem) ReadFile(name string) ([]byte, error) { return r.FS.ReadFile(r.join(name)) }
func (r RootedFilesystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	if err := r.FS.MkdirAll(filepath.Dir(r.join(name)), 0o755); err != nil {
		return err
	}
	return r.FS.WriteFile(r.join(name), data, perm)
}
func (r RootedFilesystem) MkdirAll(path string, perm os.FileMode) error { return r.FS.MkdirAll(r.join(path), perm) }
func (r RootedFilesystem) RemoveAll(path string) error { return r.FS.RemoveAll(r.join(path)) }

type Backend string

const (
	BackendAuto  Backend = model.BackendAuto
	BackendRsync Backend = model.BackendRsync
	BackendSFTP  Backend = model.BackendSFTP
)

type TransferResult struct {
	FilesTransferred int64
	FilesSkipped     int64
	BytesTransferred int64
}

type TransferOptions struct {
	Overwrite    bool
	CreateDirs   bool
	Conflict     string
	MaxFileSize  int64
	Includes     []string
	Excludes     []string
	Resume       bool
}

type Transferer interface {
	Upload(ctx context.Context, source Filesystem, sourcePath string, target Filesystem, targetPath string, opts TransferOptions) (TransferResult, error)
	Download(ctx context.Context, source Filesystem, sourcePath string, target Filesystem, targetPath string, opts TransferOptions) (TransferResult, error)
}

type Capabilities struct {
	RsyncAvailable bool
	SFTPAvailable  bool
}

type BidirectionalPlanner interface {
	PlanBidirectional(local Snapshot, remote Snapshot, policy string) ([]PlanAction, error)
}

type Snapshot struct {
	Root  string
	Files map[string]FileMeta
}

type FileMeta struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

type PlanActionType string

const (
	PlanUpload  PlanActionType = "upload"
	PlanDownload PlanActionType = "download"
	PlanSkip    PlanActionType = "skip"
	PlanConflict PlanActionType = "conflict"
)

type PlanAction struct {
	Type   PlanActionType
	Path   string
	Reason string
}
