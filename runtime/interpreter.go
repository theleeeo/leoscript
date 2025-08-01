package runtime

import (
	"fmt"
	"leoscript/lexer"
	"leoscript/parser"
	"leoscript/types"
	"reflect"
)

func (intr *Interpreter) LoadRaw(src string) error {
	if src == "" {
		return fmt.Errorf("empty source")
	}

	tokens, err := lexer.Tokenize(src)
	if err != nil {
		return fmt.Errorf("tokenizing: %w", err)
	}

	program, err := parser.NewParser(tokens).ParseFile()
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	for _, stmt := range program.FnDefs {
		intr.evaluateStatement(stmt)
	}

	for _, stmt := range program.VarDecls {
		intr.evaluateStatement(stmt)
	}

	for _, stmt := range program.StubDefs {
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

	if err := intr.verifyStubs(); err != nil {
		return nil, fmt.Errorf("stub verification failed: %w", err)
	}

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

func (intr *Interpreter) verifyStubs() error {
	for name, stub := range intr.globalScope.functions {
		if !stub.Stub {
			continue
		}
		if ef, ok := intr.stubs[name]; ok {
			if err := ef.verifySignature(stub); err != nil {
				return fmt.Errorf("stub function %s signature mismatch: %w", name, err)
			}
		} else {
			return fmt.Errorf("stub function %s not registered", name)
		}
	}

	return nil
}

func New() *Interpreter {
	globalScope := newScope(nil)
	return &Interpreter{
		globalScope: globalScope,
		activeScope: globalScope,
		stubs:       make(map[string]externalFunction),
	}
}

type Interpreter struct {
	globalScope *scope
	activeScope *scope
	stubs       map[string]externalFunction
}

type externalFunction struct {
	fn     reflect.Value
	fnType reflect.Type
}

func (ef externalFunction) verifySignature(stub parser.FnDef) error {
	if ef.fnType.NumIn() != len(stub.Args) {
		return fmt.Errorf("expected %d arguments, got %d", ef.fnType.NumIn(), len(stub.Args))
	}

	for i := 0; i < ef.fnType.NumIn(); i++ {
		if err := verifyEqualType(ef.fnType.In(i), stub.Args[i].Type); err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}
	}

	if ef.fnType.NumOut() == 1 {
		if stub.ReturnType == types.Void {
			return fmt.Errorf("expected no return value, got %s", stub.ReturnType)
		}

		if err := verifyEqualType(ef.fnType.Out(0), stub.ReturnType); err != nil {
			return fmt.Errorf("return value: %w", err)
		}
	}

	return nil
}

func verifyEqualType(goType reflect.Type, lsType types.Type) error {
	switch goType.Kind() {
	case reflect.Int:
		if lsType != types.Int {
			return fmt.Errorf("expected int, got %s", lsType)
		}
	case reflect.Bool:
		if lsType != types.Bool {
			return fmt.Errorf("expected bool, got %s", lsType)
		}
	default:
		return fmt.Errorf("unsupported type: %s", goType)
	}
	return nil
}

func (ef externalFunction) Call(args []runtimeVal) runtimeVal {
	if ef.fnType.NumIn() != len(args) {
		panic(fmt.Sprintf("expected %d arguments, got %d", ef.fnType.NumIn(), len(args)))
	}

	inVals := make([]reflect.Value, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case numberVal:
			inVals[i] = reflect.ValueOf(v.value)
		case booleanVal:
			inVals[i] = reflect.ValueOf(v.value)
		default:
			panic(fmt.Sprintf("unsupported argument type for stub function: %T", v))
		}
	}

	out := ef.fn.Call(inVals)
	if len(out) == 0 {
		return nil
	}

	// Only use the first return value
	result := out[0].Interface()
	switch v := result.(type) {
	case int:
		return numberVal{value: v}
	case bool:
		return booleanVal{value: v}
	case float64:
		return numberVal{value: int(v)}
	case nil:
		return nil
	default:
		panic(fmt.Sprintf("unsupported return type from stub function: %T", v))
	}
}

// RegisterStub allows registering any Go function as a stub, using reflection to map runtimeVals to Go values and back.
func (intr *Interpreter) RegisterStub(name string, fn any) {
	if _, exists := intr.stubs[name]; exists {
		panic(fmt.Sprintf("stub function %s already registered", name))
	}

	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected a function, got %s", fnType.Kind()))
	}

	if fnType.IsVariadic() {
		panic("variadic functions are not supported as stubs")
	}

	if fnType.NumOut() > 1 {
		panic(fmt.Errorf("only one return value is supported, got %d", fnType.NumOut()))
	}

	ef := externalFunction{
		fn:     fnVal,
		fnType: fnType,
	}

	intr.stubs[name] = ef
}

// Evaluates a statement. If the statement contains a return, it es evaluates and its value is returned from this function.
// That means that if this functions return value is != nil, the caller should return.
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

	// Evaluate the function body
	for _, stmt := range fn.Body {
		if retVal := intr.evaluateStatement(stmt); retVal != nil {
			return retVal
		}
	}

	// Reset the active scope to the parent scope
	intr.activeScope = parentScope

	return nil
}
func (intr *Interpreter) callStub(fn parser.FnDef, parameters []runtimeVal) runtimeVal {
	if len(parameters) != len(fn.Args) {
		panic(fmt.Sprintf("expected %d arguments, got %d", len(fn.Args), len(parameters)))
	}

	if stubFunc, ok := intr.stubs[fn.Name]; ok {
		return stubFunc.Call(parameters)
	}

	panic(fmt.Sprintf("stub function %s not registered", fn.Name))
}
