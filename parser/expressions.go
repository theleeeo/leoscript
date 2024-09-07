package parser

import (
	"fmt"
	"leoscript/token"
)

// Start parsing an expression, beginning with the current.
func (p *Parser) parseExpression() (Expression, error) {
	var root Expression

	for tk := p.peek(); tk != nil; tk = p.peek() {
		fmt.Printf("tk: %T\n", tk)
		// This will only hit if it is the first token in the expression.
		// Otherwise it will have been handled by another branch.
		if intTk, ok := tk.(token.Integer); ok {
			if root != nil {
				return nil, fmt.Errorf("unexpected integer token: %v", intTk)
			}
			root = IntegerLiteral{Value: intTk.Value}
		}

		if _, ok := tk.(token.Semicolon); ok {
			if root == nil {
				return nil, fmt.Errorf("unexpected semicolon token with no expression")
			}

			p.current-- // put the semicolon back, it might be used to end the parent
			return root, nil
		}

		// if _, ok := tk.(lexer.OpenParenToken); ok {
		// 	p.next() // consume the open parenthesis token

		// 	expr, err := p.parseExpression()
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	closeTk := p.next()
		// 	if _, ok := closeTk.(lexer.CloseParenToken); !ok {
		// 		return nil, fmt.Errorf("expected close parenthesis, got %v", closeTk)
		// 	}

		// 	return expr, nil
		// }

		if binTk, ok := tk.(token.Binary); ok {
			expr, err := p.parseBinaryExpression(root, binTk)
			if err != nil {
				return nil, err
			}

			root = expr
		}

		p.next() // consume the token
	}

	panic("rached EOF without completing the expression")
}

// An aotmic expression is an expression that can stand on its own and does not
// require any other expressions to be complete.
func (p *Parser) parseAtomicExpression() (Expression, error) {
	tk := p.peek()
	if tk == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if intTk, ok := tk.(token.Integer); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	//function call

	//negation of integer

	return nil, fmt.Errorf("unexpected token: %v", tk)
}

func (p *Parser) parseBinaryExpression(root Expression, binaryToken token.Binary) (Expression, error) {

	// No preceeding expression, this is the first one
	if root == nil {
		switch binaryToken.Operation {
		case "-":
			// Negation of an integer
			intTk := p.next()
			fmt.Printf("intTk: %T\n", intTk)
			if _, ok := intTk.(token.Integer); !ok {
				return nil, fmt.Errorf("expected integer token, got %v", intTk)
			}

			return IntegerLiteral{Value: -intTk.(token.Integer).Value}, nil
		case "+":
			intTk := p.next()
			if _, ok := intTk.(token.Integer); !ok {
				return nil, fmt.Errorf("expected integer token, got %v", intTk)
			}

			return IntegerLiteral{Value: intTk.(token.Integer).Value}, nil
		default:
			return nil, fmt.Errorf("unexpected operator: %v", binaryToken)
		}
	}

	left := root

	p.next() // consume the operator token
	right, err := p.parseAtomicExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to parse right expression: %w", err)
	}

	var priority int
	switch binaryToken.Operation {
	case "+", "-":
		priority = 0
	case "*", "/":
		priority = 1
	default:
		panic("invalid operator in binary expression")
	}

	// Do a right swap if the priority of the current operator is higher
	if leftBinExpr, ok := left.(BinaryExpression); ok && priority > leftBinExpr.Priority {
		right = BinaryExpression{
			Left:     leftBinExpr.Right,
			Right:    right,
			Op:       binaryToken.Operation,
			Priority: priority,
		}

		leftBinExpr.Right = right
		root = leftBinExpr
	} else {
		root = BinaryExpression{
			Left:     left,
			Right:    right,
			Op:       binaryToken.Operation,
			Priority: priority,
		}
	}

	return root, nil
}
