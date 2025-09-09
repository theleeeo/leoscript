package parser

import (
	"errors"
	"fmt"
	"leoscript/lexer"
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
		case lexer.Semicolon:
			return root, nil

		case lexer.OpenParen:
			expr, err := p.handleSubgroup()
			if err != nil {
				return nil, err
			}

			root = expr

		case lexer.CloseParen:
			return root, nil

		case lexer.Operator:
			expr, err := p.parseBinaryExpr(root)
			if err != nil {
				return nil, err
			}

			root = expr

		case lexer.OpenBrace:
			return root, nil

		case lexer.CloseBrace: // Found in struct literals
			return root, nil

		case lexer.CloseBracket: // Found in array literals
			return root, nil

		case lexer.Comma:
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
		binExpr.priority = lexer.PRIO_PAREN
		expr = binExpr
	}

	if err := p.expectCurrent(lexer.CloseParenType); err != nil {
		return nil, fmt.Errorf("parsing expression: %w", err)
	}

	return expr, nil
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
	switch tk := p.peek().(type) {
	case lexer.Integer:
		return IntegerLiteral{Value: tk.Value}, nil
	case lexer.Boolean:
		return BooleanLiteral{Value: tk.Value}, nil
	case lexer.StringLiteral:
		return StringLiteral{Value: tk.Value}, nil
	case lexer.Operator:
		return p.parseUnaryExpr()
	case lexer.OpenParen:
		return p.handleSubgroup()
	case lexer.Identifier:
		switch p.peekNext().(type) {
		case lexer.OpenParen:
			return p.parseFnCall()
		case lexer.OpenBrace:
			return p.parseStructLiteral()
		case lexer.OpenBracket:
			return p.parseArrayIndex()
		}
		return p.parseVarIdentifier()
	case lexer.OpenBrace:
		return p.parseStructLiteral()
	case lexer.OpenBracket:
		return p.parseArrayLiteral()
	}

	return nil, fmt.Errorf("unexpected token in primary expression: T=%T V=%v", p.peek(), p.peek())
}

func (p *Parser) parseVarIdentifier() (VarIdentifier, error) {
	varIden := p.peek().(lexer.Identifier)

	return VarIdentifier{Name: varIden.Value}, nil
}

func (p *Parser) parseFnCall() (Expression, error) {
	identifier := p.peek().(lexer.Identifier)

	if err := p.expectNext(lexer.OpenParenType); err != nil {
		return nil, fmt.Errorf("expected open parenthesis after function call: %w", err)
	}

	args, err := p.parseArgs()
	if err != nil {
		return nil, fmt.Errorf("parsing arguments: %w", err)
	}

	if err := p.expectCurrent(lexer.CloseParenType); err != nil {
		return nil, fmt.Errorf("expected close parenthesis after function call: %w", err)
	}

	return Call{
		Name: identifier.Value,
		Args: args,
	}, nil
}

func (p *Parser) parseArgs() ([]Expression, error) {
	if err := p.expectCurrent(lexer.OpenParenType); err != nil {
		return nil, fmt.Errorf("expected open parenthesis in argument list: %w", err)
	}

	p.next() // Consume the open parenthesis

	// If the next token is a close parenthesis, we have no arguments
	if _, ok := p.peek().(lexer.CloseParen); ok {
		return []Expression{}, nil
	}

	args := make([]Expression, 0)
	for {
		expr, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}

		args = append(args, expr)

		if _, ok := p.peek().(lexer.CloseParen); ok {
			break
		}

		if err := p.expectCurrent(lexer.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after argument in argument list: %w", err)
		}

		p.next() // Consume the comma
	}

	return args, nil
}

func (p *Parser) parseUnaryExpr() (Expression, error) {
	binTk := p.peek().(lexer.Operator)

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
	binTk := p.peek().(lexer.Operator)

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

	if ident, ok := p.peek().(lexer.Identifier); ok {
		structLit.Type = unresolvedTypeIdentifier{Name: ident.Value}
		p.next() // Consume the type identifier
	} else {
		structLit.Type = types.Unspecified
	}

	if err := p.expectCurrent(lexer.OpenBraceType); err != nil {
		return nil, fmt.Errorf("expected open brace for struct literal: %w", err)
	}

	p.next() // Consume the open brace

	for {
		field, err := p.parseFieldLiteral()
		if err != nil {
			return nil, fmt.Errorf("parsing struct field: %w", err)
		}
		structLit.Fields = append(structLit.Fields, field)

		if _, ok := p.peek().(lexer.CloseBrace); ok {
			break
		}

		if err := p.expectCurrent(lexer.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after struct field: %w", err)
		}

		p.next() // Consume the comma
	}

	return structLit, nil
}

func (p *Parser) parseFieldLiteral() (FieldLiteral, error) {
	fieldLit := FieldLiteral{}

	if err := p.expectCurrent(lexer.IdentifierType); err != nil {
		return FieldLiteral{}, fmt.Errorf("expected field name: %w", err)
	}

	fieldLit.Name = p.peek().(lexer.Identifier).Value

	if err := p.expectNext(lexer.ColonType); err != nil {
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

func (p *Parser) parseArrayLiteral() (Expression, error) {
	if err := p.expectCurrent(lexer.OpenBracketType); err != nil {
		return nil, fmt.Errorf("expected open bracket for array literal: %w", err)
	}

	p.next() // Consume the open bracket

	arrayLit := ArrayLiteral{
		ElementType: types.Unspecified,
	}
	for {
		element, err := p.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("parsing array element: %w", err)
		}
		arrayLit.Elements = append(arrayLit.Elements, element)

		if _, ok := p.peek().(lexer.CloseBracket); ok {
			break
		}

		if err := p.expectCurrent(lexer.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after array element: %w", err)
		}

		p.next() // Consume the comma
	}

	return arrayLit, nil
}

func (p *Parser) parseArrayIndex() (Expression, error) {
	ident, ok := p.peek().(lexer.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier for array index")
	}

	p.next() // Consume the identifier

	if err := p.expectCurrent(lexer.OpenBracketType); err != nil {
		return nil, fmt.Errorf("expected open bracket for array index: %w", err)
	}

	p.next() // Consume the open bracket

	index, err := p.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("parsing array index: %w", err)
	}

	if err := p.expectCurrent(lexer.CloseBracketType); err != nil {
		return nil, fmt.Errorf("expected close bracket for array index: %w", err)
	}

	return ArrayIndex{
		ElementType: types.Unspecified,
		ArrayVar:    ident.Value,
		Index:       index,
	}, nil
}
