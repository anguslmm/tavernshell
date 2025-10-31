package rotation

import (
	"fmt"
	"sort"
)

// Participant represents a participant in initiative
type Participant struct {
	Name       string
	Initiative int
	IsActive   bool // false if dead/out of combat
}

// Tracker manages initiative order and turn tracking
type Tracker struct {
	Participants []*Participant
	CurrentTurn  int // index into Participants
	Round        int
}

// NewTracker creates a new initiative tracker
func NewTracker() *Tracker {
	return &Tracker{
		Participants: []*Participant{},
		CurrentTurn:  0,
		Round:        1,
	}
}

// Add adds a participant to the initiative
func (t *Tracker) Add(name string, initiative int) {
	p := &Participant{
		Name:       name,
		Initiative: initiative,
		IsActive:   true,
	}
	t.Participants = append(t.Participants, p)
	t.sort()
}

// sort sorts participants by initiative (descending), then alphabetically by name
func (t *Tracker) sort() {
	sort.Slice(t.Participants, func(i, j int) bool {
		if t.Participants[i].Initiative == t.Participants[j].Initiative {
			return t.Participants[i].Name < t.Participants[j].Name
		}
		return t.Participants[i].Initiative > t.Participants[j].Initiative
	})
}

// Next advances to the next active participant's turn
func (t *Tracker) Next() {
	if len(t.Participants) == 0 {
		return
	}

	// Find next active participant
	startIndex := t.CurrentTurn
	for {
		t.CurrentTurn++
		
		// Wrap around to start of list
		if t.CurrentTurn >= len(t.Participants) {
			t.CurrentTurn = 0
			t.Round++
		}

		// If we've looped back to start, break (no active participants)
		if t.CurrentTurn == startIndex {
			break
		}

		// Found an active participant
		if t.Participants[t.CurrentTurn].IsActive {
			break
		}
	}
}

// MarkOut marks a participant as out (dead/incapacitated)
func (t *Tracker) MarkOut(name string) error {
	for _, p := range t.Participants {
		if p.Name == name {
			p.IsActive = false
			return nil
		}
	}
	return fmt.Errorf("participant '%s' not found", name)
}

// MarkIn marks a participant as active again
func (t *Tracker) MarkIn(name string) error {
	for _, p := range t.Participants {
		if p.Name == name {
			p.IsActive = true
			return nil
		}
	}
	return fmt.Errorf("participant '%s' not found", name)
}

// GetCurrent returns the current participant
func (t *Tracker) GetCurrent() *Participant {
	if len(t.Participants) == 0 {
		return nil
	}
	if t.CurrentTurn < 0 || t.CurrentTurn >= len(t.Participants) {
		return nil
	}
	return t.Participants[t.CurrentTurn]
}

// HasParticipants returns true if there are any participants
func (t *Tracker) HasParticipants() bool {
	return len(t.Participants) > 0
}

// ActiveCount returns the number of active participants
func (t *Tracker) ActiveCount() int {
	count := 0
	for _, p := range t.Participants {
		if p.IsActive {
			count++
		}
	}
	return count
}

