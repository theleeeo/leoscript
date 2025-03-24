package parser

import (
	"fmt"
	"leoscript/token"
	"leoscript/types"
)

func (p *Parser) ParseStatement() (Statement, error) {
	tk := p.peek()

	switch tk := tk.(type) {
	case token.EOF:
		return nil, fmt.Errorf("unexpected EOF")
	case token.Semicolon:
		return nil, fmt.Errorf("unexpected semicolon")
	case token.VarDecl, token.Type:
		varDecl, err := p.parseVarDecl()
		if err == nil {
			p.scope.RegisterVar(varDecl)
		}
		return varDecl, err
	case token.Identifier:
		if _, ok := p.peekNext().(token.OpenParen); ok {
			fc, err := p.parseFnCall()
			if err != nil {
				return nil, err
			}

			// A function call is ended by its closing parenthesis.
			// When used in an expression, that is that.
			// When used in a statement however, it is only a valid statement termination if it is a semicolon afterwards.
			// TODO: This is better handled as it being considered an expression, always. p.parseExpr should be used instead of p.parseFnCall
			if err := p.expectNext(token.SemicolonType); err != nil {
				return nil, fmt.Errorf("expected semicolon after function call statement")
			}

			return fc, nil
		}
		return p.parseAssignment()
	case token.Return:
		return p.parseReturn()
	case token.If:
		return p.parseIf()
	default:
		return nil, fmt.Errorf("unexpected token type %T", tk)
	}
}

func (p *Parser) parseReturn() (Statement, error) {
	p.next() // Consume the return token

	if _, ok := p.peek().(token.Semicolon); ok {
		return Return{}, nil
	}

	expr, err := p.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse return expression: %w", err)
	}

	if _, ok := p.peek().(token.Semicolon); !ok {
		return nil, fmt.Errorf("expected semicolon after return expression")
	}

	return Return{
		Value: expr,
	}, nil
}

func (p *Parser) parseAssignment() (Statement, error) {
	identifier := p.peek().(token.Identifier)

	if err := p.expectNext(token.OperatorType); err != nil {
		return nil, fmt.Errorf("expected assignment operator after identifier: %w", err)
	}

	if op := p.peek().(token.Operator).Op; op != "=" {
		return nil, fmt.Errorf("expected assignment operator, got %v", op)
	}

	p.next() // Consume the assignment operator

	// Parse the expression on the right side of the assignment
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse right hand expression: %w", err)
	}

	if err := p.expectCurrent(token.SemicolonType); err != nil {
		return nil, fmt.Errorf("expected semicolon after identifier: %w", err)
	}

	return Assignment{
		Name:  identifier.Value,
		Value: expr,
	}, nil
}

func (p *Parser) parseFnParams() ([]Argument, error) {
	// Check if the function has no arguments
	if _, ok := p.peekNext().(token.CloseParen); ok {
		return []Argument{}, nil
	}

	args := make([]Argument, 0)
	for {
		if err := p.expectNext(token.TypeType); err != nil {
			return nil, fmt.Errorf("expected type in argument list: %w", err)
		}

		argType := p.peek().(token.Type).Kind

		if err := p.expectNext(token.IdentifierType); err != nil {
			return nil, fmt.Errorf("expected identifier after type in argument list: %w", err)
		}

		identifier := p.peek().(token.Identifier)

		args = append(args, Argument{
			Name: identifier.Value,
			Type: argType,
		})

		if _, ok := p.peekNext().(token.CloseParen); ok {
			break
		}

		if err := p.expectNext(token.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after argument in argument list: %w", err)
		}
	}

	return args, nil
}

func (p *Parser) parseFnDef() (FnDef, error) {
	if err := p.expectNext(token.IdentifierType); err != nil {
		return FnDef{}, fmt.Errorf("expected identifier after fn: %w", err)
	}

	identifier := p.peek().(token.Identifier)

	if err := p.expectNext(token.OpenParenType); err != nil {
		return FnDef{}, fmt.Errorf("expected open parenthesis after identifier: %w", err)
	}

	args, err := p.parseFnParams()
	if err != nil {
		return FnDef{}, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Todo: Move this into the parseFnParams function
	if err := p.expectNext(token.CloseParenType); err != nil {
		return FnDef{}, fmt.Errorf("expected close parenthesis after open parenthesis: %w", err)
	}

	var returnType types.Type

	// Check if the function has a return type
	if tk, ok := p.next().(token.Type); ok {
		returnType = tk.Kind
		p.next() // Consume the type token
	} else {
		// No return type is specified
		returnType = types.Void
	}

	if _, ok := p.peek().(token.OpenBrace); !ok {
		return FnDef{}, fmt.Errorf("expected open brace after arguments in function definition")
	}

	// Consume the opening brace
	p.next()

	// Get the raw source code of the function body.
	// This is to parse it later when the full global scope is available.
	bodySrc, err := p.getFnBodySource()
	if err != nil {
		return FnDef{}, fmt.Errorf("failed to get function body source: %w", err)
	}

	return FnDef{
		Name:       identifier.Value,
		ReturnType: returnType,
		Args:       args,
		bodySrc:    bodySrc,
	}, nil
}

func (p *Parser) getFnBodySource() ([]token.Token, error) {
	bodySource := make([]token.Token, 0)
	scopeDepth := 1

	for tk := p.peek(); scopeDepth > 0; tk = p.next() {
		switch tk.(type) {
		case token.OpenBrace:
			scopeDepth++
		case token.CloseBrace:
			scopeDepth--
		case token.EOF:
			return nil, fmt.Errorf("unexpected EOF")
		}

		bodySource = append(bodySource, tk)

		if scopeDepth == 0 {
			break
		}
	}

	return bodySource, nil
}

func (p *Parser) parseVarDecl() (VarDecl, error) {
	var varType types.Type

	switch tk := p.peek().(type) {
	case token.Type:
		varType = tk.Kind
	case token.VarDecl:
		varType = nil
	default:
		panic(fmt.Sprintf("expected type or vardecl token, got %T", tk))
	}

	if err := p.expectNext(token.IdentifierType); err != nil {
		return VarDecl{}, fmt.Errorf("expected identifier after intdef: %w", err)
	}

	identifier := p.peek().(token.Identifier)

	if err := p.expectNext(token.OperatorType); err != nil {
		return VarDecl{}, fmt.Errorf("expected assignment operator after identifier: %w", err)
	}

	if op := p.peek().(token.Operator).Op; op != "=" {
		return VarDecl{}, fmt.Errorf("expected assignment operator, got %v", op)
	}

	p.next() // Consume the assignment operator

	// Parse the expression on the right side of the assignment
	expr, err := p.ParseExpr()
	if err != nil {
		return VarDecl{}, fmt.Errorf("failed to parse right hand expression: %w", err)
	}

	if varType != nil {
		// If a type is specified, verify that the expression matches the type
		if expr.ReturnType() != varType {
			return VarDecl{}, fmt.Errorf("type mismatch: expected %v, got %v", varType, expr.ReturnType())
		}
	} else {
		// If no type is specified, use the type of the expression
		varType = expr.ReturnType()
	}

	if err := p.expectCurrent(token.SemicolonType); err != nil {
		return VarDecl{}, fmt.Errorf("expected semicolon after identifier: %w", err)
	}

	return VarDecl{
		Name:  identifier.Value,
		Type:  varType,
		Value: expr,
	}, nil
}

func (p *Parser) parseIf() (If, error) {
	p.next() // Consume the if token

	expr, err := p.ParseExpr()
	if err != nil {
		return If{}, fmt.Errorf("failed to parse if condition: %w", err)
	}

	if err := p.expectCurrent(token.OpenBraceType); err != nil {
		return If{}, fmt.Errorf("expected open brace after if condition: %w", err)
	}

	p.next() // Consume the open brace

	stmts, err := p.parseBlock()
	if err != nil {
		return If{}, fmt.Errorf("failed to parse if block: %w", err)
	}

	if err := p.expectCurrent(token.CloseBraceType); err != nil {
		return If{}, fmt.Errorf("expected close brace after if block: %w", err)
	}

	return If{
		Cond: expr,
		Body: stmts,
	}, nil
}
