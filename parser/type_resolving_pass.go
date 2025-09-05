package parser

import (
	"fmt"
	"leoscript/types"
)

func typeResolvingPass(program *Program) (err error) {
	// BeforeWalk resolves function return types that are unresolved at the start of the walk.
	// This is needed because the ReturnType might be used to infer implicit variables and they are walked before the actual functions by the treewalker.
	beforeWalk := func(pg *Program, globalScope *Scope) error { // TODO: This can be implemented as its own pass...
		for i := range pg.FnDefs {
			if pg.FnDefs[i].ReturnType.Kind() != types.KindInvalid {
				continue
			}

			if ti, ok := pg.FnDefs[i].ReturnType.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := globalScope.ResolveType(ti.Name)
				if !ok {
					return fmt.Errorf("unknown type: %s", ti.Name)
				}
				// Set the function definition's return type to the resolved type.
				pg.FnDefs[i].ReturnType = resolvedType

				// Update the function in the scope.
				globalScope.deregisterFn(pg.FnDefs[i].Name)
				globalScope.RegisterFn(pg.FnDefs[i])
			}
		}

		for i := range pg.StubDefs {
			if pg.StubDefs[i].ReturnType.Kind() != types.KindInvalid {
				continue
			}

			if ti, ok := pg.StubDefs[i].ReturnType.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := globalScope.ResolveType(ti.Name)
				if !ok {
					return fmt.Errorf("unknown type: %s", ti.Name)
				}
				// Set the function definition's return type to the resolved type.
				pg.StubDefs[i].ReturnType = resolvedType

				// Update the function in the scope.
				globalScope.deregisterFn(pg.StubDefs[i].Name)
				globalScope.RegisterFn(pg.StubDefs[i])
			}
		}

		for i := range pg.Structs {
			for j := range pg.Structs[i].Fields {
				if pg.Structs[i].Fields[j].Type.Kind() != types.KindInvalid {
					continue
				}

				if ti, ok := pg.Structs[i].Fields[j].Type.(unresolvedTypeIdentifier); ok {
					resolvedType, ok := globalScope.ResolveType(ti.Name)
					if !ok {
						return fmt.Errorf("unknown type: %s", ti.Name)
					}
					pg.Structs[i].Fields[j].Type = resolvedType
				}
			}
		}

		return nil
	}

	callback := func(wctx WalkingContext, node Statement) (Statement, error) {
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
	}

	tw := NewTreeWalker(TreeWalkerConfig{
		CallbackFn: callback,
		BeforeWalk: beforeWalk,
	})

	return tw.WalkProgram(program)
}
