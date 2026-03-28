package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agent-remote/internal/model"
)

func TestChooseBackendFallsBackToSFTP(t *testing.T) {
	backend := ChooseBackend(model.SyncRequest{BackendPreference: model.BackendAuto}, model.RemoteCapabilities{})
	if backend != BackendSFTP {
		t.Fatalf("expected sftp fallback, got %s", backend)
	}
}

func TestSyncEngineUploadAndDownload(t *testing.T) {
	localRoot := t.TempDir()
	remoteRoot := t.TempDir()

	if err := os.WriteFile(filepath.Join(localRoot, "file.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(NewRootedFilesystem(localRoot, nil), NewRootedFilesystem(remoteRoot, nil), nil)
	req := model.SyncRequest{
		JobID:      "job-1",
		TargetID:   "prod",
		Direction:  model.DirectionUpload,
		LocalPath:  ".",
		RemotePath: ".",
		CreateDirs: true,
	}
	result, err := engine.Run(context.Background(), req, model.ResolvedTarget{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Summary.FilesTransferred == 0 {
		t.Fatal("expected files transferred")
	}
	if _, err := os.Stat(filepath.Join(remoteRoot, "file.txt")); err != nil {
		t.Fatalf("expected file in remote root: %v", err)
	}

	downloadRoot := t.TempDir()
	downloadEngine := NewEngine(NewRootedFilesystem(downloadRoot, nil), NewRootedFilesystem(remoteRoot, nil), nil)
	downloadReq := model.SyncRequest{
		JobID:      "job-2",
		TargetID:   "prod",
		Direction:  model.DirectionDownload,
		LocalPath:  ".",
		RemotePath: ".",
		CreateDirs: true,
	}
	_, err = downloadEngine.Run(context.Background(), downloadReq, model.ResolvedTarget{})
	if err != nil {
		t.Fatalf("download Run() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(downloadRoot, "file.txt")); err != nil {
		t.Fatalf("expected file in download root: %v", err)
	}
}

func TestPlannerConflictResolution(t *testing.T) {
	now := time.Now().UTC()
	planner := NewPlanner()
	actions, err := planner.PlanBidirectional(
		Snapshot{Files: map[string]FileMeta{
			"a.txt": {Path: "a.txt", Size: 1, ModTime: now, IsDir: false},
		}},
		Snapshot{Files: map[string]FileMeta{
			"a.txt": {Path: "a.txt", Size: 2, ModTime: now.Add(-time.Minute), IsDir: false},
		}},
		model.ConflictNewerWins,
	)
	if err != nil {
		t.Fatalf("PlanBidirectional() error = %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Type != PlanUpload {
		t.Fatalf("expected local newer to upload, got %s", actions[0].Type)
	}
}
