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
	case token.VarDef:
		return p.parseVarDef(nil)
	case token.IntDef:
		return p.parseVarDef(types.Int)
	case token.Identifier:
		// return p.parseAssignment()
		panic("assignment not implemented")
	default:
		// Todo: do not fall back to expression parsing
		return p.parseExpr()
	}
}

func (p *Parser) parseVarDef(varType types.Type) (Statement, error) {
	if varType == nil {
		panic("untyped vardef not implemented")
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

	if err := p.expect(token.SemicolonType); err != nil {
		return nil, fmt.Errorf("expected semicolon after identifier: %w", err)
	}

	return VarDef{
		Name:  identifier.Value,
		Type:  varType,
		Value: expr,
	}, nil
}
