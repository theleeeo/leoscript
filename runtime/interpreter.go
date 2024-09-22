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

	program, err := parser.NewParser(tokens, nil).ParseFile()
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	for _, stmt := range program.Body {
		intr.evaluateStatement(stmt)
	}

	return nil
}

func (intr *Interpreter) Run() (val runtimeVal, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = nil
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	for fnName, fnDef := range intr.activeScope.functions {
		if fnName == "main" {
			val = intr.callFunction(intr.activeScope, fnDef, nil)
		}
	}

	if val == nil {
		return nil, fmt.Errorf("main function not found")
	}

	if err != nil {
		return nil, err
	}

	return val, nil
}

func New() *Interpreter {
	globalScope := newScope(nil)
	return &Interpreter{
		globalScope: globalScope,
		activeScope: globalScope,
	}
}

type Interpreter struct {
	globalScope *scope
	activeScope *scope
}

func (intr *Interpreter) evaluateStatement(stmt parser.Statement) {
	switch s := stmt.(type) {
	case parser.VarDecl:
		val := intr.evaluateExpression(s.Value)
		if err := intr.activeScope.DeclareVar(s.Name, val); err != nil {
			panic(err)
		}
	case parser.FnDef:
		if err := intr.activeScope.RegisterFn(s.Name, s); err != nil {
			panic(err)
		}
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
		val, ok := intr.activeScope.GetVar(e.Name)
		if !ok {
			// TODO: Handle this better
			panic(fmt.Sprintf("variable %s not defined", e.Name))
		}
		return val

	case parser.Call:
		fn, ok := intr.activeScope.GetFn(e.Name)
		if !ok {
			panic(fmt.Sprintf("function %s not defined", e.Name))
		}

		var parameters []runtimeVal
		for _, arg := range e.Args {
			parameters = append(parameters, intr.evaluateExpression(arg))
		}

		return intr.callFunction(intr.activeScope, fn, parameters)

	default:
		panic(fmt.Sprintf("unknown expression: %T, v=%+v", e, e))
	}
}

func (intr *Interpreter) callFunction(parentScope *scope, fn parser.FnDef, parameters []runtimeVal) runtimeVal {
	if len(parameters) != len(fn.Args) {
		panic(fmt.Sprintf("expected %d arguments, got %d", len(fn.Args), len(parameters)))
	}

	// Create a new scope for the function
	fnScope := newScope(parentScope)

	// Add arguments to the scope
	for i, arg := range fn.Args {
		fnScope.DeclareVar(arg.Name, parameters[i])
	}

	// Set the active scope to the function scope
	intr.activeScope = fnScope

	// Evaluate the function body
	for _, stmt := range fn.Body {
		if retStmt, ok := stmt.(parser.Return); ok {
			return intr.evaluateExpression(retStmt.Value)
		}

		intr.evaluateStatement(stmt)
	}

	// Reset the active scope to the parent scope
	intr.activeScope = parentScope

	return nil
}
