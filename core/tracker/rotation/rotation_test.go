package rotation

import "testing"

func TestNewTracker(t *testing.T) {
	tracker := NewTracker()
	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}
	if tracker.Round != 1 {
		t.Errorf("Expected round 1, got %d", tracker.Round)
	}
	if len(tracker.Participants) != 0 {
		t.Errorf("Expected 0 participants, got %d", len(tracker.Participants))
	}
}

func TestAddParticipant(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Fighter", 18)
	tracker.Add("Goblin", 12)
	tracker.Add("Wizard", 15)

	if len(tracker.Participants) != 3 {
		t.Fatalf("Expected 3 participants, got %d", len(tracker.Participants))
	}

	// Check sorting (descending by initiative)
	expected := []struct {
		name       string
		initiative int
	}{
		{"Fighter", 18},
		{"Wizard", 15},
		{"Goblin", 12},
	}

	for i, exp := range expected {
		if tracker.Participants[i].Name != exp.name {
			t.Errorf("Position %d: expected %s, got %s", i, exp.name, tracker.Participants[i].Name)
		}
		if tracker.Participants[i].Initiative != exp.initiative {
			t.Errorf("Position %d: expected initiative %d, got %d", i, exp.initiative, tracker.Participants[i].Initiative)
		}
		if !tracker.Participants[i].IsActive {
			t.Errorf("Position %d: expected active participant", i)
		}
	}
}

func TestSortingWithTies(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Zebra", 15)
	tracker.Add("Apple", 15)
	tracker.Add("Middle", 15)

	// With same initiative, should sort alphabetically
	if tracker.Participants[0].Name != "Apple" {
		t.Errorf("Expected Apple first, got %s", tracker.Participants[0].Name)
	}
	if tracker.Participants[1].Name != "Middle" {
		t.Errorf("Expected Middle second, got %s", tracker.Participants[1].Name)
	}
	if tracker.Participants[2].Name != "Zebra" {
		t.Errorf("Expected Zebra third, got %s", tracker.Participants[2].Name)
	}
}

func TestNext(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Fighter", 18)
	tracker.Add("Wizard", 15)
	tracker.Add("Goblin", 12)

	// Start at Fighter (index 0)
	if tracker.CurrentTurn != 0 {
		t.Errorf("Expected starting turn 0, got %d", tracker.CurrentTurn)
	}

	// Next -> Wizard
	tracker.Next()
	if tracker.CurrentTurn != 1 {
		t.Errorf("Expected turn 1, got %d", tracker.CurrentTurn)
	}

	// Next -> Goblin
	tracker.Next()
	if tracker.CurrentTurn != 2 {
		t.Errorf("Expected turn 2, got %d", tracker.CurrentTurn)
	}

	// Next -> should wrap to Fighter and increment round
	tracker.Next()
	if tracker.CurrentTurn != 0 {
		t.Errorf("Expected turn 0 after wrap, got %d", tracker.CurrentTurn)
	}
	if tracker.Round != 2 {
		t.Errorf("Expected round 2 after wrap, got %d", tracker.Round)
	}
}

func TestNextSkipsInactive(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Fighter", 18)
	tracker.Add("Wizard", 15)
	tracker.Add("Goblin", 12)

	// Mark Wizard as inactive
	tracker.MarkOut("Wizard")

	// Start at Fighter
	if tracker.CurrentTurn != 0 {
		t.Errorf("Expected starting turn 0, got %d", tracker.CurrentTurn)
	}

	// Next should skip Wizard and go to Goblin
	tracker.Next()
	if tracker.CurrentTurn != 2 {
		t.Errorf("Expected turn 2 (Goblin), got %d", tracker.CurrentTurn)
	}
	if tracker.GetCurrent().Name != "Goblin" {
		t.Errorf("Expected Goblin, got %s", tracker.GetCurrent().Name)
	}
}

func TestMarkOut(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Fighter", 18)
	tracker.Add("Goblin", 12)

	err := tracker.MarkOut("Goblin")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Find Goblin and check it's inactive
	var goblin *Participant
	for _, p := range tracker.Participants {
		if p.Name == "Goblin" {
			goblin = p
			break
		}
	}
	if goblin == nil {
		t.Fatal("Goblin not found")
	}
	if goblin.IsActive {
		t.Error("Expected Goblin to be inactive")
	}

	// Try to mark out non-existent participant
	err = tracker.MarkOut("Dragon")
	if err == nil {
		t.Error("Expected error for non-existent participant")
	}
}

func TestGetCurrent(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Fighter", 18)
	tracker.Add("Wizard", 15)

	current := tracker.GetCurrent()
	if current == nil {
		t.Fatal("Expected non-nil current participant")
	}
	if current.Name != "Fighter" {
		t.Errorf("Expected Fighter, got %s", current.Name)
	}

	tracker.Next()
	current = tracker.GetCurrent()
	if current.Name != "Wizard" {
		t.Errorf("Expected Wizard, got %s", current.Name)
	}
}

func TestActiveCount(t *testing.T) {
	tracker := NewTracker()
	tracker.Add("Fighter", 18)
	tracker.Add("Wizard", 15)
	tracker.Add("Goblin", 12)

	if tracker.ActiveCount() != 3 {
		t.Errorf("Expected 3 active, got %d", tracker.ActiveCount())
	}

	tracker.MarkOut("Goblin")
	if tracker.ActiveCount() != 2 {
		t.Errorf("Expected 2 active after marking out, got %d", tracker.ActiveCount())
	}
}

