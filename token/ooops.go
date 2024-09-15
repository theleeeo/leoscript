package token

// Order Of OPerations of binary operators, ooops!
type Priority int

const (
	PRIO_LOWEST Priority = iota

	// ==, !=
	PRIO_EQUALS
	// >, <, >=, <=
	PRIO_COMPARISON
	// ||
	PRIO_OR
	// &&
	PRIO_AND
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
