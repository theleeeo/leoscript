package parser

import (
	"fmt"
	"leoscript/token"
)

func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
		program: &Program{},
	}
}

type Parser struct {
	tokens  []token.Token
	current int

	program *Program
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

type Program struct {
	// Global variable declarations
	VarDecls []VarDecl

	// Global function definitions
	FnDefs []FnDef

	// StubDefs for functions that are not defined in this file
	StubDefs []FnDef
}

func (p *Parser) Parse() (*Program, error) {
	for tk := p.peek(); tk.Type() != token.EOFType; tk = p.next() {
		switch tk.(type) {
		case token.VarDecl:
			varDecl, err := p.parseVarDecl()
			if err != nil {
				return nil, fmt.Errorf("parsing variable declaration: %w", err)
			}

			p.program.VarDecls = append(p.program.VarDecls, varDecl)

		case token.FnDef:
			fnDef, err := p.parseFnDef()
			if err != nil {
				return nil, fmt.Errorf("parsing function definition: %w", err)
			}

			p.program.FnDefs = append(p.program.FnDefs, fnDef)

		case token.StubDef:
			stub, err := p.parseStubdef()
			if err != nil {
				return nil, fmt.Errorf("parsing stub definition: %w", err)
			}

			p.program.StubDefs = append(p.program.StubDefs, stub)

		default:
			return nil, fmt.Errorf("unexpected token type %T", tk)
		}
	}

	reorderedVars, err := buildCallOrder(p.program.VarDecls)
	if err != nil {
		return nil, fmt.Errorf("building call order: %w", err)
	}
	p.program.VarDecls = reorderedVars

	if err := typeResolvingPass(p.program); err != nil {
		return nil, fmt.Errorf("type resolving pass failed: %w", err)
	}

	if err := typeValidationPass(p.program); err != nil {
		return nil, fmt.Errorf("type validation pass failed: %w", err)
	}

	return p.program, nil
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
		case If, While:
			// Do nothing, no semicolon expected
		default:
			if err := p.expectCurrent(token.SemicolonType); err != nil {
				return nil, err
			}
		}

		stmts = append(stmts, stmt)
	}

	return stmts, nil
}
