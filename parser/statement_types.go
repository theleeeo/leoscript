package parser

import "leoscript/types"

type Statement interface{}

type VarDecl struct {
	Name  string
	Type  types.Type
	Value Expression
}
