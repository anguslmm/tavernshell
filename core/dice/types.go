package dice

// Expression represents a dice notation to be rolled (input)
type Expression struct {
	Count     int        // Number of dice to roll
	Sides     int        // Sides per die
	Modifier  int        // +/- modifier to add to total
	Operation *Operation // Optional keep/drop operation
	Advantage bool       // Per-die advantage (roll each die twice, keep highest)
}

// Operation represents a keep/drop operation on rolled dice
type Operation struct {
	Type  OpType // Type of operation
	Count int    // How many dice to keep/drop
}

// OpType represents the type of keep/drop operation
type OpType int

const (
	OpKeepHighest OpType = iota // Keep the N highest dice
	OpKeepLowest                 // Keep the N lowest dice
	OpDropHighest                // Drop the N highest dice
	OpDropLowest                 // Drop the N lowest dice
)

// Result represents the outcome of rolling dice (output)
type Result struct {
	Expression *Expression // Original expression that created this result
	Rolls      []Die       // All dice rolled (including dropped)
	KeptTotal  int         // Sum of kept dice only
	Total      int         // Final result (kept + modifier)
}

// Die represents a single rolled die
type Die struct {
	Value int  // The rolled value
	Sides int  // Number of sides on this die
	Kept  bool // Whether this die counts toward the total
}

