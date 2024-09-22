package parser

import (
	"leoscript/token"
	"leoscript/types"
)

type Statement interface{}

type VarDecl struct {
	Name  string
	Type  types.Type
	Value Expression
}

type FnDef struct {
	Name       string
	ReturnType types.Type
	Args       []Argument
	Body       []Statement
	// The unprocessed source code of the function body.
	bodySrc []token.Token
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
