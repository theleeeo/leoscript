package token

// Use the same orders as in C: https://en.cppreference.com/w/c/language/operator_precedence

// Order Of OPerations of binary operators, ooops!
type Priority int

const (
	PRIO_LOWEST Priority = iota

	// ||
	PRIO_OR
	// &&
	PRIO_AND
	// ==, !=
	PRIO_EQUALS
	// >, <, >=, <=
	PRIO_COMPARISON
	// + and -
	PRIO_SUM
	// * and /
	PRIO_PRODUCT
	// // unary operators
	// PREFIX
	// // function calls
	// CALL

	// ()
	PRIO_PAREN Priority = 100
)
