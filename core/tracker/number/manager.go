package number

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Manager manages multiple number trackers
type Manager struct {
	trackers map[string]*Tracker // keyed by ID
	mu       sync.RWMutex
}

// NewManager creates a new tracker manager
func NewManager() *Manager {
	return &Manager{
		trackers: make(map[string]*Tracker),
	}
}

// Add adds a new tracker
func (m *Manager) Add(name string, current, max int) *Tracker {
	m.mu.Lock()
	defer m.mu.Unlock()

	tracker := NewTracker(name, current, max)
	m.trackers[tracker.ID] = tracker
	return tracker
}

// Get retrieves a tracker by name (case-insensitive)
func (m *Manager) Get(name string) *Tracker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nameLower := strings.ToLower(name)
	for _, t := range m.trackers {
		if strings.ToLower(t.Name) == nameLower {
			return t
		}
	}
	return nil
}

// Delete removes a tracker by name (case-insensitive)
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	nameLower := strings.ToLower(name)
	for id, t := range m.trackers {
		if strings.ToLower(t.Name) == nameLower {
			delete(m.trackers, id)
			return nil
		}
	}
	return fmt.Errorf("tracker '%s' not found", name)
}

// DeleteAll removes all trackers
func (m *Manager) DeleteAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trackers = make(map[string]*Tracker)
}

// List returns all trackers sorted by name
func (m *Manager) List() []*Tracker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	trackers := make([]*Tracker, 0, len(m.trackers))
	for _, t := range m.trackers {
		trackers = append(trackers, t)
	}

	// Sort by name
	sort.Slice(trackers, func(i, j int) bool {
		return trackers[i].Name < trackers[j].Name
	})

	return trackers
}

// GetPinned returns all pinned trackers sorted by name
func (m *Manager) GetPinned() []*Tracker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pinned []*Tracker
	for _, t := range m.trackers {
		if t.Pinned {
			pinned = append(pinned, t)
		}
	}

	// Sort by name
	sort.Slice(pinned, func(i, j int) bool {
		return pinned[i].Name < pinned[j].Name
	})

	return pinned
}

// Search finds trackers whose names contain the search pattern (case-insensitive)
func (m *Manager) Search(pattern string) []*Tracker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	patternLower := strings.ToLower(pattern)
	var results []*Tracker

	for _, t := range m.trackers {
		if strings.Contains(strings.ToLower(t.Name), patternLower) {
			results = append(results, t)
		}
	}

	// Sort by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// PinAll pins all unpinned trackers
func (m *Manager) PinAll() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, t := range m.trackers {
		if !t.Pinned {
			t.Pinned = true
			count++
		}
	}
	return count
}

// Count returns the total number of trackers
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.trackers)
}

// PinnedCount returns the number of pinned trackers
func (m *Manager) PinnedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, t := range m.trackers {
		if t.Pinned {
			count++
		}
	}
	return count
}
