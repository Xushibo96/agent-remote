package session

import (
	"sync"
	"time"

	"agent-remote/internal/model"
)

type RingBuffer struct {
	mu       sync.RWMutex
	items    []model.ExecEvent
	capacity int
	startSeq int64
	nextSeq  int64
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 128
	}
	return &RingBuffer{
		items:    make([]model.ExecEvent, 0, capacity),
		capacity: capacity,
	}
}

func (b *RingBuffer) Append(event model.ExecEvent) model.ExecEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	if event.Seq <= 0 {
		event.Seq = b.nextSeq + 1
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if len(b.items) == b.capacity {
		copy(b.items, b.items[1:])
		b.items[len(b.items)-1] = event
		b.startSeq = b.items[0].Seq
	} else {
		b.items = append(b.items, event)
		if len(b.items) == 1 {
			b.startSeq = event.Seq
		}
	}
	b.nextSeq = event.Seq
	return event
}

func (b *RingBuffer) SnapshotAfter(cursor string) ([]model.ExecEvent, string, bool, error) {
	seq, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", false, err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.items) == 0 {
		return nil, cursor, false, nil
	}

	out := make([]model.ExecEvent, 0, len(b.items))
	for _, event := range b.items {
		if event.Seq > seq {
			out = append(out, event)
		}
	}
	if len(out) == 0 {
		return nil, EncodeCursor(b.items[len(b.items)-1].Seq), false, nil
	}

	nextCursor := EncodeCursor(out[len(out)-1].Seq)
	truncated := (seq > 0 && seq < b.startSeq) || (seq == 0 && b.nextSeq > int64(len(b.items)))
	return out, nextCursor, truncated, nil
}

func (b *RingBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.items)
}

func (b *RingBuffer) NextSeq() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.nextSeq + 1
}

func (b *RingBuffer) LatestCursor() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.items) == 0 {
		return ""
	}
	return EncodeCursor(b.items[len(b.items)-1].Seq)
}
