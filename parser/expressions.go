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

// parseFirstExpression will parse the first part of an expression which requires special handling in some cases.
func (p *Parser) parseFirstExpression() (Expression, error) {
	switch tk := p.peek().(type) {
	case token.EOF:
		return nil, fmt.Errorf("unexpected EOF")
	case token.Semicolon:
		return nil, fmt.Errorf("unexpected semicolon token with no expression")
	case token.CloseParen:
		return nil, fmt.Errorf("unexpected close-paren token")
	case token.Integer:
		return IntegerLiteral{Value: tk.Value}, nil
	case token.Boolean:
		return BooleanLiteral{Value: tk.Value}, nil
	case token.Operator:
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
	switch tk := p.peek().(type) {
	case token.EOF:
		return nil, fmt.Errorf("unexpected EOF")
	case token.Integer:
		return IntegerLiteral{Value: tk.Value}, nil
	case token.Boolean:
		return BooleanLiteral{Value: tk.Value}, nil
	case token.Operator:
		return p.parseUnaryExpr()
	}

	return nil, fmt.Errorf("unexpected token: %v", p.peek())
}

func (p *Parser) parseUnaryExpr() (Expression, error) {
	binTk := p.peek().(token.Operator)

	switch binTk.Op {
	case "-", "+":
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

	var right Expression

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
