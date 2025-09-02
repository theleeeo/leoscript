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
		switch p.peekNext().(type) {
		case token.OpenParen:
			return p.ParseExpr()
		case token.Identifier:
			return p.parseVarDecl()
		case token.Operator:
			return p.parseAssignment()
		default:
			return nil, fmt.Errorf("unexpected token after identifier: %T", p.peek())
		}
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
		return Return{
			Value: VoidLiteral{},
		}, nil
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

func (p *Parser) parseFnParams() ([]Parameter, error) {
	// Check if the function has no parameters
	if _, ok := p.next().(token.CloseParen); ok {
		return []Parameter{}, nil
	}

	args := make([]Parameter, 0)
	for {
		var argType types.Type
		switch tk := p.peek().(type) {
		case token.Type:
			argType = tk.Kind
		case token.Identifier:
			argType = unresolvedTypeIdentifier{Name: tk.Value}
		default:
			return nil, fmt.Errorf("expected type in parameter list, got %T", tk)
		}

		p.next() // Consume the type

		if err := p.expectCurrent(token.IdentifierType); err != nil {
			return nil, fmt.Errorf("expected identifier after type in parameter list: %w", err)
		}

		identifier := p.peek().(token.Identifier)

		args = append(args, Parameter{
			Name: identifier.Value,
			Type: argType,
		})

		p.next() // Consume the identifier

		// If we have hit the close parenthesis, we have parsed all parameters
		if _, ok := p.peek().(token.CloseParen); ok {
			break
		}

		// If there is another parameter, there should be a comma
		if err := p.expectCurrent(token.CommaType); err != nil {
			return nil, fmt.Errorf("expected comma after parameter in parameter list: %w", err)
		}

		// Consume the comma
		p.next()
	}

	return args, nil
}

func (p *Parser) parseFnDef() (FnDef, error) {
	var exported bool
	var stub bool

	switch tk := p.peek().(type) {
	case token.Exported:
		exported = true

		if err := p.expectNext(token.FnDefType); err != nil {
			return FnDef{}, fmt.Errorf("expected fn after exported: %w", err)
		}

	case token.StubDef:
		stub = true

		// Stubs do not have an "fn" token
	case token.FnDef:
		// Do nothing, we are already at the fn token
	default:
		panic(fmt.Sprintf("expected exported, stub or fn token, got %T", tk))
	}

	if err := p.expectNext(token.IdentifierType); err != nil {
		return FnDef{}, fmt.Errorf("expected identifier after fn: %w", err)
	}

	identifier := p.peek().(token.Identifier)

	if err := p.expectNext(token.OpenParenType); err != nil {
		return FnDef{}, fmt.Errorf("expected open parenthesis after identifier: %w", err)
	}

	args, err := p.parseFnParams()
	if err != nil {
		return FnDef{}, fmt.Errorf("parsing parameters: %w", err)
	}

	var returnType types.Type

	p.next() // Consume the close parenthesis

	// Check if the function has a return type
	switch tk := p.peek().(type) {
	case token.Type:
		returnType = tk.Kind
		p.next() // Consume the type token
	case token.Identifier:
		returnType = unresolvedTypeIdentifier{Name: tk.Value}
		p.next() // Consume the identifier token
	default:
		// No return type is specified
		returnType = types.Void
	}

	fnDef := FnDef{
		Name:       identifier.Value,
		ReturnType: returnType,
		Params:     args,
		Body:       nil,
		Stub:       stub,
		Exported:   exported,
	}

	// Stub functions do not have a body
	if stub {
		if err := p.expectCurrent(token.SemicolonType); err != nil {
			return FnDef{}, fmt.Errorf("expected semicolon after stub definition: %w", err)
		}

		return fnDef, nil
	}

	if err := p.expectCurrent(token.OpenBraceType); err != nil {
		return FnDef{}, fmt.Errorf("expected open brace after parameters in function definition: %w", err)
	}

	body, err := p.parseBlock()
	if err != nil {
		return FnDef{}, fmt.Errorf("parsing function body: %w", err)
	}

	// As syntactic sugar, we allow a void function to end without an explicit return statement.
	// The ast should however contain a void-return to keep the ast consistent.
	if returnType == types.Void {
		if len(body) == 0 { // TODO: Remove, no functions should be empty?
			body = append(body, Return{Value: VoidLiteral{}})
		}

		_, hasEndReturn := body[len(body)-1].(Return)
		if !hasEndReturn {
			body = append(body, Return{Value: VoidLiteral{}})
		}
	}
	fnDef.Body = body

	return fnDef, nil
}

func (p *Parser) parseVarDecl() (VarDecl, error) {
	var exported bool
	if _, ok := p.peek().(token.Exported); ok {
		exported = true
		p.next() // Consume the exported token
	}

	var varType types.Type

	switch tk := p.peek().(type) {
	case token.Type:
		varType = tk.Kind
	case token.VarDecl:
		varType = types.Unspecified
	case token.Identifier:
		varType = unresolvedTypeIdentifier{Name: tk.Value}
	default:
		panic(fmt.Sprintf("expected type or vardecl token, got %T", tk))
	}

	if err := p.expectNext(token.IdentifierType); err != nil {
		return VarDecl{}, fmt.Errorf("expected identifier after intdef: %w", err)
	}

	identifier := p.peek().(token.Identifier)

	p.next() // Consume the identifier

	var valExpr Expression
	if op, ok := p.peek().(token.Operator); ok {
		if op.Op != "=" {
			return VarDecl{}, fmt.Errorf("expected assignment operator, got %v", op.Op)
		}

		p.next() // Consume the assignment operator

		// Parse the expression on the right side of the assignment
		expr, err := p.ParseExpr()
		if err != nil {
			return VarDecl{}, fmt.Errorf("parsing right hand expression: %w", err)
		}
		valExpr = expr
	}

	if err := p.expectCurrent(token.SemicolonType); err != nil {
		return VarDecl{}, fmt.Errorf("expected semicolon after identifier: %w", err)
	}

	return VarDecl{
		Name:     identifier.Value,
		Type:     varType,
		Value:    valExpr,
		Exported: exported,
	}, nil
}

type unresolvedTypeIdentifier struct {
	Name string
}

func (u unresolvedTypeIdentifier) String() string {
	return fmt.Sprintf("{unresolved type: %s}", u.Name)
}

func (u unresolvedTypeIdentifier) Kind() types.Kind {
	return types.KindInvalid
}

func (u unresolvedTypeIdentifier) Size() uint64 {
	return 0
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
