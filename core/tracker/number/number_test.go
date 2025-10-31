package number

import "testing"

func TestNewTracker(t *testing.T) {
	tracker := NewTracker("HP", 45, 50)
	
	if tracker.Name != "HP" {
		t.Errorf("Expected name 'HP', got '%s'", tracker.Name)
	}
	if tracker.Current != 45 {
		t.Errorf("Expected current 45, got %d", tracker.Current)
	}
	if tracker.Max != 50 {
		t.Errorf("Expected max 50, got %d", tracker.Max)
	}
	if tracker.Pinned {
		t.Error("Expected tracker to not be pinned by default")
	}
	if tracker.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestSet(t *testing.T) {
	tracker := NewTracker("HP", 45, 50)
	tracker.Set(30)
	
	if tracker.Current != 30 {
		t.Errorf("Expected current 30, got %d", tracker.Current)
	}
}

func TestAdjust(t *testing.T) {
	tracker := NewTracker("HP", 45, 50)
	
	// Subtract (take damage)
	tracker.Adjust(-10)
	if tracker.Current != 35 {
		t.Errorf("Expected current 35 after -10, got %d", tracker.Current)
	}
	
	// Add (heal)
	tracker.Adjust(5)
	if tracker.Current != 40 {
		t.Errorf("Expected current 40 after +5, got %d", tracker.Current)
	}
}

func TestPin(t *testing.T) {
	tracker := NewTracker("HP", 45, 50)
	
	if tracker.Pinned {
		t.Error("Expected tracker to not be pinned initially")
	}
	
	tracker.Pin()
	if !tracker.Pinned {
		t.Error("Expected tracker to be pinned after Pin()")
	}
	
	tracker.Unpin()
	if tracker.Pinned {
		t.Error("Expected tracker to not be pinned after Unpin()")
	}
}

func TestString(t *testing.T) {
	tracker := NewTracker("HP", 35, 45)
	expected := "[HP] 35/45"
	
	if tracker.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, tracker.String())
	}
}

func TestManagerAdd(t *testing.T) {
	manager := NewManager()
	tracker := manager.Add("HP", 45, 50)
	
	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}
	if manager.Count() != 1 {
		t.Errorf("Expected count 1, got %d", manager.Count())
	}
}

func TestManagerGet(t *testing.T) {
	manager := NewManager()
	manager.Add("HP", 45, 50)
	manager.Add("AC", 18, 18)
	
	tracker := manager.Get("HP")
	if tracker == nil {
		t.Fatal("Expected to find HP tracker")
	}
	if tracker.Name != "HP" {
		t.Errorf("Expected HP, got %s", tracker.Name)
	}
	
	// Test case-insensitive
	tracker = manager.Get("hp")
	if tracker == nil {
		t.Fatal("Expected to find tracker with lowercase 'hp'")
	}
	
	// Test not found
	tracker = manager.Get("Nonexistent")
	if tracker != nil {
		t.Error("Expected nil for nonexistent tracker")
	}
}

func TestManagerDelete(t *testing.T) {
	manager := NewManager()
	manager.Add("HP", 45, 50)
	manager.Add("AC", 18, 18)
	
	err := manager.Delete("HP")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if manager.Count() != 1 {
		t.Errorf("Expected count 1 after delete, got %d", manager.Count())
	}
	
	tracker := manager.Get("HP")
	if tracker != nil {
		t.Error("Expected HP tracker to be deleted")
	}
	
	// Try to delete non-existent
	err = manager.Delete("Nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent tracker")
	}
}

func TestManagerDeleteAll(t *testing.T) {
	manager := NewManager()
	manager.Add("HP", 45, 50)
	manager.Add("AC", 18, 18)
	manager.Add("Speed", 30, 30)
	
	manager.DeleteAll()
	
	if manager.Count() != 0 {
		t.Errorf("Expected count 0 after DeleteAll, got %d", manager.Count())
	}
}

func TestManagerList(t *testing.T) {
	manager := NewManager()
	manager.Add("HP", 45, 50)
	manager.Add("AC", 18, 18)
	manager.Add("Speed", 30, 30)
	
	list := manager.List()
	if len(list) != 3 {
		t.Fatalf("Expected 3 trackers, got %d", len(list))
	}
	
	// Check sorting (alphabetical)
	expected := []string{"AC", "HP", "Speed"}
	for i, name := range expected {
		if list[i].Name != name {
			t.Errorf("Position %d: expected %s, got %s", i, name, list[i].Name)
		}
	}
}

func TestManagerGetPinned(t *testing.T) {
	manager := NewManager()
	hp := manager.Add("HP", 45, 50)
	ac := manager.Add("AC", 18, 18)
	manager.Add("Speed", 30, 30)
	
	hp.Pin()
	ac.Pin()
	
	pinned := manager.GetPinned()
	if len(pinned) != 2 {
		t.Fatalf("Expected 2 pinned trackers, got %d", len(pinned))
	}
	
	// Check sorting
	if pinned[0].Name != "AC" || pinned[1].Name != "HP" {
		t.Error("Expected AC, HP in that order")
	}
}

func TestManagerSearch(t *testing.T) {
	manager := NewManager()
	manager.Add("HP", 45, 50)
	manager.Add("Hit Points", 30, 40)
	manager.Add("AC", 18, 18)
	
	results := manager.Search("hp")
	if len(results) != 1 {
		t.Fatalf("Expected 1 result for 'hp', got %d", len(results))
	}
	if results[0].Name != "HP" {
		t.Errorf("Expected HP, got %s", results[0].Name)
	}
	
	results = manager.Search("H")
	if len(results) != 2 {
		t.Fatalf("Expected 2 results for 'H', got %d", len(results))
	}
}

func TestManagerPinAll(t *testing.T) {
	manager := NewManager()
	hp := manager.Add("HP", 45, 50)
	manager.Add("AC", 18, 18)
	manager.Add("Speed", 30, 30)
	
	hp.Pin()
	
	count := manager.PinAll()
	if count != 2 {
		t.Errorf("Expected 2 trackers pinned, got %d", count)
	}
	
	if manager.PinnedCount() != 3 {
		t.Errorf("Expected 3 total pinned, got %d", manager.PinnedCount())
	}
}

