package parser

import (
	"fmt"
)

func TypeValidationPass(program Program) (pg Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	for i := range program.VarDecls {
		verifyTypesOfStmt(program.VarDecls[i], parsingContext{})
	}

	for i := range program.FnDefs {
		for j := range program.FnDefs[i].Body {
			verifyTypesOfStmt(program.FnDefs[i].Body[j], parsingContext{
				currentFn: &program.FnDefs[i],
			})
		}
	}

	return program, nil
}

type parsingContext struct {
	currentFn *FnDef
}

func verifyTypesOfStmt(stmt Statement, pctx parsingContext) {
	switch stmt := stmt.(type) {
	case VarDecl:
		if stmt.Value.ReturnType() != stmt.Type {
			panic(fmt.Sprintf("type mismatch: expected %s, got %s", stmt.Type, stmt.Value.ReturnType()))
		}
	case Return:
		if stmt.Value.ReturnType() != pctx.currentFn.ReturnType {
			panic(fmt.Sprintf("type mismatch: expected %s, got %s", pctx.currentFn.ReturnType, stmt.Value.ReturnType()))
		}
	// case If:
	// 	if stmt.Cond.ReturnType() != types.Bool {
	// 		panic("if condition must be of type bool")
	// 	}
	default:
		panic(fmt.Sprintf("unhandled statement type %T", stmt))
	}
}

// func verifyTypesOfExpr(expr Expression) {
// 	switch expr := expr.(type) {
// 	case VarIdentifier:
// 		fmt.Println("Verifying variable identifier:", expr.Name)
// 		// varIdent, ok := scope.ResolveVar(expr.Name)
// 		// if !ok {
// 		// 	panic(fmt.Sprint("unknown variable", "name", expr.Name))
// 		// }

// 		// expr.returnType = varIdent.Value.ReturnType()
// 	default:
// 		panic(fmt.Sprintf("unhandled expression type %T", expr))
// 	}
// }
