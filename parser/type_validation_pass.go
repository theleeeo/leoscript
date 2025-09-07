package parser

import (
	"fmt"
	"leoscript/types"
)

func typeValidationPass(program *Program) (err error) {
	tw := NewTreeWalker(TreeWalkerConfig{
		CallbackFn: func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case VarDecl:
				if n.Type.Kind() == types.KindInvalid {
					panic("invalid type")
				}

				if n.Type == types.Void {
					panic("cannot declare variable with void type")
				}

				if n.Value != nil {
					if n.Value.ReturnType() != n.Type {
						panic(fmt.Sprintf("type mismatch: expected %s, got %s", n.Type, n.Value.ReturnType()))
					}
				}

			case Return:
				if n.Value.ReturnType() != wctx.ParentFn.ReturnType {
					panic(fmt.Sprintf("type mismatch: expected %s, got %s", wctx.ParentFn.ReturnType, n.Value.ReturnType()))
				}
			case If:
				if n.Cond.ReturnType() != types.Bool {
					panic(fmt.Sprintf("type mismatch: expected Bool, got %s", n.Cond.ReturnType()))
				}
			case While:
				if n.Cond.ReturnType() != types.Bool {
					panic(fmt.Sprintf("type mismatch: expected Bool, got %s", n.Cond.ReturnType()))
				}
			case Assignment:
				switch n.Target.(type) {
				case VariableTarget:
					vt, _ := wctx.Scope.ResolveVarType(string(n.Target.(VariableTarget)))
					if n.Value.ReturnType() != vt {
						panic(fmt.Sprintf("type mismatch in assignment: expected %s, got %s", vt, n.Value.ReturnType()))
					}
				case IndexTarget:
					vt, _ := wctx.Scope.ResolveVarType(n.Target.(IndexTarget).VariableName)
					at := vt.(types.Array)
					if n.Value.ReturnType() != at.ElementType {
						panic(fmt.Sprintf("type mismatch in assignment: expected %s, got %s", at.ElementType, n.Value.ReturnType()))
					}
				}
			case UnaryExpression:
				switch n.Op {
				case "!":
					if n.Expression.ReturnType() != types.Bool {
						panic(fmt.Sprintf("type mismatch: expected Bool for '!', got %s", n.Expression.ReturnType()))
					}
				case "-":
					if n.Expression.ReturnType() != types.Int {
						panic(fmt.Sprintf("type mismatch: expected Int for '-', got %s", n.Expression.ReturnType()))
					}
				case "+":
					if n.Expression.ReturnType() != types.Int {
						panic(fmt.Sprintf("type mismatch: expected Int for '+', got %s", n.Expression.ReturnType()))
					}
				default:
					panic(fmt.Sprintf("unsupported unary operator: %s", n.Op))
				}
			case BinaryExpression:
				if n.Left.ReturnType() != n.Right.ReturnType() {
					panic(fmt.Sprintf("type mismatch in binary expression: left %s, right %s", n.Left.ReturnType(), n.Right.ReturnType()))
				}

				switch n.Op {
				case "==", "!=", "<", ">", "<=", ">=":
					if n.Left.ReturnType() != types.Int && n.Left.ReturnType() != types.Bool {
						panic(fmt.Sprintf("type mismatch: expected Int or Bool for comparison, got %s", n.Left.ReturnType()))
					}
				case "+", "-", "*", "/":
					if n.Left.ReturnType() != types.Int {
						panic(fmt.Sprintf("type mismatch: expected Int for arithmetic operation, got %s", n.Left.ReturnType()))
					}
				case "&&", "||":
					if n.Left.ReturnType() != types.Bool {
						panic(fmt.Sprintf("type mismatch: expected Bool for logical operation, got %s", n.Left.ReturnType()))
					}
				default:
					panic(fmt.Sprintf("unsupported binary operator: %s", n.Op))
				}
			case ArrayLiteral:
				for i, elem := range n.Elements {
					if n.ElementType != elem.ReturnType() {
						panic(fmt.Sprintf("type mismatch in array literal: element %d is %s, expected %s", i, elem.ReturnType(), n.ElementType))
					}
				}
			}
			return node, nil
		},
	})

	return tw.WalkProgram(program)
}
