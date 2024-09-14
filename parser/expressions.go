package parser

import (
	"errors"
	"fmt"
	"leoscript/token"
)

func (p *Parser) parseExpr() (Expression, error) {
	var root Expression

	// Parse the first part in the expression
	// This is a special case because it can encounter cases that would never be encountered in the top of the main.
	// For example, a unary operator or a number. Thy would be handled by the parsing of a binary expression normally.
	firstExpr, err := p.parseFirstExpression()
	if err != nil {
		return nil, err
	}
	root = firstExpr

	for tk := p.next(); tk != nil; tk = p.next() {
		switch tk := tk.(type) {
		case token.Semicolon:
			p.current-- // put the semicolon back, it might be used to end the parent
			return root, nil

		case token.OpenParen:
			expr, err := p.handleSubgroup()
			if err != nil {
				return nil, err
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
	}

	return nil, errors.New("reached EOF without completing the expression")
}

func (p *Parser) handleSubgroup() (Expression, error) {
	p.next() // consume the open-paren token
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	// If the expression is a binary expression, set the priority to 100 to that it is not reordered
	if binExpr, ok := expr.(BinaryExpression); ok {
		binExpr.Priority = 100
		expr = binExpr
	}

	if err := p.expect(token.CloseParenType); err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	return expr, nil
}

func (p *Parser) parseFirstExpression() (Expression, error) {
	switch tk := p.peek().(type) {
	case token.Integer: // This will only hit if it is the first token in the expression, otherwise it will have been handled by another branch.
		return IntegerLiteral{Value: tk.Value}, nil
	case token.Semicolon:
		return nil, fmt.Errorf("unexpected semicolon token with no expression")
	case token.MathOp:
		expr, err := p.parseUnaryExpr()
		if err != nil {
			return nil, fmt.Errorf("failed to parse unary expression: %w", err)
		}
		return expr, nil
	case token.OpenParen:
		expr, err := p.handleSubgroup()
		if err != nil {
			return nil, err
		}

		return expr, nil
	}

	// No token requiring special handling when being first was found, return nil
	return nil, nil
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
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
		expr, err := p.parsePrimaryExpression()
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
		expr, err := p.handleSubgroup()
		if err != nil {
			return nil, err
		}

		right = expr

	} else {
		right, err = p.parsePrimaryExpression()
		if err != nil {
			return nil, fmt.Errorf("failed to parse right expression: %w", err)
		}
	}

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
