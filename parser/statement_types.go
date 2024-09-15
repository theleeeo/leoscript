package parser

import "leoscript/types"

type Statement interface{}

type VarDecl struct {
	Name  string
	Type  types.Type
	Value Expression
}

type FnDef struct {
	Name string
	Args []string
	Body []Statement
}

type Return struct {
	Value Expression
}
