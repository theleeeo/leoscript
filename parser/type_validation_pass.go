package parser

import (
	"fmt"
	"leoscript/types"
)

func TypeValidationPass(program Program) (pg Program, err error) {
	tw := NewTreeWalker(func(wctx WalkingContext, node Statement) Statement {
		switch expr := node.(type) {
		case VarDecl:
			if expr.Value.ReturnType() != expr.Type {
				panic(fmt.Sprintf("type mismatch: expected %s, got %s", expr.Type, expr.Value.ReturnType()))
			}
		case Return:
			if expr.Value.ReturnType() != wctx.ParentFn.ReturnType {
				panic(fmt.Sprintf("type mismatch: expected %s, got %s", wctx.ParentFn.ReturnType, expr.Value.ReturnType()))
			}
		case If:
			if expr.Cond.ReturnType() != types.Bool {
				panic(fmt.Sprintf("type mismatch: expected Bool, got %s", expr.Cond.ReturnType()))
			}
		}
		return node
	})

	return tw.WalkProgram(program)
}
