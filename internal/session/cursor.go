package session

import (
	"fmt"
	"strconv"
	"strings"
)

const cursorPrefix = "seq:"

func EncodeCursor(seq int64) string {
	if seq <= 0 {
		return ""
	}
	return cursorPrefix + strconv.FormatInt(seq, 10)
}

func DecodeCursor(cursor string) (int64, error) {
	if cursor == "" {
		return 0, nil
	}
	if !strings.HasPrefix(cursor, cursorPrefix) {
		return 0, fmt.Errorf("invalid cursor %q", cursor)
	}
	return strconv.ParseInt(strings.TrimPrefix(cursor, cursorPrefix), 10, 64)
}

