package token

// Use the same orders as in C: https://en.cppreference.com/w/c/language/operator_precedence

// Order Of OPerations of binary operators, ooops!
type Priority int

const (
	_ Priority = iota

	PRIO_ASSIGN     // =, +=, -=, *=, /=
	PRIO_OR         // ||
	PRIO_AND        // &&
	PRIO_EQUALS     // ==, !=
	PRIO_COMPARISON // >, <, >=, <=
	PRIO_SUM        // + and -
	PRIO_PRODUCT    // * and /

	PRIO_PAREN Priority = 100 // ()
)
