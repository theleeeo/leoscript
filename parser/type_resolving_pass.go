package parser

import (
	"fmt"
	"leoscript/types"
)

func typeResolvingPass(program *Program) (err error) {
	tw := NewTreeWalker(func(wctx WalkingContext, node Statement) (Statement, error) {
		switch expr := node.(type) {
		case Call:
			resFn, ok := wctx.Scope.ResolveFn(expr.Name)
			if !ok {
				panic(fmt.Sprint("unknown function:", expr.Name))
			}

			expr.returnType = resFn.ReturnType

			return expr, nil
		case VarIdentifier:
			varIdent, ok := wctx.Scope.ResolveVar(expr.Name)
			if !ok {
				panic(fmt.Sprint("unknown variable:", expr.Name))
			}

			expr.returnType = varIdent.Type

			return expr, nil
		case VarDecl:
			if expr.Type == types.Unspecified {
				// If the variable is implicitly typed and the value is not yet resolved, come back after the child is visited.
				if expr.Value.ReturnType() == types.Unspecified {
					return nil, ErrReturnLater
				}
				// If the type is unspecified, we need to resolve it from the value.
				expr.Type = expr.Value.ReturnType()
			}

			return expr, nil
		}
		return node, nil
	})

	return tw.WalkProgram(program)
}
