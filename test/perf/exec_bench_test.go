package perf

import (
	"testing"
	"time"

	"agent-remote/internal/model"
	"agent-remote/internal/session"
)

func BenchmarkExecEventAppend(b *testing.B) {
	store := session.NewStore()
	store.CreateSession(model.JobSummary{ID: "bench-exec", Kind: "exec", State: "running", StartedAt: time.Now().UTC()}, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.AppendEvent("bench-exec", model.ExecEvent{
			SessionID: "bench-exec",
			Type:      "stdout",
			Stream:    "stdout",
			Payload:   "payload",
		}); err != nil {
			b.Fatalf("AppendEvent() error = %v", err)
		}
	}
}

