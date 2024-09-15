package runtime

import (
	"fmt"
	"leoscript/lexer"
	"leoscript/parser"
	"leoscript/types"
)

func (intr *Interpreter) LoadRaw(src string) error {
	if src == "" {
		return fmt.Errorf("empty source")
	}

	tokens, err := lexer.Tokenize(src)
	if err != nil {
		return fmt.Errorf("failed to tokenize: %w", err)
	}

	program, err := parser.NewParser(tokens).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	intr.program = program

	return nil
}

func (intr *Interpreter) Run() (val runtimeVal, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = nil
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	for _, st := range intr.program.Body {
		return intr.evaluateStatement(st), nil
	}

	panic("unreachable")
}

func New() *Interpreter {
	return &Interpreter{
		scope: newScope(nil),
	}
}

type Interpreter struct {
	scope *scope

	program parser.Program
}

// Todo: Should not return runtimeVal
func (intr *Interpreter) evaluateStatement(stmt parser.Statement) runtimeVal {
	switch s := stmt.(type) {
	case parser.Expression:
		return intr.evaluateExpression(s)
	case parser.VarDecl:
		val := intr.evaluateExpression(s.Value)
		intr.scope.DeclareVar(s.Name, val)
		return val
	// case parser.Assignment:

	default:
		panic(fmt.Sprintf("unknown statement: %T, v=%+v", s, s))
	}
}

func (intr *Interpreter) evaluateExpression(expr parser.Expression) runtimeVal {
	switch e := expr.(type) {
	case parser.BinaryExpression:
		left := intr.evaluateExpression(e.Left)
		right := intr.evaluateExpression(e.Right)

		switch e.Op {
		// Arithmetic
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

		// Boolean
		case "&&":
			return booleanVal{value: left.(booleanVal).value && right.(booleanVal).value}
		case "||":
			return booleanVal{value: left.(booleanVal).value || right.(booleanVal).value}
		case "<":
			return booleanVal{value: left.(numberVal).value < right.(numberVal).value}
		case ">":
			return booleanVal{value: left.(numberVal).value > right.(numberVal).value}
		case "<=":
			return booleanVal{value: left.(numberVal).value <= right.(numberVal).value}
		case ">=":
			return booleanVal{value: left.(numberVal).value >= right.(numberVal).value}

		// Equality
		case "==":
			if left.Type() == types.Int {
				return booleanVal{value: left.(numberVal).value == right.(numberVal).value}
			}
			return booleanVal{value: left.(booleanVal).value == right.(booleanVal).value}
		case "!=":
			if left.Type() == types.Int {
				return booleanVal{value: left.(numberVal).value != right.(numberVal).value}
			}
			return booleanVal{value: left.(booleanVal).value != right.(booleanVal).value}

		default:
			panic(fmt.Sprintf("unknown operator: %s", e.Op))
		}

	case parser.UnaryExpression:
		val := intr.evaluateExpression(e.Expression)

		switch e.Op {
		case "-":
			return numberVal{value: -val.(numberVal).value}
		case "+":
			return numberVal{value: val.(numberVal).value}
		case "!":
			return booleanVal{value: !val.(booleanVal).value}
		default:
			panic(fmt.Sprintf("unknown operator: %s", e.Op))
		}

	case parser.IntegerLiteral:
		return numberVal{value: e.Value}

	case parser.BooleanLiteral:
		return booleanVal{value: e.Value}

	case parser.Identifier:
		val, ok := intr.scope.GetVar(e.Name)
		if !ok {
			// TODO: Handle this better
			panic(fmt.Sprintf("variable %s not defined", e.Name))
		}
		return val

	default:
		panic(fmt.Sprintf("unknown expression: %T, v=%+v", e, e))
	}
}
