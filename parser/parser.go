package parser

import (
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

func (p *Parser) next() token.Token {
	p.current++

	if p.current >= len(p.tokens) {
		return nil
	}

	return p.tokens[p.current]
}

func (p *Parser) peek() token.Token {
	if p.current >= len(p.tokens) {
		return nil
	}

	return p.tokens[p.current]
}

type Program struct {
	Body []Expression
}

func (p *Parser) Parse() (Program, error) {
	for tk := p.peek(); tk != nil; tk = p.peek() {
		if _, ok := tk.(token.Semicolon); ok {
			p.next()
			continue
		}

		expr, err := p.parseExpr()
		if err != nil {
			return Program{}, err
		}

		p.Program.Body = append(p.Program.Body, expr)
		p.next()
	}

	return p.Program, nil
}
