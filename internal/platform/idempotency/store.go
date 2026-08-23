package idempotency

import (
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	RequestHash string
	Status      int
	Body        []byte
	ExpiresAt   time.Time
}
type Store struct {
	mu      sync.Mutex
	entries map[string]Entry
}

func New() *Store { return &Store{entries: map[string]Entry{}} }
func (s *Store) Lookup(key, hash string, now time.Time) (Entry, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok || now.After(entry.ExpiresAt) {
		delete(s.entries, key)
		return Entry{}, false, nil
	}
	if entry.RequestHash != hash {
		return Entry{}, false, fmt.Errorf("idempotency key reused with another request")
	}
	entry.Body = append([]byte(nil), entry.Body...)
	return entry, true, nil
}
func (s *Store) Save(key string, entry Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry.Body = append([]byte(nil), entry.Body...)
	s.entries[key] = entry
}
