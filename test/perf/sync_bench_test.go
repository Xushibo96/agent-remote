package perf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agent-remote/internal/model"
	"agent-remote/internal/rsync"
	syncengine "agent-remote/internal/sync"
)

func BenchmarkSyncUploadDirectory(b *testing.B) {
	srcRoot := b.TempDir()
	dstRoot := b.TempDir()
	for i := 0; i < 100; i++ {
		name := filepath.Join(srcRoot, "dir", "file-"+string(rune('a'+(i%26)))+".txt")
		if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			b.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(name, []byte("payload"), 0o600); err != nil {
			b.Fatalf("WriteFile() error = %v", err)
		}
	}
	engine := syncengine.NewEngine(syncengine.NewRootedFilesystem(srcRoot, nil), syncengine.NewRootedFilesystem(dstRoot, nil), rsync.NewRunner(""))
	req := model.SyncRequest{JobID: "bench-sync", TargetID: "prod", Direction: model.DirectionUpload, LocalPath: ".", RemotePath: ".", CreateDirs: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := engine.Run(context.Background(), req, model.ResolvedTarget{}); err != nil {
			b.Fatalf("Run() error = %v", err)
		}
	}
}

