package timer

import (
	"sort"
	"sync"
)

// Manager manages multiple timers
type Manager struct {
	timers map[string]*Timer
	mu     sync.RWMutex
}

// NewManager creates a new timer manager
func NewManager() *Manager {
	return &Manager{
		timers: make(map[string]*Timer),
	}
}

// Add adds a new timer to the manager
func (m *Manager) Add(timer *Timer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timers[timer.ID] = timer
}

// Remove removes a timer by ID
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.timers, id)
}

// GetActive returns all active (non-expired) timers sorted by remaining time (ascending)
func (m *Manager) GetActive() []*Timer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*Timer
	for _, timer := range m.timers {
		if !timer.IsExpired() {
			active = append(active, timer)
		}
	}

	// Sort by remaining time (shortest first)
	sort.Slice(active, func(i, j int) bool {
		return active[i].Remaining() < active[j].Remaining()
	})

	return active
}

// GetExpired returns all expired timers and removes them from the manager
func (m *Manager) GetExpired() []*Timer {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expired []*Timer
	for id, timer := range m.timers {
		if timer.IsExpired() {
			expired = append(expired, timer)
			delete(m.timers, id)
		}
	}

	return expired
}

// Count returns the total number of timers (including expired)
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.timers)
}

// ActiveCount returns the number of active (non-expired) timers
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, timer := range m.timers {
		if !timer.IsExpired() {
			count++
		}
	}
	return count
}

// Clear removes all timers
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timers = make(map[string]*Timer)
}

