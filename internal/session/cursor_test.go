package session

import "testing"

func TestCursorRoundTrip(t *testing.T) {
	cursor := EncodeCursor(42)
	seq, err := DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("DecodeCursor() error = %v", err)
	}
	if seq != 42 {
		t.Fatalf("expected 42, got %d", seq)
	}
}

