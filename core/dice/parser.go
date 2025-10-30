package dice

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Parse parses a dice notation string into an Expression
// Supports: XdY, XdY+Z, XdY!, XdYkhN, XdYdlN, etc.
func Parse(notation string) (*Expression, error) {
	if notation == "" {
		return nil, fmt.Errorf("empty dice notation")
	}

	// Remove all whitespace for easier parsing
	notation = strings.ReplaceAll(notation, " ", "")
	notation = strings.ToLower(notation) // Case-insensitive

	expr := &Expression{
		Count: 1, // Default to 1 die
	}

	i := 0

	// Step 1: Parse optional count (digits before 'd')
	start := i
	for i < len(notation) && unicode.IsDigit(rune(notation[i])) {
		i++
	}
	if i > start {
		count, err := strconv.Atoi(notation[start:i])
		if err != nil {
			return nil, fmt.Errorf("invalid die count")
		}
		if count < 1 {
			return nil, fmt.Errorf("die count must be at least 1")
		}
		if count > 1000 {
			return nil, fmt.Errorf("die count too large (max 1000)")
		}
		expr.Count = count
	}

	// Step 2: Parse 'd'
	if i >= len(notation) || notation[i] != 'd' {
		return nil, fmt.Errorf("expected 'd' at position %d", i)
	}
	i++

	// Step 3: Parse sides (digits after 'd')
	if i >= len(notation) || !unicode.IsDigit(rune(notation[i])) {
		return nil, fmt.Errorf("expected number of sides after 'd'")
	}
	start = i
	for i < len(notation) && unicode.IsDigit(rune(notation[i])) {
		i++
	}
	sides, err := strconv.Atoi(notation[start:i])
	if err != nil {
		return nil, fmt.Errorf("invalid die sides")
	}
	if sides < 2 {
		return nil, fmt.Errorf("die must have at least 2 sides")
	}
	expr.Sides = sides

	// Step 4: Parse optional advantage (!), operation (kh/kl/dh/dl), and modifier (+/-)
	for i < len(notation) {
		ch := notation[i]

		switch ch {
		case '!':
			// Advantage
			expr.Advantage = true
			i++

		case 'k', 'd':
			// Operation (kh, kl, dh, dl)
			if i+1 >= len(notation) {
				return nil, fmt.Errorf("incomplete operation at position %d", i)
			}
			op := notation[i : i+2]
			var opType OpType
			switch op {
			case "kh":
				opType = OpKeepHighest
			case "kl":
				opType = OpKeepLowest
			case "dh":
				opType = OpDropHighest
			case "dl":
				opType = OpDropLowest
			default:
				return nil, fmt.Errorf("unknown operation: %s (expected kh, kl, dh, or dl)", op)
			}
			i += 2

			// Parse count after operation
			if i >= len(notation) || !unicode.IsDigit(rune(notation[i])) {
				return nil, fmt.Errorf("expected number after operation %s", op)
			}
			start = i
			for i < len(notation) && unicode.IsDigit(rune(notation[i])) {
				i++
			}
			count, err := strconv.Atoi(notation[start:i])
			if err != nil {
				return nil, fmt.Errorf("invalid operation count")
			}
			if count < 1 {
				return nil, fmt.Errorf("operation count must be at least 1")
			}

			expr.Operation = &Operation{
				Type:  opType,
				Count: count,
			}

		case '+', '-':
			// Modifier
			sign := ch
			i++
			if i >= len(notation) || !unicode.IsDigit(rune(notation[i])) {
				return nil, fmt.Errorf("expected number after '%c'", sign)
			}
			start = i
			for i < len(notation) && unicode.IsDigit(rune(notation[i])) {
				i++
			}
			modifier, err := strconv.Atoi(notation[start:i])
			if err != nil {
				return nil, fmt.Errorf("invalid modifier")
			}
			if sign == '-' {
				modifier = -modifier
			}
			expr.Modifier = modifier

		default:
			return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, i)
		}
	}

	return expr, nil
}

