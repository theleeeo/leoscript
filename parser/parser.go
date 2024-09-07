package parser

import (
	"fmt"
	"leoscript/lexer"
)

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, current: -1}
}

type Parser struct {
	tokens  []lexer.Token
	current int
}

func (p *Parser) next() lexer.Token {
	p.current++

	if p.current >= len(p.tokens) {
		// TODO: EOF token?
		return nil
	}

	return p.tokens[p.current]
}

func (p *Parser) peek() lexer.Token {
	if p.current >= len(p.tokens) {
		panic("tries to peek past EOF")
	}

	return p.tokens[p.current]
}

type Program struct {
	Body []Expression
}

func (p *Parser) Parse() (Program, error) {
	var program Program

	for tk := p.next(); tk != nil; tk = p.next() {
		fmt.Println(tk)
		expr, err := p.parseExpression()
		if err != nil {
			return Program{}, err
		}

		fmt.Println(expr)

		program.Body = append(program.Body, expr)
	}

	fmt.Println(program)

	return program, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	tk := p.peek()

	if intTk, ok := tk.(lexer.IntegerToken); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	return nil, nil
}
