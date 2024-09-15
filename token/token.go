package token

import "leoscript/types"

type Token interface {
	Type() TokenType
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=TokenType

type TokenType int

const (
	EOFType TokenType = iota

	// Literals
	IntegerType
	BooleanType

	// Parentheses
	OpenParenType
	CloseParenType
	OpenBraceType
	CloseBraceType

	VarDeclType
	TypeType // lol
	SemicolonType
	IdentifierType
	OperatorType
	FnDefType
	ReturnType
	CommaType
)

type EOF struct{}

func (EOF) Type() TokenType { return EOFType }

type Integer struct {
	Value int
}

func (Integer) Type() TokenType { return IntegerType }

type Operator struct {
	Op string
}

func (Operator) Type() TokenType { return OperatorType }

func (t Operator) Priority() Priority {
	switch t.Op {
	case "=":
		return PRIO_ASSIGN
	case "==", "!=":
		return PRIO_EQUALS
	case "<", ">", "<=", ">=":
		return PRIO_COMPARISON
	case "&&":
		return PRIO_AND
	case "||":
		return PRIO_OR
	case "+", "-":
		return PRIO_SUM
	case "*", "/":
		return PRIO_PRODUCT
	}

	panic("invalid operator in binary expression")
}

type OpenParen struct{}

func (OpenParen) String() string {
	return "{(}"
}

func (OpenParen) Type() TokenType { return OpenParenType }

type CloseParen struct{}

func (CloseParen) String() string {
	return "{)}"
}

func (CloseParen) Type() TokenType { return CloseParenType }

type Semicolon struct{}

func (Semicolon) String() string {
	return "{;}"
}

func (Semicolon) Type() TokenType { return SemicolonType }

type Identifier struct {
	Value string
}

func (Identifier) Type() TokenType { return IdentifierType }

type Boolean struct {
	Value bool
}

func (Boolean) Type() TokenType { return BooleanType }

type VarDecl struct{}

func (VarDecl) Type() TokenType { return VarDeclType }

type FnDef struct{}

func (FnDef) Type() TokenType { return FnDefType }

type OpenBrace struct{}

func (OpenBrace) Type() TokenType { return OpenBraceType }

type CloseBrace struct{}

func (CloseBrace) Type() TokenType { return CloseBraceType }

type Return struct{}

func (Return) Type() TokenType { return ReturnType }

type Type struct {
	Kind types.Type
}

func (Type) Type() TokenType { return TypeType }

type Comma struct{}

func (Comma) Type() TokenType { return CommaType }
