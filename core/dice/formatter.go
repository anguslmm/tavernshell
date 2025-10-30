package dice

import (
	"fmt"
	"strings"
)

// String returns a formatted string representation of the result
func (r *Result) String() string {
	if r == nil {
		return "<nil result>"
	}

	var b strings.Builder

	// Build the dice list with kept/dropped indication
	keptDice := []string{}
	droppedDice := []string{}
	for _, die := range r.Rolls {
		dieStr := fmt.Sprintf("%d", die.Value)
		if die.Kept {
			keptDice = append(keptDice, dieStr)
		} else {
			droppedDice = append(droppedDice, dieStr)
		}
	}

	// Reconstruct notation
	notation := fmt.Sprintf("%dd%d", r.Expression.Count, r.Expression.Sides)
	if r.Expression.Advantage {
		notation += "!"
	}
	if r.Expression.Operation != nil {
		notation += formatOperation(r.Expression.Operation)
	}
	if r.Expression.Modifier != 0 {
		if r.Expression.Modifier > 0 {
			notation += fmt.Sprintf("+%d", r.Expression.Modifier)
		} else {
			notation += fmt.Sprintf("%d", r.Expression.Modifier)
		}
	}

	b.WriteString(notation)
	b.WriteString(": ")

	// Show all dice (kept and dropped)
	b.WriteString("[")
	
	// Interleave kept and dropped dice in order as they were rolled
	diceStrs := make([]string, len(r.Rolls))
	for i, die := range r.Rolls {
		if die.Kept {
			diceStrs[i] = fmt.Sprintf("%d", die.Value)
		} else {
			// Mark dropped dice with angle brackets
			diceStrs[i] = fmt.Sprintf("‹%d›", die.Value)
		}
	}
	b.WriteString(strings.Join(diceStrs, ", "))
	b.WriteString("]")

	// Show modifier if present
	if r.Expression.Modifier != 0 {
		if r.Expression.Modifier > 0 {
			b.WriteString(fmt.Sprintf(" +%d", r.Expression.Modifier))
		} else {
			b.WriteString(fmt.Sprintf(" %d", r.Expression.Modifier))
		}
	}

	// Show total
	b.WriteString(fmt.Sprintf(" = %d", r.Total))

	// Add description if there are dropped dice
	if len(droppedDice) > 0 {
		b.WriteString(" (")
		if r.Expression.Advantage {
			b.WriteString("advantage")
		} else if r.Expression.Operation != nil {
			b.WriteString(formatOperationDescription(r.Expression.Operation))
		}
		b.WriteString(")")
	}

	return b.String()
}

// formatOperation formats an operation for display in notation
func formatOperation(op *Operation) string {
	switch op.Type {
	case OpKeepHighest:
		return fmt.Sprintf("kh%d", op.Count)
	case OpKeepLowest:
		return fmt.Sprintf("kl%d", op.Count)
	case OpDropHighest:
		return fmt.Sprintf("dh%d", op.Count)
	case OpDropLowest:
		return fmt.Sprintf("dl%d", op.Count)
	default:
		return ""
	}
}

// formatOperationDescription returns a human-readable description of an operation
func formatOperationDescription(op *Operation) string {
	switch op.Type {
	case OpKeepHighest:
		return fmt.Sprintf("kept highest %d", op.Count)
	case OpKeepLowest:
		return fmt.Sprintf("kept lowest %d", op.Count)
	case OpDropHighest:
		return fmt.Sprintf("dropped highest %d", op.Count)
	case OpDropLowest:
		return fmt.Sprintf("dropped lowest %d", op.Count)
	default:
		return ""
	}
}

