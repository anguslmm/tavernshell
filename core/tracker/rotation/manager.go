package rotation

import "sync"

// Manager manages the initiative tracker state
type Manager struct {
	tracker *Tracker
	active  bool
	mu      sync.RWMutex
}

// NewManager creates a new initiative manager
func NewManager() *Manager {
	return &Manager{
		tracker: nil,
		active:  false,
	}
}

// Start starts a new initiative session
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tracker = NewTracker()
	m.active = true
}

// End ends the current initiative session
func (m *Manager) End() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tracker = nil
	m.active = false
}

// IsActive returns true if initiative is currently active
func (m *Manager) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active
}

// GetTracker returns the current tracker (may be nil)
func (m *Manager) GetTracker() *Tracker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tracker
}

// Add adds a participant to the current initiative
func (m *Manager) Add(name string, initiative int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tracker != nil {
		m.tracker.Add(name, initiative)
	}
}

// Next advances to the next turn
func (m *Manager) Next() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tracker != nil {
		m.tracker.Next()
	}
}

// MarkOut marks a participant as out
func (m *Manager) MarkOut(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tracker != nil {
		return m.tracker.MarkOut(name)
	}
	return nil
}

// MarkIn marks a participant as active again
func (m *Manager) MarkIn(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tracker != nil {
		return m.tracker.MarkIn(name)
	}
	return nil
}

