package parser

import (
	"fmt"
	"leoscript/token"
	"slices"
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
	// fmt.Println("consumed: ", p.tokens[p.current-1].Type(), "new: ", p.peek().Type())

	if p.current >= len(p.tokens) {
		return token.EOF{}
	}

	return p.tokens[p.current]
}

// expectNext will consume the token and return an error if the next one is not of the expected type
func (p *Parser) expectNext(tk token.TokenType) error {
	if tk != p.next().Type() {
		return fmt.Errorf("expected token type %v, got %v", tk, p.peek().Type())
	}

	return nil
}

// expectConsume will return an error if the current token is not of the expected type
func (p *Parser) expectCurrent(tk token.TokenType) error {
	if tk != p.peek().Type() {
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

func (p *Parser) peekNext() token.Token {
	if p.current+1 >= len(p.tokens) {
		return token.EOF{}
	}

	return p.tokens[p.current+1]
}

// putBack will move the current token back one step
// this is useful when we want to "undo" a token consumption
// func (p *Parser) putBack() {
// 	// fmt.Println("putting back token", p.tokens[p.current].Type())
// 	p.current--
// }

type Program struct {
	// Global variable declarations
	VarDecls []VarDecl

	// Global function definitions
	FnDefs []FnDef
}

func (p *Parser) ParseFile() (Program, error) {
	for tk := p.peek(); tk.Type() != token.EOFType; tk = p.next() {
		switch tk.(type) {
		case token.VarDecl:
			varDecl, err := p.parseVarDecl()
			if err != nil {
				return Program{}, err
			}

			p.Program.VarDecls = append(p.Program.VarDecls, varDecl)

		case token.FnDef:
			fnDef, err := p.parseFnDef()
			if err != nil {
				return Program{}, err
			}

			p.Program.FnDefs = append(p.Program.FnDefs, fnDef)

		default:
			return Program{}, fmt.Errorf("unexpected token type %T", tk)
		}
	}

	if !slices.ContainsFunc(p.Program.FnDefs, func(fnDef FnDef) bool {
		return fnDef.Name == "main"
	}) {
		return Program{}, fmt.Errorf("no main function found in file")
	}

	return p.Program, nil
}

func (p *Parser) parseBlock() ([]Statement, error) {
	p.next() // Consume the open brace

	stmts := []Statement{}

	for tk := p.peek(); tk.Type() != token.CloseBraceType; tk = p.next() {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, fmt.Errorf("parsing statement: %w", err)
		}

		// If the statement is an If statement, we don't expect a semicolon after it
		switch stmt.(type) {
		case If:
			// Do nothing, no semicolon expected after an If statement
		case While:
			// Do nothing, no semicolon expected after a While statement
		default:
			if err := p.expectCurrent(token.SemicolonType); err != nil {
				return nil, err
			}
		}

		stmts = append(stmts, stmt)
	}

	return stmts, nil
}
