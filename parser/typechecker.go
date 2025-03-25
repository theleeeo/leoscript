package parser

import "fmt"

func TypeResolvingPass(program Program, scope *Scope) (pg Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	for _, varDecl := range program.VarDecls {
		scope.RegisterVar(varDecl)
	}

	for _, fn := range program.FnDefs {
		scope.RegisterFn(fn)
	}

	for i := range program.FnDefs {
		for j := range program.FnDefs[i].Body {
			program.FnDefs[i].Body[j] = resolveTypesOfStmt(program.FnDefs[i].Body[j], scope)
		}
	}

	for i := range program.VarDecls {
		program.VarDecls[i].Value = resolveTypesOfExpr(program.VarDecls[i].Value, scope)
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
		return stmt
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
			panic(fmt.Sprint("unknown function", "name", expr.Name))
		}

		expr.returnType = resFn.ReturnType

		return expr
	case VarIdentifier:
		varIdent, ok := scope.ResolveVar(expr.Name)
		if !ok {
			panic(fmt.Sprint("unknown variable", "name", expr.Name))
		}

		expr.returnType = varIdent.Value.ReturnType()

		return expr
	default:
		return expr
	}
}

// func TypeValidationPass(program Program, scope *Scope) { }
