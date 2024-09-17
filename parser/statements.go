package parser

import (
	"fmt"
	"leoscript/token"
	"leoscript/types"
)

func (p *Parser) parseStatement() (Statement, error) {
	tk := p.peek()

	switch tk.(type) {
	case token.EOF:
		return nil, fmt.Errorf("unexpected EOF")
	case token.Semicolon:
		return nil, fmt.Errorf("unexpected semicolon")
	case token.VarDecl:
		return p.parseVarDecl()
	case token.Type:
		return p.parseVarDecl()
	case token.FnDef: // TODO: Not a regular statement, only valid at the file level. Or Is it?? Hello, VSauce, Michael here.
		return p.parseFnDef()
	case token.Identifier:
		if _, ok := p.peekNext().(token.OpenParen); ok {
			return p.parseFnCall()
		}
		return p.parseAssignment()
	// case token.OpenBrace:
	// 	return p.parseBlock()
	// case token.If:
	// 	return p.parseIf()
	case token.Return:
		return p.parseReturn()
	default:
		// Todo: do not fall back to expression parsing
		return p.parseExpr()
	}
}

func (p *Parser) parseReturn() (Statement, error) {
	if _, ok := p.peekNext().(token.Semicolon); ok {
		return Return{}, nil
	}

	p.next() // Consume the return token

	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse return expression: %w", err)
	}

	if _, ok := p.peekNext().(token.Semicolon); !ok {
		return nil, fmt.Errorf("expected semicolon after return expression")
	}

	return Return{
		Value: expr,
	}, nil
}

func (p *Parser) parseAssignment() (Statement, error) {
	identifier := p.peek().(token.Identifier)

	if err := p.expect(token.OperatorType); err != nil {
		return nil, fmt.Errorf("expected assignment operator after identifier: %w", err)
	}

	if op := p.peek().(token.Operator).Op; op != "=" {
		return nil, fmt.Errorf("expected assignment operator, got %v", op)
	}

	p.next() // Consume the assignment operator

	// Parse the expression on the right side of the assignment
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse right hand expression: %w", err)
	}

	if err := p.expect(token.SemicolonType); err != nil {
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
		if err := p.expect(token.TypeType); err != nil {
			return nil, fmt.Errorf("expected type in argument list: %w", err)
		}

		argType := p.peek().(token.Type).Kind

		if err := p.expect(token.IdentifierType); err != nil {
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

		if err := p.expect(token.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after argument in argument list: %w", err)
		}
	}

	return args, nil
}

func (p *Parser) parseFnDef() (Statement, error) {
	if err := p.expect(token.IdentifierType); err != nil {
		return nil, fmt.Errorf("expected identifier after fn: %w", err)
	}

	identifier := p.peek().(token.Identifier)

	if err := p.expect(token.OpenParenType); err != nil {
		return nil, fmt.Errorf("expected open parenthesis after identifier: %w", err)
	}

	args, err := p.parseFnParams()
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Todo: Move this into the parseFnParams function
	if err := p.expect(token.CloseParenType); err != nil {
		return nil, fmt.Errorf("expected close parenthesis after open parenthesis: %w", err)
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
		return nil, fmt.Errorf("expected open brace after arguments in function definition")
	}

	// Consume the opening brace
	p.next()

	// Parse the function body
	body, err := p.parseBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to parse function body: %w", err)
	}

	return FnDef{
		Name:       identifier.Value,
		ReturnType: returnType,
		Args:       args,
		Body:       body,
	}, nil
}

func (p *Parser) parseVarDecl() (Statement, error) {
	var varType types.Type

	switch tk := p.peek().(type) {
	case token.Type:
		varType = tk.Kind
	case token.VarDecl:
		varType = nil
	default:
		panic(fmt.Sprintf("expected type or vardecl token, got %T", tk))
	}

	if err := p.expect(token.IdentifierType); err != nil {
		return nil, fmt.Errorf("expected identifier after intdef: %w", err)
	}

	identifier := p.peek().(token.Identifier)

	if err := p.expect(token.OperatorType); err != nil {
		return nil, fmt.Errorf("expected assignment operator after identifier: %w", err)
	}

	if op := p.peek().(token.Operator).Op; op != "=" {
		return nil, fmt.Errorf("expected assignment operator, got %v", op)
	}

	p.next() // Consume the assignment operator

	// Parse the expression on the right side of the assignment
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse right hand expression: %w", err)
	}

	if varType != nil {
		// If a type is specified, verify that the expression matches the type
		if expr.ReturnType() != varType {
			return nil, fmt.Errorf("type mismatch: expected %v, got %v", varType, expr.ReturnType())
		}
	} else {
		// If no type is specified, use the type of the expression
		varType = expr.ReturnType()
	}

	if err := p.expect(token.SemicolonType); err != nil {
		return nil, fmt.Errorf("expected semicolon after identifier: %w", err)
	}

	return VarDecl{
		Name:  identifier.Value,
		Type:  varType,
		Value: expr,
	}, nil
}
