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

	// Indicates if this variable is exported and can be accessed from outside the VM.
	Exported bool
}

type FnDef struct {
	Name       string
	ReturnType types.Type
	Params     []Parameter
	Body       []Statement

	// Indicates if this function is a stub, meaning it has no body and is only declared.
	// Its body will be supplied from an external source.
	Stub bool

	// Indicates if this function is exported, meaning it can be accessed from outside the VM.
	Exported bool
}

type Return struct {
	Value Expression
}

type Parameter struct {
	Name string
	Type types.Type
}

type Assignment struct {
	Target Assignable
	Value  Expression
}

type If struct {
	Cond Expression
	Then []Statement
	Else []Statement
}

type While struct {
	Cond Expression
	Body []Statement
}

type Assignable interface {
	isAssignable()
}
