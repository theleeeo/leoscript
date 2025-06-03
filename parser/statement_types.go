package parser

import (
	"leoscript/types"
)

type Statement interface{}

type VarDecl struct {
	Name string
	// The type of the variable will only known at the parsing pass if it's explicitly declared.
	// If it should be inferred it will be Unspecified until the type resolving pass.
	Type  types.Type
	Value Expression
}

type FnDef struct {
	Name       string
	ReturnType types.Type
	Args       []Argument
	Body       []Statement
}

type Return struct {
	Value Expression
}

type Argument struct {
	Name string
	Type types.Type
}

type Assignment struct {
	Name  string
	Value Expression
}

type If struct {
	Cond Expression
	Then []Statement
	Else []Statement
}
