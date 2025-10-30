package dice

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sort"
)

// Roll represents the result of a dice roll (legacy/backward compatibility)
type Roll struct {
	Notation string // e.g., "2d6"
	Count    int    // number of dice
	Sides    int    // sides per die
	Results  []int  // individual die results
	Total    int    // sum of all results
}

// rollDie rolls a single die with the specified number of sides
// Uses crypto/rand for cryptographically secure random numbers
func rollDie(sides int) (int, error) {
	if sides < 2 {
		return 0, fmt.Errorf("die must have at least 2 sides")
	}

	// Generate a random number using crypto/rand
	var randomBytes [8]byte
	_, err := rand.Read(randomBytes[:])
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}

	// Convert bytes to uint64
	randomNum := binary.BigEndian.Uint64(randomBytes[:])

	// Use modulo to get a number in range [0, sides-1], then add 1
	// This is slightly biased, but the bias is negligible for crypto/rand
	result := int(randomNum%uint64(sides)) + 1

	return result, nil
}

// RollExpression rolls dice according to an Expression and returns a Result
func RollExpression(expr *Expression) (*Result, error) {
	if expr == nil {
		return nil, fmt.Errorf("nil expression")
	}

	// Determine how many dice to actually roll
	diceToRoll := expr.Count
	if expr.Advantage {
		diceToRoll *= 2 // Per-die advantage: roll each die twice
	}

	// Roll all the dice
	rolls := make([]Die, diceToRoll)
	for i := 0; i < diceToRoll; i++ {
		value, err := rollDie(expr.Sides)
		if err != nil {
			return nil, err
		}
		rolls[i] = Die{
			Value: value,
			Sides: expr.Sides,
			Kept:  true, // Initially all dice are kept
		}
	}

	// Apply per-die advantage if needed
	if expr.Advantage {
		// Pair consecutive dice and keep the highest from each pair
		for i := 0; i < len(rolls); i += 2 {
			if rolls[i].Value >= rolls[i+1].Value {
				rolls[i+1].Kept = false // Drop the lower one
			} else {
				rolls[i].Kept = false // Drop the lower one
			}
		}
	}

	// Apply keep/drop operation if specified
	if expr.Operation != nil {
		applyOperation(rolls, expr.Operation)
	}

	// Calculate totals
	keptTotal := 0
	for _, die := range rolls {
		if die.Kept {
			keptTotal += die.Value
		}
	}
	total := keptTotal + expr.Modifier

	return &Result{
		Expression: expr,
		Rolls:      rolls,
		KeptTotal:  keptTotal,
		Total:      total,
	}, nil
}

// applyOperation applies a keep/drop operation to rolled dice
func applyOperation(rolls []Die, op *Operation) {
	// Only operate on dice that are currently kept
	keptIndices := []int{}
	for i, die := range rolls {
		if die.Kept {
			keptIndices = append(keptIndices, i)
		}
	}

	if len(keptIndices) == 0 {
		return
	}

	// Sort the kept dice by value
	sort.Slice(keptIndices, func(i, j int) bool {
		return rolls[keptIndices[i]].Value < rolls[keptIndices[j]].Value
	})

	// Determine which dice to keep/drop based on operation type
	indicesToDrop := []int{}

	switch op.Type {
	case OpKeepHighest:
		// Drop the lowest (len - count) dice
		if op.Count < len(keptIndices) {
			dropCount := len(keptIndices) - op.Count
			indicesToDrop = keptIndices[:dropCount]
		}

	case OpKeepLowest:
		// Drop the highest (len - count) dice
		if op.Count < len(keptIndices) {
			dropCount := len(keptIndices) - op.Count
			indicesToDrop = keptIndices[len(keptIndices)-dropCount:]
		}

	case OpDropHighest:
		// Drop the highest count dice
		if op.Count <= len(keptIndices) {
			indicesToDrop = keptIndices[len(keptIndices)-op.Count:]
		}

	case OpDropLowest:
		// Drop the lowest count dice
		if op.Count <= len(keptIndices) {
			indicesToDrop = keptIndices[:op.Count]
		}
	}

	// Mark the selected dice as dropped
	for _, idx := range indicesToDrop {
		rolls[idx].Kept = false
	}
}

// RollDice rolls the specified dice and returns the results (backward compatibility)
func RollDice(notation string) (*Roll, error) {
	expr, err := Parse(notation)
	if err != nil {
		return nil, err
	}

	result, err := RollExpression(expr)
	if err != nil {
		return nil, err
	}

	// Convert Result to legacy Roll format
	results := []int{}
	total := 0
	for _, die := range result.Rolls {
		if die.Kept {
			results = append(results, die.Value)
			total += die.Value
		}
	}
	total += expr.Modifier

	return &Roll{
		Notation: notation,
		Count:    expr.Count,
		Sides:    expr.Sides,
		Results:  results,
		Total:    total,
	}, nil
}

// String returns a formatted string representation of the roll
func (r *Roll) String() string {
	if len(r.Results) == 1 {
		return fmt.Sprintf("%s: %d", r.Notation, r.Total)
	}
	return fmt.Sprintf("%s: %v = %d", r.Notation, r.Results, r.Total)
}
