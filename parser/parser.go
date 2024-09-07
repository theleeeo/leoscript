package parser

import (
	"leoscript/lexer"
)

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, current: -1}
}

type Parser struct {
	tokens  []lexer.Token
	current int

	Program Program
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

	for tk := p.next(); tk != nil; tk = p.next() {
		expr, err := p.parseExpression()
		if err != nil {
			return Program{}, err
		}

		p.Program.PushExpression(expr)
	}

	return p.Program, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	tk := p.peek()

	if intTk, ok := tk.(lexer.IntegerToken); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	if binTk, ok := tk.(lexer.BinaryToken); ok {
		p.next() // consume the operator token

		left := p.Program.PopExpression()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		var priority int
		switch binTk.Operation {
		case "+", "-":
			priority = 0
		case "*", "/":
			priority = 1
		default:
			panic("invalid operator in binary expression")
		}

		if leftBinExpr, ok := left.(BinaryExpression); ok && priority > leftBinExpr.Priority {
			right = BinaryExpression{
				Left:     leftBinExpr.Right,
				Right:    right,
				Op:       binTk.Operation,
				Priority: priority,
			}

			leftBinExpr.Right = right
			return leftBinExpr, nil
		}

		return BinaryExpression{
			Left:     left,
			Right:    right,
			Op:       binTk.Operation,
			Priority: priority,
		}, nil
	}

	return nil, nil
}

func (p *Program) PushExpression(expr Expression) {
	p.Body = append(p.Body, expr)
}

func (p *Program) PopExpression() Expression {
	if len(p.Body) == 0 {
		panic("no expressions to pop")
	}

	expr := p.Body[len(p.Body)-1]
	p.Body = p.Body[:len(p.Body)-1]

	return expr
}
