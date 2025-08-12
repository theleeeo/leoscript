package parser

import (
	"fmt"
	"leoscript/types"
)

func typeValidationPass(program *Program) (err error) {
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
		case While:
			if expr.Cond.ReturnType() != types.Bool {
				panic(fmt.Sprintf("type mismatch: expected Bool, got %s", expr.Cond.ReturnType()))
			}
		case Assignment:
			v, _ := wctx.Scope.ResolveVar(expr.Name)
			if expr.Value.ReturnType() != v.Type {
				panic(fmt.Sprintf("type mismatch in assignment: expected %s, got %s", v.Type, expr.Value.ReturnType()))
			}
		case UnaryExpression:
			switch expr.Op {
			case "!":
				if expr.Expression.ReturnType() != types.Bool {
					panic(fmt.Sprintf("type mismatch: expected Bool for '!', got %s", expr.Expression.ReturnType()))
				}
			case "-":
				if expr.Expression.ReturnType() != types.Int {
					panic(fmt.Sprintf("type mismatch: expected Int for '-', got %s", expr.Expression.ReturnType()))
				}
			case "+":
				if expr.Expression.ReturnType() != types.Int {
					panic(fmt.Sprintf("type mismatch: expected Int for '+', got %s", expr.Expression.ReturnType()))
				}
			default:
				panic(fmt.Sprintf("unsupported unary operator: %s", expr.Op))
			}
		case BinaryExpression:
			if expr.Left.ReturnType() != expr.Right.ReturnType() {
				panic(fmt.Sprintf("type mismatch in binary expression: left %s, right %s", expr.Left.ReturnType(), expr.Right.ReturnType()))
			}

			switch expr.Op {
			case "==", "!=", "<", ">", "<=", ">=":
				if expr.Left.ReturnType() != types.Int && expr.Left.ReturnType() != types.Bool {
					panic(fmt.Sprintf("type mismatch: expected Int or Bool for comparison, got %s", expr.Left.ReturnType()))
				}
			case "+", "-", "*", "/":
				if expr.Left.ReturnType() != types.Int {
					panic(fmt.Sprintf("type mismatch: expected Int for arithmetic operation, got %s", expr.Left.ReturnType()))
				}
			case "&&", "||":
				if expr.Left.ReturnType() != types.Bool {
					panic(fmt.Sprintf("type mismatch: expected Bool for logical operation, got %s", expr.Left.ReturnType()))
				}
			default:
				panic(fmt.Sprintf("unsupported binary operator: %s", expr.Op))
			}
		}
		return node
	})

	return tw.WalkProgram(program)
}
