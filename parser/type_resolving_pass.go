package parser

import "fmt"

func TypeResolvingPass(program Program) (pg Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	globalScope := NewScope(nil)

	for _, varDecl := range program.VarDecls {
		globalScope.RegisterVar(varDecl)
	}

	for _, fn := range program.FnDefs {
		globalScope.RegisterFn(fn)
	}

	for i := range program.FnDefs {
		functionScope := NewScope(globalScope)

		// Register function parameters in the function scope
		for _, param := range program.FnDefs[i].Args {
			functionScope.RegisterVar(VarDecl{
				Name: param.Name,
				Type: param.Type,
			})
		}

		// Resolve types of function body statements
		for j := range program.FnDefs[i].Body {
			program.FnDefs[i].Body[j] = resolveTypesOfStmt(program.FnDefs[i].Body[j], functionScope)
		}
	}

	for i := range program.VarDecls {
		program.VarDecls[i].Value = resolveTypesOfExpr(program.VarDecls[i].Value, globalScope)
	}

	return program, nil
}

func resolveTypesOfStmt(stmt Statement, scope *Scope) Statement {
	switch stmt := stmt.(type) {
	case VarDecl:
		stmt.Value = resolveTypesOfExpr(stmt.Value, scope)
		return stmt
	case Return:
		stmt.Value = resolveTypesOfExpr(stmt.Value, scope)
		return stmt
	case Assignment:
		stmt.Value = resolveTypesOfExpr(stmt.Value, scope)
		return stmt
	case If:
		stmt.Cond = resolveTypesOfExpr(stmt.Cond, scope)
		return stmt
	default:
		panic(fmt.Sprintf("unhandled statement type: %T", stmt))
	}
}

func resolveTypesOfExpr(expr Expression, scope *Scope) Expression {
	switch expr := expr.(type) {
	case BinaryExpression:
		expr.Left = resolveTypesOfExpr(expr.Left, scope)
		expr.Right = resolveTypesOfExpr(expr.Right, scope)

		return expr
	case UnaryExpression:
		expr.Expression = resolveTypesOfExpr(expr.Expression, scope)
		return expr
	case Call:
		resFn, ok := scope.ResolveFn(expr.Name)
		if !ok {
			panic(fmt.Sprint("unknown function:", expr.Name))
		}

		expr.returnType = resFn.ReturnType

		return expr
	case VarIdentifier:
		varIdent, ok := scope.ResolveVar(expr.Name)
		if !ok {
			panic(fmt.Sprint("unknown variable:", expr.Name))
		}

		expr.returnType = varIdent.Type

		return expr
	case IntegerLiteral:
		return expr
	case BooleanLiteral:
		return expr
	default:
		panic(fmt.Sprintf("unhandled expression type: %T", expr))
	}
}
