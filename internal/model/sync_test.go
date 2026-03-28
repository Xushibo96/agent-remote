package model

import "testing"

func TestNormalizeSyncRequestInvalidDirection(t *testing.T) {
	_, err := NormalizeSyncRequest(SyncRequest{
		TargetID:  "prod",
		Direction: "sideways",
		LocalPath: ".",
		RemotePath: "/tmp",
	})
	if err == nil {
		t.Fatal("expected invalid direction error")
	}
}
