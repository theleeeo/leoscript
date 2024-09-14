package runtime

import "leoscript/parser"

func Run(pg parser.Program) int {
	if len(pg.Body) == 0 {
		panic("empty program")
	}

	if len(pg.Body) > 1 {
		panic("only one expression is allowed")
	}

	for _, expr := range pg.Body {
		val := evaluateExpression(expr)
		return val.(numberVal).value
	}

	return 0
}

type runtimeVal interface{}

type numberVal struct {
	value int
}

func evaluateExpression(expr parser.Expression) runtimeVal {
	switch e := expr.(type) {
	case parser.BinaryExpression:
		left := evaluateExpression(e.Left)
		right := evaluateExpression(e.Right)

		switch e.Op {
		case "+":
			return numberVal{value: left.(numberVal).value + right.(numberVal).value}
		case "-":
			return numberVal{value: left.(numberVal).value - right.(numberVal).value}
		case "*":
			return numberVal{value: left.(numberVal).value * right.(numberVal).value}
		case "/":
			if right.(numberVal).value == 0 {
				panic("division by zero")
			}
			return numberVal{value: left.(numberVal).value / right.(numberVal).value}
		default:
			panic("unknown operator")
		}

	case parser.IntegerLiteral:
		return numberVal{value: e.Value}
	default:
		panic("unknown expression")
	}
}
