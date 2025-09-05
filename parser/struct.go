package parser

import (
	"fmt"
	"leoscript/lexer"
	"leoscript/types"
)

func (p *Parser) parseStructDef() (types.Struct, error) {
	if err := p.expectCurrent(lexer.StructDefType); err != nil {
		return types.Struct{}, err
	}

	if err := p.expectNext(lexer.IdentifierType); err != nil {
		return types.Struct{}, fmt.Errorf("expected struct name, got %v", p.peek().Type())
	}

	name := p.peek().(lexer.Identifier).Value

	if err := p.expectNext(lexer.OpenBraceType); err != nil {
		return types.Struct{}, err
	}
	p.next() // Consume the '{' token

	fields := []types.Field{}
	for p.peek().Type() != lexer.CloseBraceType {
		field, err := p.parseField()
		if err != nil {
			return types.Struct{}, fmt.Errorf("parsing field: %w", err)
		}
		fields = append(fields, field)

		p.next() // Move on from this field
	}

	if err := p.expectCurrent(lexer.CloseBraceType); err != nil {
		return types.Struct{}, err
	}

	return types.Struct{Name: name, Fields: fields}, nil
}

func (p *Parser) parseField() (types.Field, error) {
	if err := p.expectCurrent(lexer.IdentifierType); err != nil {
		return types.Field{}, err
	}

	fieldType := p.peek().(lexer.Identifier).Value

	if err := p.expectNext(lexer.IdentifierType); err != nil {
		return types.Field{}, fmt.Errorf("expected field name, got %v", p.peek())
	}

	name := p.peek().(lexer.Identifier).Value

	if err := p.expectNext(lexer.SemicolonType); err != nil {
		return types.Field{}, err
	}

	return types.Field{Name: name, Type: unresolvedTypeIdentifier{Name: fieldType}}, nil
}
