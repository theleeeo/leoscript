package parser

import "fmt"

func TypeResolvingPass(program Program) (pg Program, err error) {
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
		}
		return node
	})

	return tw.WalkProgram(program)
}
