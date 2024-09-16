package parser

import "leoscript/types"

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
