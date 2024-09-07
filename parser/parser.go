package parser

import (
	"fmt"
	"leoscript/token"
)

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: -1}
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
	for tk := p.next(); tk != nil; tk = p.next() {
		if _, ok := tk.(token.Semicolon); ok {
			continue
		}

		expr, err := p.parseExpression()
		if err != nil {
			return Program{}, err
		}

		p.Program.PushExpression(expr)
	}

	return p.Program, nil
}

// Start parsing an expression, beginning with the current.
func (p *Parser) parseExpression() (Expression, error) {
	var root Expression

	for tk := p.peek(); tk != nil; tk = p.peek() {
		// fmt.Printf("tk: %+T v=%v\n", tk, tk)
		// This will only hit if it is the first token in the expression.
		// Otherwise it will have been handled by another branch.
		if intTk, ok := tk.(token.Integer); ok {
			if root != nil {
				return nil, fmt.Errorf("unexpected integer token: %v", intTk)
			}
			root = IntegerLiteral{Value: intTk.Value}
		}

		if _, ok := tk.(token.Semicolon); ok {
			if root == nil {
				return nil, fmt.Errorf("unexpected semicolon token with no expression")
			}

			p.current-- // put the semicolon back, it might be used to end the parent
			return root, nil
		}

		// if _, ok := tk.(lexer.OpenParenToken); ok {
		// 	p.next() // consume the open parenthesis token

		// 	expr, err := p.parseExpression()
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	closeTk := p.next()
		// 	if _, ok := closeTk.(lexer.CloseParenToken); !ok {
		// 		return nil, fmt.Errorf("expected close parenthesis, got %v", closeTk)
		// 	}

		// 	return expr, nil
		// }

		if binTk, ok := tk.(token.Binary); ok {
			p.next() // consume the operator token

			if root == nil {
				panic("unhandled binary operator without left expression")
			}

			left := root

			right, err := p.parseAtomicExpression()
			if err != nil {
				return nil, fmt.Errorf("failed to parse right expression: %w", err)
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

			// Do a right swap if the priority of the current operator is higher
			if leftBinExpr, ok := left.(BinaryExpression); ok && priority > leftBinExpr.Priority {
				right = BinaryExpression{
					Left:     leftBinExpr.Right,
					Right:    right,
					Op:       binTk.Operation,
					Priority: priority,
				}

				leftBinExpr.Right = right
				root = leftBinExpr
			} else {
				root = BinaryExpression{
					Left:     left,
					Right:    right,
					Op:       binTk.Operation,
					Priority: priority,
				}
			}
		}

		p.next() // consume the token
	}

	panic("rached EOF without completing the expression")
}

// An aotmic expression is an expression that can stand on its own and does not
// require any other expressions to be complete.
func (p *Parser) parseAtomicExpression() (Expression, error) {
	tk := p.peek()
	if tk == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if intTk, ok := tk.(token.Integer); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	//function call

	//negation of integer

	return nil, fmt.Errorf("unexpected token: %v", tk)
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
