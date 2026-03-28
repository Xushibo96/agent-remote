package session

import (
	"testing"

	"agent-remote/internal/model"
)

func TestRingBufferSnapshotAfter(t *testing.T) {
	buf := NewRingBuffer(3)
	for i := 1; i <= 5; i++ {
		buf.Append(model.ExecEvent{Seq: int64(i), Type: "stdout", Payload: string(rune('a' + i - 1))})
	}

	events, cursor, truncated, err := buf.SnapshotAfter("")
	if err != nil {
		t.Fatalf("SnapshotAfter() error = %v", err)
	}
	if !truncated {
		t.Fatal("expected truncated to be true when earlier events were dropped")
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if cursor == "" {
		t.Fatal("expected non-empty cursor")
	}
}

