package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"agent-remote/internal/model"
)

var ErrNotFound = errors.New("session not found")

type record struct {
	summary model.JobSummary
	events  *RingBuffer
	updated time.Time
}

type Store struct {
	mu      sync.RWMutex
	jobs    map[string]model.JobSummary
	records map[string]*record
	path    string
}

func NewStore() *Store {
	return &Store{
		jobs:    make(map[string]model.JobSummary),
		records: make(map[string]*record),
	}
}

func NewStoreWithPath(path string) (*Store, error) {
	store := NewStore()
	store.path = path
	if path == "" {
		return store, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) PutJob(_ context.Context, summary model.JobSummary) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[summary.ID] = summary
	return s.persistLocked()
}

func (s *Store) GetJob(_ context.Context, id string) (model.JobSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.jobs[id]
	if !ok {
		return model.JobSummary{}, ErrNotFound
	}
	return summary, nil
}

func (s *Store) CreateSession(summary model.JobSummary, capacity int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if summary.ID == "" {
		return
	}
	s.records[summary.ID] = &record{
		summary: summary,
		events:  NewRingBuffer(capacity),
		updated: time.Now().UTC(),
	}
	s.jobs[summary.ID] = summary
	_ = s.persistLocked()
}

func (s *Store) AppendEvent(sessionID string, event model.ExecEvent) (model.ExecEvent, error) {
	s.mu.RLock()
	rec, ok := s.records[sessionID]
	s.mu.RUnlock()
	if !ok {
		return model.ExecEvent{}, ErrNotFound
	}

	evt := rec.events.Append(event)

	var err error
	s.mu.Lock()
	rec.summary.EventCount++
	rec.updated = time.Now().UTC()
	s.jobs[sessionID] = rec.summary
	err = s.persistLocked()
	s.mu.Unlock()
	if err != nil {
		return model.ExecEvent{}, err
	}

	return evt, nil
}

func (s *Store) ReadEvents(sessionID string, cursor string) ([]model.ExecEvent, string, bool, model.JobSummary, error) {
	s.mu.RLock()
	rec, ok := s.records[sessionID]
	s.mu.RUnlock()
	if !ok {
		return nil, "", false, model.JobSummary{}, ErrNotFound
	}

	events, nextCursor, truncated, err := rec.events.SnapshotAfter(cursor)
	if err != nil {
		return nil, "", false, model.JobSummary{}, err
	}

	return events, nextCursor, truncated, rec.summary, nil
}

func (s *Store) UpdateSummary(summary model.JobSummary) error {
	if summary.ID == "" {
		return fmt.Errorf("summary id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[summary.ID] = summary
	if rec, ok := s.records[summary.ID]; ok {
		rec.summary = summary
		rec.updated = time.Now().UTC()
	}
	return s.persistLocked()
}

type persistedStore struct {
	Jobs    map[string]model.JobSummary `json:"jobs"`
	Records map[string]persistedRecord  `json:"records"`
}

type persistedRecord struct {
	Summary  model.JobSummary   `json:"summary"`
	Events   []model.ExecEvent  `json:"events"`
	Updated  time.Time          `json:"updated"`
	Capacity int                `json:"capacity"`
}

func (s *Store) persistLocked() error {
	if s.path == "" {
		return nil
	}
	payload := persistedStore{
		Jobs:    s.jobs,
		Records: make(map[string]persistedRecord, len(s.records)),
	}
	for id, rec := range s.records {
		events, _, _, err := rec.events.SnapshotAfter("")
		if err != nil {
			return err
		}
		payload.Records[id] = persistedRecord{
			Summary:  rec.summary,
			Events:   events,
			Updated:  rec.updated,
			Capacity: max(rec.events.capacity, len(events)),
		}
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var payload persistedStore
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	if payload.Jobs != nil {
		s.jobs = payload.Jobs
	}
	for id, rec := range payload.Records {
		ring := NewRingBuffer(rec.Capacity)
		for _, event := range rec.Events {
			ring.Append(event)
		}
		s.records[id] = &record{
			summary: rec.Summary,
			events:  ring,
			updated: rec.Updated,
		}
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
