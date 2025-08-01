package runtime

import (
	"fmt"
	"leoscript/parser"
	"leoscript/types"
	"reflect"
)

// RegisterStub allows registering any Go function as a stub, using reflection to map runtimeVals to Go values and back.
func (intr *Interpreter) RegisterStub(name string, fn any) *Interpreter {
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

	return intr
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

func (intr *Interpreter) callStub(fn parser.FnDef, parameters []runtimeVal) runtimeVal {
	if len(parameters) != len(fn.Args) {
		panic(fmt.Sprintf("expected %d arguments, got %d", len(fn.Args), len(parameters)))
	}

	if stubFunc, ok := intr.stubs[fn.Name]; ok {
		return stubFunc.Call(parameters)
	}

	panic(fmt.Sprintf("stub function %s not registered", fn.Name))
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
