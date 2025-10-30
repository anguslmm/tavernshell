package dice

import (
	"testing"
)

// TestParse tests the new Parse function that returns an Expression
// (The comprehensive tests are in parser_test.go)
func TestParse(t *testing.T) {
	tests := []struct {
		notation string
		count    int
		sides    int
		wantErr  bool
	}{
		{"2d6", 2, 6, false},
		{"d20", 1, 20, false},
		{"4d8", 4, 8, false},
		{"1d12", 1, 12, false},
		{"10d10", 10, 10, false},
		{"", 0, 0, true},       // invalid: empty
		{"2d", 0, 0, true},     // invalid: no sides
		{"d", 0, 0, true},      // invalid: no sides
		{"2x6", 0, 0, true},    // invalid: wrong separator
		{"0d6", 0, 0, true},    // invalid: zero count
		{"2d1", 0, 0, true},    // invalid: die must have at least 2 sides
		{"1001d6", 0, 0, true}, // invalid: too many dice
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			expr, err := Parse(tt.notation)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.notation, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if expr.Count != tt.count {
					t.Errorf("Parse(%q) count = %v, want %v", tt.notation, expr.Count, tt.count)
				}
				if expr.Sides != tt.sides {
					t.Errorf("Parse(%q) sides = %v, want %v", tt.notation, expr.Sides, tt.sides)
				}
			}
		})
	}
}

func TestRollDice(t *testing.T) {
	tests := []struct {
		notation string
		wantErr  bool
	}{
		{"2d6", false},
		{"d20", false},
		{"4d8", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			result, err := RollDice(tt.notation)
			if (err != nil) != tt.wantErr {
				t.Errorf("RollDice(%q) error = %v, wantErr %v", tt.notation, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result == nil {
					t.Errorf("RollDice(%q) returned nil result", tt.notation)
					return
				}
				if result.Notation != tt.notation {
					t.Errorf("RollDice(%q) notation = %v, want %v", tt.notation, result.Notation, tt.notation)
				}
				if len(result.Results) != result.Count {
					t.Errorf("RollDice(%q) result count = %v, want %v", tt.notation, len(result.Results), result.Count)
				}
				// Verify each die result is in valid range [1, sides]
				for i, roll := range result.Results {
					if roll < 1 || roll > result.Sides {
						t.Errorf("RollDice(%q) result[%d] = %v, want in range [1, %v]", tt.notation, i, roll, result.Sides)
					}
				}
				// Verify total is sum of results
				sum := 0
				for _, roll := range result.Results {
					sum += roll
				}
				if result.Total != sum {
					t.Errorf("RollDice(%q) total = %v, want %v", tt.notation, result.Total, sum)
				}
			}
		})
	}
}

func TestRollDiceRange(t *testing.T) {
	// Test that rolls are within valid range
	// Run multiple times to get better coverage
	for i := 0; i < 100; i++ {
		result, err := RollDice("1d6")
		if err != nil {
			t.Fatalf("RollDice failed: %v", err)
		}
		if len(result.Results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result.Results))
		}
		roll := result.Results[0]
		if roll < 1 || roll > 6 {
			t.Errorf("Roll %d out of range [1, 6]", roll)
		}
	}
}

func TestRollString(t *testing.T) {
	tests := []struct {
		roll Roll
		want string
	}{
		{
			Roll{Notation: "d20", Count: 1, Sides: 20, Results: []int{15}, Total: 15},
			"d20: 15",
		},
		{
			Roll{Notation: "2d6", Count: 2, Sides: 6, Results: []int{3, 4}, Total: 7},
			"2d6: [3 4] = 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.roll.Notation, func(t *testing.T) {
			got := tt.roll.String()
			if got != tt.want {
				t.Errorf("Roll.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
