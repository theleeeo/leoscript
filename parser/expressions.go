package parser

import (
	"errors"
	"fmt"
	"leoscript/token"
)

func (p *Parser) parseExpr() (Expression, error) {
	// Parse the first part in the expression.
	// This will be the root of the expression tree.
	root, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for tk := p.next(); tk != nil; tk = p.next() {
		switch tk := tk.(type) {
		case token.Semicolon:
			p.putBack() // put the semicolon back, it might be used to end the parent
			return root, nil

		case token.OpenParen:
			expr, err := p.handleSubgroup()
			if err != nil {
				return nil, err
			}

			root = expr

		case token.CloseParen:
			p.putBack() // put the close-paren back, it will be verified by the parent
			return root, nil

		case token.Operator:
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

	// If the expression is a binary expression, set the priority to the max so that it is never reordered
	if binExpr, ok := expr.(BinaryExpression); ok {
		binExpr.priority = token.PRIO_PAREN
		expr = binExpr
	}

	if err := p.expect(token.CloseParenType); err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	return expr, nil
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
	switch tk := p.peek().(type) {
	case token.Integer:
		return IntegerLiteral{Value: tk.Value}, nil
	case token.Boolean:
		return BooleanLiteral{Value: tk.Value}, nil
	case token.Operator:
		return p.parseUnaryExpr()
	case token.OpenParen:
		return p.handleSubgroup()

	}

	return nil, fmt.Errorf("unexpected token: %v", p.peek())
}

func (p *Parser) parseUnaryExpr() (Expression, error) {
	binTk := p.peek().(token.Operator)

	switch binTk.Op {
	case "-", "+", "!":
		p.next() // consume the operator token
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, fmt.Errorf("failed to parse right hand expression: %w", err)
		}

		return UnaryExpression{
			Expression: expr,
			Op:         binTk.Op,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected operator in unary expression: %v", binTk)
	}
}

func (p *Parser) parseBinaryExpr(root Expression) (Expression, error) {
	binTk := p.peek().(token.Operator)

	p.next() // consume the operator token
	right, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to parse right expression: %w", err)
	}

	// Do a right swap if the priority of the current operator is higher
	if rootBinExpr, ok := root.(BinaryExpression); ok {
		return rootBinExpr.PriorityMerge(binTk, right), nil
	}

	// Left side was not a binary expression.
	return BinaryExpression{
		Left:     root,
		Right:    right,
		Op:       binTk.Op,
		priority: binTk.Priority(),
	}, nil
}
