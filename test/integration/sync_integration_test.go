package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agent-remote/internal/model"
	"agent-remote/internal/rsync"
	syncengine "agent-remote/internal/sync"
)

func TestLocalSyncUploadDownloadRoundTrip(t *testing.T) {
	srcRoot := t.TempDir()
	dstRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcRoot, "a.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	engine := syncengine.NewEngine(syncengine.NewRootedFilesystem(srcRoot, nil), syncengine.NewRootedFilesystem(dstRoot, nil), rsync.NewRunner(""))
	_, err := engine.Run(context.Background(), model.SyncRequest{
		JobID:      "sync-1",
		TargetID:   "prod",
		Direction:  model.DirectionUpload,
		LocalPath:  ".",
		RemotePath: ".",
		CreateDirs: true,
	}, model.ResolvedTarget{})
	if err != nil {
		t.Fatalf("Run(upload) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dstRoot, "a.txt")); err != nil {
		t.Fatalf("expected uploaded file: %v", err)
	}
}

