package liveflux

import "sync"

// Store defines how component instances are persisted between requests.
type Store interface {
	Get(id string) (Component, bool)
	Set(c Component)
	Delete(id string)
}

// MemoryStore is a simple in-memory implementation suitable for development
// and single-instance deployments. Replace with a session or DB-backed
// implementation for multi-instance deployments.
type MemoryStore struct {
	mu sync.RWMutex
	m  map[string]Component
}

// NewMemoryStore creates a MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{m: map[string]Component{}}
}

// Get returns a component by id.
func (s *MemoryStore) Get(id string) (Component, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.m[id]
	return c, ok
}

// Set stores a component by its ID.
func (s *MemoryStore) Set(c Component) {
	if c == nil || c.GetID() == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[c.GetID()] = c
}

// Delete removes a component by id.
func (s *MemoryStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
}

// StoreDefault is the default process-local store used by the handler.
var StoreDefault Store = NewMemoryStore()
