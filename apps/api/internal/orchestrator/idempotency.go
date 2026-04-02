package orchestrator

import (
	"sync"
	"time"
)

type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]IdempotencyEntry
	ttl     time.Duration
}

type IdempotencyEntry struct {
	Response    any
	StatusCode  int
	CreatedAt   time.Time
	CompletedAt time.Time
}

func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	store := &IdempotencyStore{
		entries: make(map[string]IdempotencyEntry),
		ttl:     ttl,
	}
	go store.cleanup()
	return store
}

func (s *IdempotencyStore) Get(key string) (any, int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.entries[key]
	if !exists {
		return nil, 0, false
	}

	if time.Since(entry.CreatedAt) > s.ttl {
		return nil, 0, false
	}

	return entry.Response, entry.StatusCode, true
}

func (s *IdempotencyStore) Set(key string, response any, statusCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = IdempotencyEntry{
		Response:    response,
		StatusCode:  statusCode,
		CreatedAt:   time.Now(),
		CompletedAt: time.Now(),
	}
}

func (s *IdempotencyStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.entries {
			if now.Sub(entry.CreatedAt) > s.ttl {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}
