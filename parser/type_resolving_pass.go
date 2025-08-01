package parser

import (
	"fmt"
	"leoscript/types"
)

func typeResolvingPass(program *Program) (err error) {
	tw := NewTreeWalker(func(wctx WalkingContext, node Statement) Statement {
		switch expr := node.(type) {
		case Call:
			resFn, ok := wctx.Scope.ResolveFn(expr.Name)
			if !ok {
				panic(fmt.Sprint("unknown function:", expr.Name))
			}

			expr.returnType = resFn.ReturnType

			return expr
		case VarIdentifier:
			varIdent, ok := wctx.Scope.ResolveVar(expr.Name)
			if !ok {
				panic(fmt.Sprint("unknown variable:", expr.Name))
			}

			expr.returnType = varIdent.Type

			return expr
		case VarDecl:
			if expr.Type == types.Unspecified {
				// If the type is unspecified, we need to resolve it from the value.
				expr.Type = expr.Value.ReturnType()
			}

			// Re-register the varDecl to the scope so the variable in the scope contains the correct type.
			wctx.Scope.deregisterVar(expr.Name)
			wctx.Scope.RegisterVar(expr)

			return expr
		}
		return node
	})

	return tw.WalkProgram(program)
}
