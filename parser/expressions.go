package parser

import (
	"fmt"
	"leoscript/token"
)

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

		case token.MathOp:
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

func (p *Parser) parseRightHandExpression() (Expression, error) {
	tk := p.peek()
	if tk == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if intTk, ok := tk.(token.Integer); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	if _, ok := tk.(token.MathOp); ok {
		return p.parseUnaryExpr()
	}

	//function call

	return nil, fmt.Errorf("unexpected token: %v", tk)
}

func (p *Parser) parseUnaryExpr() (Expression, error) {
	binTk := p.peek().(token.MathOp)

	switch binTk.Operation {
	case "-", "+":
		p.next() // consume the operator token
		expr, err := p.parseRightHandExpression()
		if err != nil {
			return nil, fmt.Errorf("failed to parse right hand expression: %w", err)
		}

		return UnaryExpression{
			Expression: expr,
			Op:         binTk.Operation,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected operator: %v", binTk)
	}
}

func (p *Parser) parseBinaryExpr(root Expression) (Expression, error) {
	binTk := p.peek().(token.MathOp)

	// No preceeding expression, this is the first one
	if root == nil {
		return p.parseUnaryExpr()
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
		right, err = p.parseRightHandExpression()
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
