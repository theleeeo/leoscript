package parser

import (
	"fmt"
	"leoscript/token"
	"slices"
)

func NewParser(tokens []token.Token, scope *Scope) *Parser {
	if scope == nil {
		scope = NewScope(nil)
	}
	return &Parser{tokens: tokens, current: 0, scope: scope}
}

type Parser struct {
	tokens  []token.Token
	current int

	scope *Scope

	Program Program
}

func (p *Parser) Scope() *Scope {
	return p.scope
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

func (p *Parser) peekNext() token.Token {
	if p.current+1 >= len(p.tokens) {
		return token.EOF{}
	}

	return p.tokens[p.current+1]
}

// putBack will move the current token back one step
// this is useful when we want to "undo" a token consumption
func (p *Parser) putBack() {
	p.current--
}

type Program struct {
	Body []Statement
}

func (p *Parser) ParseFile() (Program, error) {
	globalScope := NewScope(p.scope)

	for tk := p.peek(); tk.Type() != token.EOFType; tk = p.next() {
		var stmt Statement
		switch tk.(type) {
		case token.VarDecl:
			fmt.Println("Parsing variable declaration")
			varDecl, err := p.parseVarDecl()
			if err != nil {
				return Program{}, err
			}
			stmt = varDecl

			err = globalScope.RegisterVar(varDecl)
			if err != nil {
				return Program{}, err
			}

		case token.FnDef:
			fmt.Println("Parsing function definition")
			fnDef, err := p.parseFnDef()
			if err != nil {
				return Program{}, err
			}
			stmt = fnDef

			err = globalScope.RegisterFn(fnDef)
			if err != nil {
				return Program{}, err
			}

		default:
			return Program{}, fmt.Errorf("unexpected token type %T", tk)
		}

		p.Program.Body = append(p.Program.Body, stmt)
	}

	// Parse all the function bodies in the file now that the global scope has been built.
	for i, fnDef := range p.Program.Body {
		if fnDef, ok := fnDef.(FnDef); ok {
			parsedFnDef, err := fnDef.parseBody(globalScope)
			if err != nil {
				return Program{}, fmt.Errorf("failed to parse function body for %s: %w", fnDef.Name, err)
			}
			p.Program.Body[i] = parsedFnDef
		}
	}

	if !slices.ContainsFunc(p.Program.Body, func(stmt Statement) bool {
		if fnDef, ok := stmt.(FnDef); ok {
			return fnDef.Name == "main"
		}
		return false
	}) {
		return Program{}, fmt.Errorf("no main function found in file")
	}

	return p.Program, nil
}

func (p *Parser) parseBlock() ([]Statement, error) {
	stmts := []Statement{}

	for tk := p.peek(); tk.Type() != token.CloseBraceType; tk = p.next() {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}

		if err := p.expect(token.SemicolonType); err != nil {
			return nil, err
		}

		stmts = append(stmts, stmt)
	}

	return stmts, nil
}

func (fn FnDef) parseBody(s *Scope) (FnDef, error) {
	fmt.Println("Scope: ", s.varDecls)
	p := Parser{
		tokens: fn.bodySrc,
		scope:  NewScope(s),
	}

	stmts, err := p.parseBlock()
	if err != nil {
		return FnDef{}, fmt.Errorf("failed to parse function body: %w", err)
	}

	fn.Body = stmts
	fn.bodySrc = nil

	return fn, nil
}
