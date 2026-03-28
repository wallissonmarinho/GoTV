package state

import (
	"sync"

	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// MemoryStore implements ports.OutputStore in RAM.
type MemoryStore struct {
	mu   sync.RWMutex
	m3u  []byte
	epg  []byte
	meta ports.OutputMeta
	ok   bool
}

var _ ports.OutputStore = (*MemoryStore)(nil)

// Set stores merged outputs.
func (m *MemoryStore) Set(m3u, epg []byte, meta ports.OutputMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m3u = append([]byte(nil), m3u...)
	m.epg = append([]byte(nil), epg...)
	m.meta = meta
	m.ok = true
}

// Get returns last successful artifacts.
func (m *MemoryStore) Get() (m3u, epg []byte, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.ok {
		return nil, nil, false
	}
	return append([]byte(nil), m.m3u...), append([]byte(nil), m.epg...), true
}

// Meta returns last run metadata.
func (m *MemoryStore) Meta() ports.OutputMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.meta
}
