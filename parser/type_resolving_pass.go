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
		switch n := node.(type) {
		case Call:
			resFn, ok := wctx.Scope.ResolveFn(n.Name)
			if !ok {
				panic(fmt.Sprint("unknown function:", n.Name))
			}

			n.returnType = resFn.ReturnType

			return n, nil
		case VarIdentifier:
			vt, ok := wctx.Scope.ResolveVarType(n.Name)
			if !ok {
				panic(fmt.Sprint("unknown variable:", n.Name))
			}

			n.returnType = vt

			return n, nil
		case VarDecl:
			// Already fine
			if n.Type.Kind() != types.KindInvalid {
				return n, nil
			}

			// TODO: Handle array types

			if t, ok := n.Type.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.Name))
				}
				n.Type = resolvedType

				return n, nil
			}

			if t, ok := n.Type.(types.Array); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.ElementType.(unresolvedTypeIdentifier).Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.ElementType.(unresolvedTypeIdentifier).Name))
				}
				t.ElementType = resolvedType
				n.Type = t

				return n, nil
			}

			//
			// Must resolve the type from the value
			//

			if n.Value == nil {
				panic("cannot infer type on an uninitialized variable")
			}

			// Child is a variable identifier whose type is not yet resolved
			if n.Value.ReturnType() == types.Unspecified {
				return nil, ErrReturnLater
			}

			// Child has a type identifier that is not yet resolved
			if _, ok := n.Value.ReturnType().(unresolvedTypeIdentifier); ok {
				return nil, ErrReturnLater
			}

			// If the type is unspecified, we need to resolve it from the value.
			n.Type = n.Value.ReturnType()

			return n, nil
		case StructLiteral:
			if t, ok := n.Type.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.Name))
				}
				n.Type = resolvedType

				return n, nil
			}

			switch pn := wctx.ParentNode.(type) {
			case VarDecl: // If the parent node is a variable declaration, we can infer the type from it.
				// If the variable is implicitly typed and the value is not yet resolved, come back after the child is visited.
				if pn.Type == types.Unspecified {
					panic("cannot use an implicit struct literal without an explicitly typed variable")
				}

				n.Type = pn.Type
			case Assignment:
				switch pn.Target.(type) {
				case VariableTarget:
					vt, ok := wctx.Scope.ResolveVarType(string(pn.Target.(VariableTarget)))
					if !ok {
						panic(fmt.Sprint("unknown variable:", pn.Target.(VariableTarget)))
					}
					n.Type = vt
				case IndexTarget:
					// Handle array index assignments
					vt, ok := wctx.Scope.ResolveVarType(pn.Target.(IndexTarget).VariableName)
					if !ok {
						panic(fmt.Sprint("unknown array:", pn.Target.(IndexTarget).VariableName))
					}
					at := vt.(types.Array)
					n.Type = at.ElementType
				}
			case Call:
				resFn, ok := wctx.Scope.ResolveFn(pn.Name)
				if !ok {
					panic(fmt.Sprint("unknown function:", pn.Name))
				}
				n.Type = resFn.ReturnType
			}

			return n, nil
		case FnDef:
			if t, ok := n.ReturnType.(unresolvedTypeIdentifier); ok {
				resolvedType, ok := wctx.Scope.ResolveType(t.Name)
				if !ok {
					panic(fmt.Sprint("unknown type:", t.Name))
				}
				n.ReturnType = resolvedType

				return n, nil
			}
		case Parameter:
			if n.Type.Kind() != types.KindInvalid {
				return n, nil
			}

			t := n.Type.(unresolvedTypeIdentifier)
			resolvedType, ok := wctx.Scope.ResolveType(t.Name)
			if !ok {
				panic(fmt.Sprint("unknown type:", t.Name))
			}
			n.Type = resolvedType

			return n, nil
		case ArrayLiteral:
			if n.ElementType.Kind() != types.KindInvalid {
				return n, nil
			}

			switch pn := wctx.ParentNode.(type) {
			case Assignment:
				switch pn.Target.(type) {
				case VariableTarget:
					vt, ok := wctx.Scope.ResolveVarType(string(pn.Target.(VariableTarget)))
					if !ok {
						panic(fmt.Sprint("unknown variable:", pn.Target.(VariableTarget)))
					}
					at := vt.(types.Array)
					n.ElementType = at.ElementType
				case IndexTarget:
					// Handle array index assignments
					vt, ok := wctx.Scope.ResolveVarType(string(pn.Target.(IndexTarget).VariableName))
					if !ok {
						panic(fmt.Sprint("unknown array:", pn.Target.(IndexTarget).VariableName))
					}
					at := vt.(types.Array)
					n.ElementType = at.ElementType
				}
			case VarDecl:
				if pn.Type != types.Unspecified {
					n.ElementType = pn.Type
				}
			}

			// Still not resolved, try to infer the type from the elements
			if n.ElementType.Kind() == types.KindInvalid {
				if len(n.Elements) == 0 {
					return nil, fmt.Errorf("cannot infer type of empty array literal")
				}

				// Only use the first element for simplicity and to make the behaviour clear.
				firstElem := n.Elements[0]
				if firstElem.ReturnType() == types.Unspecified {
					return nil, ErrReturnLater
				}

				n.ElementType = firstElem.ReturnType()
			}

			return n, nil
		case ArrayIndex:
			resolvedType, ok := wctx.Scope.ResolveVarType(n.ArrayVar)
			if !ok {
				panic(fmt.Sprint("unknown variable:", n.ArrayVar))
			}
			n.ElementType = resolvedType.(types.Array).ElementType

			return n, nil
		}
		return node, nil
	}

	tw := NewTreeWalker(TreeWalkerConfig{
		CallbackFn: callback,
		BeforeWalk: beforeWalk,
	})

	return tw.WalkProgram(program)
}
