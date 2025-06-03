package token

import (
	"fmt"
	"leoscript/types"
)

type Token interface {
	Type() TokenType
	String() string
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

	// Control Flow
	IfType
	ElseType
	WhileType
)

type EOF struct{}

func (EOF) Type() TokenType { return EOFType }

func (EOF) String() string {
	return "{EOF}"
}

type Integer struct {
	Value int
}

func (Integer) Type() TokenType { return IntegerType }

func (i Integer) String() string {
	return fmt.Sprintf("{int:%d}", i.Value)
}

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

func (t Operator) String() string {
	return fmt.Sprintf("{op:%s}", t.Op)
}

type OpenParen struct{}

func (OpenParen) Type() TokenType { return OpenParenType }

func (OpenParen) String() string {
	return "{(}"
}

type CloseParen struct{}

func (CloseParen) Type() TokenType { return CloseParenType }

func (CloseParen) String() string {
	return "{)}"
}

type Semicolon struct{}

func (Semicolon) Type() TokenType { return SemicolonType }

func (Semicolon) String() string {
	return "{;}"
}

type Identifier struct {
	Value string
}

func (Identifier) Type() TokenType { return IdentifierType }

func (i Identifier) String() string {
	return fmt.Sprintf("{id:%s}", i.Value)
}

type Boolean struct {
	Value bool
}

func (Boolean) Type() TokenType { return BooleanType }

func (b Boolean) String() string {
	if b.Value {
		return "{bool:true}"
	}
	return "{bool:false}"
}

type VarDecl struct{}

func (VarDecl) Type() TokenType { return VarDeclType }

func (VarDecl) String() string {
	return "{var}"
}

type FnDef struct{}

func (FnDef) Type() TokenType { return FnDefType }

func (FnDef) String() string {
	return "{fn}"
}

type OpenBrace struct{}

func (OpenBrace) Type() TokenType { return OpenBraceType }

func (OpenBrace) String() string {
	return "{{}"
}

type CloseBrace struct{}

func (CloseBrace) Type() TokenType { return CloseBraceType }

func (CloseBrace) String() string {
	return "{}}"
}

type Return struct{}

func (Return) Type() TokenType { return ReturnType }

func (Return) String() string {
	return "{return}"
}

type Type struct {
	Kind types.Type
}

func (Type) Type() TokenType { return TypeType }

func (t Type) String() string {
	return fmt.Sprintf("{type:%s}", t.Kind)
}

type Comma struct{}

func (Comma) Type() TokenType { return CommaType }

func (Comma) String() string {
	return "{,}"
}

type If struct{}

func (If) Type() TokenType { return IfType }

func (If) String() string {
	return "{if}"
}

type Else struct{}

func (Else) Type() TokenType { return ElseType }

func (Else) String() string {
	return "{else}"
}

type While struct{}

func (While) Type() TokenType { return WhileType }

func (While) String() string {
	return "{while}"
}
