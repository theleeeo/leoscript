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
			vt, ok := wctx.Scope.ResolveVarType(expr.Name)
			if !ok {
				panic(fmt.Sprint("unknown variable:", expr.Name))
			}

			expr.returnType = vt

			return expr, nil
		case VarDecl:
			// Already fine
			if expr.Type.Kind() != types.KindInvalid {
				return expr, nil
			}

			if t, ok := expr.Type.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.Name))
				}
				expr.Type = resolvedType

				return expr, nil
			}

			//
			// Must resolve the type from the value
			//

			if expr.Value == nil {
				panic("cannot infer type on an uninitialized variable")
			}

			// Child is a variable identifier whose type is not yet resolved
			if expr.Value.ReturnType() == types.Unspecified {
				return nil, ErrReturnLater
			}

			// Child has a type identifier that is not yet resolved
			if _, ok := expr.Value.ReturnType().(unresolvedTypeIdentifier); ok {
				return nil, ErrReturnLater
			}

			// If the type is unspecified, we need to resolve it from the value.
			expr.Type = expr.Value.ReturnType()

			return expr, nil
		case StructLiteral:
			if t, ok := expr.Type.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.Name))
				}
				expr.Type = resolvedType

				return expr, nil
			}

			switch pn := wctx.ParentNode.(type) {
			case VarDecl: // If the parent node is a variable declaration, we can infer the type from it.
				// If the variable is implicitly typed and the value is not yet resolved, come back after the child is visited.
				if pn.Type == types.Unspecified {
					panic("cannot use an implicit struct literal without an explicitly typed variable")
				}

				expr.Type = pn.Type
			case Assignment:
				vt, ok := wctx.Scope.ResolveVarType(pn.Name)
				if !ok {
					panic(fmt.Sprint("unknown variable:", pn.Name))
				}
				expr.Type = vt
			case Call:
				resFn, ok := wctx.Scope.ResolveFn(pn.Name)
				if !ok {
					panic(fmt.Sprint("unknown function:", pn.Name))
				}
				expr.Type = resFn.ReturnType
			}

			return expr, nil
		case FnDef:
			if t, ok := expr.ReturnType.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.Name))
				}
				expr.ReturnType = resolvedType

				return expr, nil
			}
		case Parameter:
			if expr.Type.Kind() != types.KindInvalid {
				return expr, nil
			}

			t := expr.Type.(unresolvedTypeIdentifier)
			resolvedType, ok := wctx.Scope.ResolveType(t.Name)
			if !ok {
				panic(fmt.Sprint("unknown type:", t.Name))
			}
			expr.Type = resolvedType

			return expr, nil
		}
		return node, nil
	})

	return tw.WalkProgram(program)
}
