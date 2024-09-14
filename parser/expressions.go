package parser

import (
	"fmt"
	"leoscript/token"
)

// Start parsing an expression group.
// An expression group is a group of expressions whose length is not known statically and can contain any number of sub expressions.
// An expression group is terminated by a semicolon token or a close parenthesis token.
func (p *Parser) parseExpr() (Expression, error) {
	var root Expression

	for tk := p.peek(); tk != nil; tk = p.peek() {
		// fmt.Printf("tk: %T %v\n", tk, tk)

		switch tk := tk.(type) {

		case token.Integer: // This will only hit if it is the first token in the expression, otherwise it will have been handled by another branch.

			if root != nil {
				return nil, fmt.Errorf("unexpected integer token preceeding another expression, v=%v", tk)
			}
			root = IntegerLiteral{Value: tk.Value}

		case token.Semicolon:
			if root == nil {
				return nil, fmt.Errorf("unexpected semicolon token with no expression")
			}

			// TODO: Really needed?
			p.current-- // put the semicolon back, it might be used to end the parent
			return root, nil

		case token.OpenParen:
			p.next() // consume the open parenthesis token
			expr, err := p.parseExpr()
			if err != nil {
				return nil, fmt.Errorf("failed to parse expression group: %w", err)
			}

			if binExpr, ok := expr.(BinaryExpression); ok {
				binExpr.Priority = 100
				expr = binExpr
			}

			nextTk := p.next()
			if _, ok := nextTk.(token.CloseParen); !ok {
				return nil, fmt.Errorf("expected close parenthesis token, got %v", nextTk)
			}

			root = expr

		case token.CloseParen:

			p.current-- // put the close-paren back, it will be verified by the parent
			return root, nil

		case token.Binary:
			expr, err := p.parseBinaryExpr(root)
			if err != nil {
				return nil, err
			}

			root = expr

		default:
			return nil, fmt.Errorf("unexpected token in expression: T=%T V=%v", tk, tk)
		}

		p.next() // consume the token
	}

	panic("reached EOF without completing the expression")
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
	tk := p.peek()
	if tk == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if intTk, ok := tk.(token.Integer); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	if _, ok := tk.(token.Binary); ok {
		return p.parseNumberOperation()
	}

	//function call

	return nil, fmt.Errorf("unexpected token: %v", tk)
}

func (p *Parser) parseNumberOperation() (Expression, error) {
	var binTk token.Binary
	tk := p.peek()
	if bt, ok := tk.(token.Binary); ok {
		binTk = bt
	} else {
		panic("unexpected token type")
	}

	switch binTk.Operation {
	case "-":
		// Negation of an integer
		intTk := p.next()
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
		return nil, fmt.Errorf("unexpected operator: %v", binTk)
	}
}

func (p *Parser) parseBinaryExpr(root Expression) (Expression, error) {
	var binTk token.Binary
	tk := p.peek()
	if bt, ok := tk.(token.Binary); ok {
		binTk = bt
	} else {
		panic("unexpected token type")
	}

	// No preceeding expression, this is the first one
	if root == nil {
		return p.parseNumberOperation()
	}

	var left Expression = root
	var right Expression

	var priority int
	switch binTk.Operation {
	case "+", "-":
		priority = 0
	case "*", "/":
		priority = 1
	default:
		panic("invalid operator in binary expression")
	}

	nextToken := p.next() // consume the operator token
	var err error
	if _, ok := nextToken.(token.OpenParen); ok {
		p.next() // consume the open parenthesis token
		right, err = p.parseExpr()
		if err != nil {
			return nil, fmt.Errorf("failed to parse expression group: %w", err)
		}

		if binExpr, ok := right.(BinaryExpression); ok {
			binExpr.Priority = 100
			right = binExpr
		}

		nextTk := p.next()
		if _, ok := nextTk.(token.CloseParen); !ok {
			return nil, fmt.Errorf("expected close parenthesis token, got %v", nextTk)
		}

	} else {
		right, err = p.parsePrimaryExpression()
		if err != nil {
			return nil, fmt.Errorf("failed to parse right expression: %w", err)
		}
	}

	// fmt.Printf("binary consumes tk: %T %v\n", right, right)

	// Do a right swap if the priority of the current operator is higher
	if leftBinExpr, ok := left.(BinaryExpression); ok && priority > leftBinExpr.Priority {
		right = BinaryExpression{
			Left:     leftBinExpr.Right,
			Right:    right,
			Op:       binTk.Operation,
			Priority: priority,
		}

		leftBinExpr.Right = right
		root = leftBinExpr
	} else {
		root = BinaryExpression{
			Left:     left,
			Right:    right,
			Op:       binTk.Operation,
			Priority: priority,
		}
	}

	return root, nil
}
