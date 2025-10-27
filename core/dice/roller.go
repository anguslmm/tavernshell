package dice

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
)

// Roll represents the result of a dice roll
type Roll struct {
	Notation string // e.g., "2d6"
	Count    int    // number of dice
	Sides    int    // sides per die
	Results  []int  // individual die results
	Total    int    // sum of all results
}

// Parser for simple dice notation (e.g., "2d6", "d20")
var dicePattern = regexp.MustCompile(`^(\d*)d(\d+)$`)

// Parse parses a dice notation string like "2d6" or "d20"
func Parse(notation string) (count, sides int, err error) {
	matches := dicePattern.FindStringSubmatch(notation)
	if matches == nil {
		return 0, 0, fmt.Errorf("invalid dice notation: %s (expected format: NdS, e.g., 2d6 or d20)", notation)
	}

	// If no count specified (e.g., "d20"), default to 1
	if matches[1] == "" {
		count = 1
	} else {
		count, err = strconv.Atoi(matches[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid dice count: %s", matches[1])
		}
	}

	sides, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid die sides: %s", matches[2])
	}

	// Validation
	if count < 1 {
		return 0, 0, fmt.Errorf("dice count must be at least 1")
	}
	if sides < 2 {
		return 0, 0, fmt.Errorf("die must have at least 2 sides")
	}
	if count > 1000 {
		return 0, 0, fmt.Errorf("dice count too large (max 1000)")
	}

	return count, sides, nil
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

// RollDice rolls the specified dice and returns the results
func RollDice(notation string) (*Roll, error) {
	count, sides, err := Parse(notation)
	if err != nil {
		return nil, err
	}

	results := make([]int, count)
	total := 0

	for i := 0; i < count; i++ {
		roll, err := rollDie(sides)
		if err != nil {
			return nil, err
		}
		results[i] = roll
		total += roll
	}

	return &Roll{
		Notation: notation,
		Count:    count,
		Sides:    sides,
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
