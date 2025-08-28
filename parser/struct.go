package parser

import (
	"fmt"
	"leoscript/token"
	"leoscript/types"
)

func (p *Parser) parseStructDef() (types.Struct, error) {
	if err := p.expectCurrent(token.StructDefType); err != nil {
		return types.Struct{}, err
	}

	if err := p.expectNext(token.IdentifierType); err != nil {
		return types.Struct{}, fmt.Errorf("expected struct name, got %v", p.peek().Type())
	}

	name := p.peek().(token.Identifier).Value

	if err := p.expectNext(token.OpenBraceType); err != nil {
		return types.Struct{}, err
	}
	p.next() // Consume the '{' token

	fields := []types.Field{}
	for p.peek().Type() != token.CloseBraceType {
		field, err := p.parseField()
		if err != nil {
			return types.Struct{}, fmt.Errorf("parsing field: %w", err)
		}
		fields = append(fields, field)

		p.next() // Move on from this field
	}

	if err := p.expectCurrent(token.CloseBraceType); err != nil {
		return types.Struct{}, err
	}

	return types.Struct{Name: name, Fields: fields}, nil
}

func (p *Parser) parseField() (types.Field, error) {
	if err := p.expectCurrent(token.TypeType); err != nil {
		return types.Field{}, err
	}

	fieldType := p.peek().(token.Type).Kind

	if err := p.expectNext(token.IdentifierType); err != nil {
		return types.Field{}, fmt.Errorf("expected field name, got %v", p.peek())
	}

	name := p.peek().(token.Identifier).Value

	if err := p.expectNext(token.SemicolonType); err != nil {
		return types.Field{}, err
	}

	return types.Field{Name: name, Type: fieldType}, nil
}
