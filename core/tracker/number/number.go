package number

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Tracker represents a number tracker (e.g., HP, AC, etc.)
type Tracker struct {
	ID      string
	Name    string
	Current int
	Max     int
	Pinned  bool
}

// NewTracker creates a new number tracker
func NewTracker(name string, current, max int) *Tracker {
	return &Tracker{
		ID:      generateID(),
		Name:    name,
		Current: current,
		Max:     max,
		Pinned:  false,
	}
}

// Set sets the current value
func (t *Tracker) Set(value int) {
	t.Current = value
}

// Adjust adjusts the current value by a delta (can be positive or negative)
func (t *Tracker) Adjust(delta int) {
	t.Current += delta
}

// Pin pins the tracker to display
func (t *Tracker) Pin() {
	t.Pinned = true
}

// Unpin unpins the tracker from display
func (t *Tracker) Unpin() {
	t.Pinned = false
}

// String returns a string representation of the tracker
func (t *Tracker) String() string {
	return fmt.Sprintf("[%s] %d/%d", t.Name, t.Current, t.Max)
}

var idCounter uint64

// generateID generates a simple unique ID
func generateID() string {
	id := atomic.AddUint64(&idCounter, 1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), id)
}

