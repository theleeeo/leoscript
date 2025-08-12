package runtime

import (
	"fmt"
	"leoscript/parser"
	"leoscript/types"
)

func NewInterpreter(program *parser.Program) *Interpreter {
	return &Interpreter{
		program: program,
		stubs:   make(map[string]externalFunction),
	}
}

type Interpreter struct {
	program *parser.Program

	globalScope *scope
	activeScope *scope
	stubs       map[string]externalFunction
}

func (intr *Interpreter) Initialize() (*Interpreter, error) {
	globalScope := newScope(nil)
	intr.globalScope = globalScope
	intr.activeScope = globalScope

	// If the program is nil, the interpreter is "hollow".
	if intr.program != nil {
		for _, fnDef := range intr.program.FnDefs {
			intr.evaluateStatement(fnDef)
		}

		for _, stubDef := range intr.program.StubDefs {
			intr.evaluateStatement(stubDef)
		}

		for _, varDecl := range intr.program.VarDecls {
			intr.evaluateStatement(varDecl)
		}

	}

	if err := intr.verifyStubs(); err != nil {
		return nil, fmt.Errorf("stub verification failed: %w", err)
	}

	return intr, nil
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

func (intr *Interpreter) Invoke(funcName string, parameters ...any) (any, error) {
	fn, ok := intr.activeScope.GetFn(funcName)
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	if fn.Stub {
		vals := make([]runtimeVal, len(parameters))
		for i, param := range parameters {
			vals[i] = toRuntimeVal(param)
		}

		retVal := intr.callStub(fn, vals)
		if retVal == nil {
			return nil, nil // No return value
		}
		return fromRuntimeVal(retVal), nil
	}

	var runtimeParams []runtimeVal
	for _, param := range parameters {
		runtimeParams = append(runtimeParams, toRuntimeVal(param))
	}

	retVal := intr.callFunction(intr.activeScope, fn, runtimeParams)
	if retVal == nil {
		return nil, nil // No return value
	}
	return fromRuntimeVal(retVal), nil
}

// Evaluates a statement. If the statement contains a return, it es evaluates and its value is returned from this function.
// That means that if this functions return value is != nil, the caller should return.
// TODO: Disallow before initialization
func (intr *Interpreter) evaluateStatement(stmt parser.Statement) runtimeVal {
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
	case parser.If:
		cond := intr.evaluateExpression(s.Cond)
		if cond.Type() != types.Bool { // TODO: Remove these runtime checks when the type validation pass is fully implemented
			panic("if condition must be a boolean")
		}

		if cond.(booleanVal).value {
			for _, stmt := range s.Then {
				if ret := intr.evaluateStatement(stmt); ret != nil {
					return ret
				}
			}
		} else {
			if s.Else != nil {
				for _, stmt := range s.Else {
					if ret := intr.evaluateStatement(stmt); ret != nil {
						return ret
					}
				}
			}
		}
	case parser.Return:
		return intr.evaluateExpression(s.Value)
	case parser.While:
		for {
			cond := intr.evaluateExpression(s.Cond)
			if cond.Type() != types.Bool { // TODO: Remove these runtime checks when the type validation pass is fully implemented
				panic("while condition must be a boolean")
			}
			if !cond.(booleanVal).value {
				break
			}
			for _, stmt := range s.Body {
				if ret := intr.evaluateStatement(stmt); ret != nil {
					return ret
				}
			}
		}
	case parser.Assignment:
		val := intr.evaluateExpression(s.Value)
		if err := intr.activeScope.SetVar(s.Name, val); err != nil {
			panic(fmt.Sprintf("assignment error: %v", err))
		}
	case parser.Call:
		// A function call in a place where a statement is expected (e.g. not in an expression context)
		// means that we are calling a function for its side effects, not its return value.
		_ = intr.evaluateExpression(s)
	default:
		panic(fmt.Sprintf("unknown statement: %T, v=%+v", s, s))
	}

	return nil
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

	case parser.VarIdentifier:
		val, ok := intr.activeScope.GetVar(e.Name)
		if !ok {
			panic(fmt.Sprintf("variable %s not defined", e.Name)) // TODO: remove this once we have a validation pass
		}
		return val

	case parser.Call:
		fn, ok := intr.activeScope.GetFn(e.Name)
		if !ok {
			panic(fmt.Sprintf("function %s not defined", e.Name)) // TODO: remove this once we have a validation pass
		}

		var parameters []runtimeVal
		for _, arg := range e.Args {
			parameters = append(parameters, intr.evaluateExpression(arg))
		}

		if fn.Stub {
			return intr.callStub(fn, parameters)
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
	defer func() {
		// Reset the active scope to the parent scope
		intr.activeScope = parentScope
	}()

	// Evaluate the function body
	for _, stmt := range fn.Body {
		if retVal := intr.evaluateStatement(stmt); retVal != nil {
			return retVal
		}
	}

	return nil
}
