package parser

import "leoscript/token"

type Expression interface{}

type IntegerLiteral struct {
	Value int
}

type BooleanLiteral struct {
	Value bool
}

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Op       string
	priority token.Priority
}

// PriorityMerge will merge the current binary expression with a new expression based on the priorities of the operators
// A new expression tree will be returned with the order of operations handled correctly.
func (root BinaryExpression) PriorityMerge(binTk token.Operator, newExpr Expression) Expression {
	priority := binTk.Priority()

	// If the new priority is lower, it should be higher in the expression tree to be evaluated later.
	// If it is the same, it should also be higher to preserve left-to-right evaluation.
	if priority <= root.priority {
		// No priority swap needed, create a new root expression
		return BinaryExpression{
			Left:     root,
			Right:    newExpr,
			Op:       binTk.Op,
			priority: priority,
		}
	}

	// The right side of the root binary expression is also a binary expression
	// We need to also do a priority merge on that to support multiple layers of priority
	if rbin, ok := root.Right.(BinaryExpression); ok {
		root.Right = rbin.PriorityMerge(binTk, newExpr)
		return root
	}

	// We have a higher priority operator, so we need to swap the root right side
	newRight := BinaryExpression{
		Left:     root.Right,
		Right:    newExpr,
		Op:       binTk.Op,
		priority: priority,
	}
	root.Right = newRight

	return root
}

type UnaryExpression struct {
	Expression Expression
	Op         string
}
