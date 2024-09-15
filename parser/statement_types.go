package parser

import "leoscript/types"

type Statement interface{}

type VarDef struct {
	Name  string
	Type  types.Type
	Value Expression
}
