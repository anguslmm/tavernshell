package dice

import (
	"testing"
)

func TestParse_Basic(t *testing.T) {
	tests := []struct {
		notation string
		want     Expression
		wantErr  bool
	}{
		{"2d6", Expression{Count: 2, Sides: 6}, false},
		{"d20", Expression{Count: 1, Sides: 20}, false},
		{"4d8", Expression{Count: 4, Sides: 8}, false},
		{"1d12", Expression{Count: 1, Sides: 12}, false},
		{"10d10", Expression{Count: 10, Sides: 10}, false},
		{"100d6", Expression{Count: 100, Sides: 6}, false},
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			got, err := Parse(tt.notation)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.notation, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Count != tt.want.Count || got.Sides != tt.want.Sides {
					t.Errorf("Parse(%q) = Count:%d Sides:%d, want Count:%d Sides:%d",
						tt.notation, got.Count, got.Sides, tt.want.Count, tt.want.Sides)
				}
			}
		})
	}
}

func TestParse_Modifiers(t *testing.T) {
	tests := []struct {
		notation string
		want     Expression
	}{
		{"2d6+3", Expression{Count: 2, Sides: 6, Modifier: 3}},
		{"1d20-2", Expression{Count: 1, Sides: 20, Modifier: -2}},
		{"d8+5", Expression{Count: 1, Sides: 8, Modifier: 5}},
		{"3d10-10", Expression{Count: 3, Sides: 10, Modifier: -10}},
		{"2d6+0", Expression{Count: 2, Sides: 6, Modifier: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			got, err := Parse(tt.notation)
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.notation, err)
				return
			}
			if got.Count != tt.want.Count || got.Sides != tt.want.Sides || got.Modifier != tt.want.Modifier {
				t.Errorf("Parse(%q) = Count:%d Sides:%d Modifier:%d, want Count:%d Sides:%d Modifier:%d",
					tt.notation, got.Count, got.Sides, got.Modifier, tt.want.Count, tt.want.Sides, tt.want.Modifier)
			}
		})
	}
}

func TestParse_Advantage(t *testing.T) {
	tests := []struct {
		notation string
		want     Expression
	}{
		{"d20!", Expression{Count: 1, Sides: 20, Advantage: true}},
		{"2d6!", Expression{Count: 2, Sides: 6, Advantage: true}},
		{"d20!+5", Expression{Count: 1, Sides: 20, Advantage: true, Modifier: 5}},
		{"4d10!-2", Expression{Count: 4, Sides: 10, Advantage: true, Modifier: -2}},
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			got, err := Parse(tt.notation)
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.notation, err)
				return
			}
			if got.Count != tt.want.Count || got.Sides != tt.want.Sides ||
				got.Advantage != tt.want.Advantage || got.Modifier != tt.want.Modifier {
				t.Errorf("Parse(%q) = Count:%d Sides:%d Adv:%v Mod:%d, want Count:%d Sides:%d Adv:%v Mod:%d",
					tt.notation, got.Count, got.Sides, got.Advantage, got.Modifier,
					tt.want.Count, tt.want.Sides, tt.want.Advantage, tt.want.Modifier)
			}
		})
	}
}

func TestParse_Operations(t *testing.T) {
	tests := []struct {
		notation string
		wantOp   OpType
		wantCnt  int
	}{
		{"4d6kh3", OpKeepHighest, 3},
		{"4d6kl2", OpKeepLowest, 2},
		{"5d10dh1", OpDropHighest, 1},
		{"4d6dl1", OpDropLowest, 1},
		{"10d6kh5", OpKeepHighest, 5},
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			got, err := Parse(tt.notation)
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.notation, err)
				return
			}
			if got.Operation == nil {
				t.Errorf("Parse(%q) Operation is nil", tt.notation)
				return
			}
			if got.Operation.Type != tt.wantOp || got.Operation.Count != tt.wantCnt {
				t.Errorf("Parse(%q) Operation = Type:%v Count:%d, want Type:%v Count:%d",
					tt.notation, got.Operation.Type, got.Operation.Count, tt.wantOp, tt.wantCnt)
			}
		})
	}
}

func TestParse_Combined(t *testing.T) {
	tests := []struct {
		notation string
		desc     string
	}{
		{"4d6kh3+2", "keep highest with modifier"},
		{"2d20kh1", "advantage via keep highest"},
		{"d20!+5", "advantage with modifier"},
		{"4d10!kh3", "advantage with keep operation"},
		{"3d8dl1-1", "drop lowest with negative modifier"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := Parse(tt.notation)
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.notation, err)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		notation string
		desc     string
	}{
		{"", "empty string"},
		{"2d", "missing sides"},
		{"d", "missing sides"},
		{"2x6", "wrong separator"},
		{"0d6", "zero count"},
		{"2d1", "die with 1 side"},
		{"1001d6", "too many dice"},
		{"2d6kh", "operation without count"},
		{"2d6+", "modifier without value"},
		{"2d6xyz", "invalid operation"},
		{"2d6kh0", "operation count zero"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := Parse(tt.notation)
			if err == nil {
				t.Errorf("Parse(%q) expected error for %s, got nil", tt.notation, tt.desc)
			}
		})
	}
}

func TestParse_Whitespace(t *testing.T) {
	tests := []struct {
		notation string
		want     Expression
	}{
		{"2d6 + 3", Expression{Count: 2, Sides: 6, Modifier: 3}},
		{"d20 !", Expression{Count: 1, Sides: 20, Advantage: true}},
		{" 4d6kh3 ", Expression{Count: 4, Sides: 6}}, // Operation will be non-nil
		{"2 d 6", Expression{Count: 2, Sides: 6}},
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			got, err := Parse(tt.notation)
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.notation, err)
				return
			}
			// Just verify it parses without error
			if got.Count != tt.want.Count || got.Sides != tt.want.Sides {
				t.Errorf("Parse(%q) = Count:%d Sides:%d, want Count:%d Sides:%d",
					tt.notation, got.Count, got.Sides, tt.want.Count, tt.want.Sides)
			}
		})
	}
}

