package parser

import (
	"fmt"
	"leoscript/token"
)

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

type Parser struct {
	tokens  []token.Token
	current int

	Program Program
}

// next will consume the current token and return the next one
func (p *Parser) next() token.Token {
	p.current++

	if p.current >= len(p.tokens) {
		return token.EOF{}
	}

	return p.tokens[p.current]
}

// expect will return an error if the next token is not of the expected type
func (p *Parser) expect(tk token.TokenType) error {
	if tk != p.next().Type() {
		return fmt.Errorf("expected token type %v, got %v", tk, p.peek().Type())
	}

	return nil
}

// peek will return the current token without consuming it
func (p *Parser) peek() token.Token {
	if p.current >= len(p.tokens) {
		return token.EOF{}
	}

	return p.tokens[p.current]
}

// putBack will move the current token back one step
// this is useful when we want to "undo" a token consumption
func (p *Parser) putBack() {
	p.current--
}

type Program struct {
	Body []Statement
}

func (p *Parser) Parse() (Program, error) {
	for tk := p.peek(); tk.Type() != token.EOFType; tk = p.next() {
		if _, ok := tk.(token.Semicolon); ok {
			continue
		}

		expr, err := p.parseStatement()
		if err != nil {
			return Program{}, err
		}

		p.Program.Body = append(p.Program.Body, expr)
	}

	return p.Program, nil
}
