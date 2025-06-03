package parser

import (
	"fmt"
	"leoscript/token"
	"leoscript/types"
)

func (p *Parser) ParseStatement() (Statement, error) {
	tk := p.peek()

	switch tk := tk.(type) {
	case token.VarDecl, token.Type:
		return p.parseVarDecl()
	case token.Identifier:
		if _, ok := p.peekNext().(token.OpenParen); ok {
			return p.ParseExpr()
		}
		return p.parseAssignment()
	case token.Return:
		return p.parseReturn()
	case token.If:
		return p.parseIf()
	case token.Else:
		return p.parseIf()
	case token.While:
		return p.parseWhile()
	default:
		return nil, fmt.Errorf("unexpected token type %T", tk)
	}
}

func (p *Parser) parseReturn() (Statement, error) {
	p.next() // Consume the return token

	// An empty return statement
	if _, ok := p.peek().(token.Semicolon); ok {
		return Return{}, nil
	}

	expr, err := p.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("parsing return expression: %w", err)
	}

	if err := p.expectCurrent(token.SemicolonType); err != nil {
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
		return nil, fmt.Errorf("parsing right hand expression: %w", err)
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
	if _, ok := p.next().(token.CloseParen); ok {
		return []Argument{}, nil
	}

	args := make([]Argument, 0)
	for {
		if err := p.expectCurrent(token.TypeType); err != nil {
			return nil, fmt.Errorf("expected type in argument list: %w", err)
		}

		argType := p.peek().(token.Type).Kind

		p.next() // Consume the type

		if err := p.expectCurrent(token.IdentifierType); err != nil {
			return nil, fmt.Errorf("expected identifier after type in argument list: %w", err)
		}

		identifier := p.peek().(token.Identifier)

		args = append(args, Argument{
			Name: identifier.Value,
			Type: argType,
		})

		p.next() // Consume the identifier

		// If we have hit the close parenthesis, we have parsed all arguments
		if _, ok := p.peek().(token.CloseParen); ok {
			break
		}

		// If there is another argument, there should be a comma
		if err := p.expectCurrent(token.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after argument in argument list: %w", err)
		}

		// Consume the comma
		p.next()
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
		return FnDef{}, fmt.Errorf("parsing arguments: %w", err)
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

	body, err := p.parseBlock()
	if err != nil {
		return FnDef{}, fmt.Errorf("parsing function body: %w", err)
	}

	return FnDef{
		Name:       identifier.Value,
		ReturnType: returnType,
		Args:       args,
		Body:       body,
	}, nil
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
		return VarDecl{}, fmt.Errorf("parsing right hand expression: %w", err)
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
		return If{}, fmt.Errorf("parsing if condition: %w", err)
	}

	if err := p.expectCurrent(token.OpenBraceType); err != nil {
		return If{}, fmt.Errorf("expected open brace after if condition: %w", err)
	}

	thenBlock, err := p.parseBlock()
	if err != nil {
		return If{}, fmt.Errorf("parsing if block: %w", err)
	}

	if err := p.expectCurrent(token.CloseBraceType); err != nil {
		return If{}, fmt.Errorf("expected close brace after if block: %w", err)
	}

	ifStmt := If{
		Cond: expr,
		Then: thenBlock,
	}

	if _, ok := p.peekNext().(token.Else); ok {
		p.next() // Consume the close brace

		elseBlock, err := p.parseElse()
		if err != nil {
			return If{}, fmt.Errorf("parsing else block: %w", err)
		}
		ifStmt.Else = elseBlock
	}

	return ifStmt, nil
}

func (p *Parser) parseElse() ([]Statement, error) {
	p.next() // Consume the else token

	// If the next token is an if, this is an else if statement
	if _, ok := p.peek().(token.If); ok {
		if err := p.expectCurrent(token.IfType); err != nil {
			return nil, fmt.Errorf("expected if after else: %w", err)
		}

		ifStmt, err := p.parseIf()
		if err != nil {
			return nil, fmt.Errorf("parsing else if statement: %w", err)
		}

		return []Statement{ifStmt}, nil
	}

	// If the next token is an open brace, this is a simple else block

	if err := p.expectCurrent(token.OpenBraceType); err != nil {
		return nil, fmt.Errorf("expected open brace after else: %w", err)
	}

	elseBlock, err := p.parseBlock()
	if err != nil {
		return nil, fmt.Errorf("parsing else block: %w", err)
	}

	if err := p.expectCurrent(token.CloseBraceType); err != nil {
		return nil, fmt.Errorf("expected close brace after else block: %w", err)
	}

	return elseBlock, nil
}

func (p *Parser) parseWhile() (While, error) {
	p.next() // Consume the while token

	expr, err := p.ParseExpr()
	if err != nil {
		return While{}, fmt.Errorf("parsing while condition: %w", err)
	}

	if err := p.expectCurrent(token.OpenBraceType); err != nil {
		return While{}, fmt.Errorf("expected open brace after while condition: %w", err)
	}

	body, err := p.parseBlock()
	if err != nil {
		return While{}, fmt.Errorf("parsing while body: %w", err)
	}

	if err := p.expectCurrent(token.CloseBraceType); err != nil {
		return While{}, fmt.Errorf("expected close brace after while body: %w", err)
	}

	return While{
		Cond: expr,
		Body: body,
	}, nil
}
