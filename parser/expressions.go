package parser

import (
	"errors"
	"fmt"
	"leoscript/token"
	"leoscript/types"
)

func (p *Parser) ParseExpr() (Expression, error) {
	// Parse the first part in the expression.
	// This will be the root of the expression tree.
	root, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for tk := p.next(); tk != nil; tk = p.next() {
		switch tk := tk.(type) {
		case token.Semicolon:
			return root, nil

		case token.OpenParen:
			expr, err := p.handleSubgroup()
			if err != nil {
				return nil, err
			}

			root = expr

		case token.CloseParen:
			return root, nil

		case token.Operator:
			expr, err := p.parseBinaryExpr(root)
			if err != nil {
				return nil, err
			}

			root = expr

		case token.OpenBrace:
			return root, nil

		case token.CloseBrace: // Found in struct literals
			return root, nil

		case token.Comma:
			// Commas are used in function calls.
			return root, nil

		default:
			return nil, fmt.Errorf("unexpected token in expression: T=%T V=%v", tk, tk)
		}
	}

	return nil, errors.New("reached EOF without completing the expression")
}

func (p *Parser) handleSubgroup() (Expression, error) {
	p.next() // consume the open-paren token
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("parsing expression: %w", err)
	}

	// If the expression is a binary expression, set the priority to the max so that it is never reordered
	if binExpr, ok := expr.(BinaryExpression); ok {
		binExpr.priority = token.PRIO_PAREN
		expr = binExpr
	}

	if err := p.expectCurrent(token.CloseParenType); err != nil {
		return nil, fmt.Errorf("parsing expression: %w", err)
	}

	return expr, nil
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
	switch tk := p.peek().(type) {
	case token.Integer:
		return IntegerLiteral{Value: tk.Value}, nil
	case token.Boolean:
		return BooleanLiteral{Value: tk.Value}, nil
	case token.StringLiteral:
		return StringLiteral{Value: tk.Value}, nil
	case token.Operator:
		return p.parseUnaryExpr()
	case token.OpenParen:
		return p.handleSubgroup()
	case token.Identifier:
		switch p.peekNext().(type) {
		case token.OpenParen:
			return p.parseFnCall()
		case token.OpenBrace:
			return p.parseStructLiteral()
		}

		return p.parseVarIdentifier()
	case token.OpenBrace:
		return p.parseStructLiteral()
	}

	return nil, fmt.Errorf("unexpected token in primary expression: T=%T V=%v", p.peek(), p.peek())
}

func (p *Parser) parseVarIdentifier() (Expression, error) {
	varIden := p.peek().(token.Identifier)

	return VarIdentifier{Name: varIden.Value}, nil
}

func (p *Parser) parseFnCall() (Expression, error) {
	identifier := p.peek().(token.Identifier)

	if err := p.expectNext(token.OpenParenType); err != nil {
		return nil, fmt.Errorf("expected open parenthesis after function call: %w", err)
	}

	args, err := p.parseArgs()
	if err != nil {
		return nil, fmt.Errorf("parsing arguments: %w", err)
	}

	if err := p.expectCurrent(token.CloseParenType); err != nil {
		return nil, fmt.Errorf("expected close parenthesis after function call: %w", err)
	}

	return Call{
		Name: identifier.Value,
		Args: args,
	}, nil
}

func (p *Parser) parseArgs() ([]Expression, error) {
	if err := p.expectCurrent(token.OpenParenType); err != nil {
		return nil, fmt.Errorf("expected open parenthesis in argument list: %w", err)
	}

	p.next() // Consume the open parenthesis

	// If the next token is a close parenthesis, we have no arguments
	if _, ok := p.peek().(token.CloseParen); ok {
		return []Expression{}, nil
	}

	args := make([]Expression, 0)
	for {
		expr, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}

		args = append(args, expr)

		if _, ok := p.peek().(token.CloseParen); ok {
			break
		}

		if err := p.expectCurrent(token.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after argument in argument list: %w", err)
		}

		p.next() // Consume the comma
	}

	return args, nil
}

func (p *Parser) parseUnaryExpr() (Expression, error) {
	binTk := p.peek().(token.Operator)

	p.next() // consume the operator token

	switch binTk.Op {
	case "-", "+", "!":
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, fmt.Errorf("parsing right hand expression: %w", err)
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
		return nil, fmt.Errorf("parsing right expression: %w", err)
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

func (p *Parser) parseStructLiteral() (Expression, error) {
	structLit := StructLiteral{}

	if ident, ok := p.peek().(token.Identifier); ok {
		structLit.Type = unresolvedTypeIdentifier{Name: ident.Value}
		p.next() // Consume the type identifier
	} else {
		structLit.Type = types.Unspecified
	}

	if err := p.expectCurrent(token.OpenBraceType); err != nil {
		return nil, fmt.Errorf("expected open brace for struct literal: %w", err)
	}

	p.next() // Consume the open brace

	for {
		field, err := p.parseFieldLiteral()
		if err != nil {
			return nil, fmt.Errorf("parsing struct field: %w", err)
		}
		structLit.Fields = append(structLit.Fields, field)

		if _, ok := p.peek().(token.CloseBrace); ok {
			break
		}

		if err := p.expectCurrent(token.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after struct field: %w", err)
		}

		p.next() // Consume the comma
	}

	if err := p.expectCurrent(token.CloseBraceType); err != nil {
		return nil, fmt.Errorf("expected close brace for struct literal: %w", err)
	}

	return structLit, nil
}

func (p *Parser) parseFieldLiteral() (FieldLiteral, error) {
	fieldLit := FieldLiteral{}

	if err := p.expectCurrent(token.IdentifierType); err != nil {
		return FieldLiteral{}, fmt.Errorf("expected field name: %w", err)
	}

	fieldLit.Name = p.peek().(token.Identifier).Value

	if err := p.expectNext(token.ColonType); err != nil {
		return FieldLiteral{}, fmt.Errorf("expected colon after field name: %w", err)
	}

	p.next() // Consume the colon

	value, err := p.ParseExpr()
	if err != nil {
		return FieldLiteral{}, fmt.Errorf("parsing field value: %w", err)
	}
	fieldLit.Value = value

	return fieldLit, nil
}
