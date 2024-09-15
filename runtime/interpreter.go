package runtime

import (
	"fmt"
	"leoscript/lexer"
	"leoscript/parser"
)

func RunRaw(src string) (runtimeVal, error) {
	tokens, err := lexer.Tokenize(src)
	if err != nil {
		return 0, fmt.Errorf("failed to tokenize: %w", err)
	}

	program, err := parser.NewParser(tokens).Parse()
	if err != nil {
		return 0, fmt.Errorf("failed to parse: %w", err)
	}

	return Run(program), nil
}

func Run(pg parser.Program) runtimeVal {
	if len(pg.Body) == 0 {
		panic("empty program")
	}

	if len(pg.Body) > 1 {
		panic("only one expression is allowed")
	}

	for _, expr := range pg.Body {
		return evaluateExpression(expr)
	}

	return 0
}

type runtimeVal interface{}

type numberVal struct {
	value int
}

type booleanVal struct {
	value bool
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
		case "&&":
			return booleanVal{value: left.(booleanVal).value && right.(booleanVal).value}
		case "||":
			return booleanVal{value: left.(booleanVal).value || right.(booleanVal).value}
		default:
			panic(fmt.Sprintf("unknown operator: %s", e.Op))
		}

	case parser.IntegerLiteral:
		return numberVal{value: e.Value}

	case parser.BooleanLiteral:
		return booleanVal{value: e.Value}

	default:
		panic(fmt.Sprintf("unknown expression: %s", e))
	}
}
