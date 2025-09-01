package runtime

import (
	"fmt"
	"leoscript/compiler"
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
	if len(parameters) != len(fn.Params) {
		panic(fmt.Sprintf("expected %d arguments, got %d", len(fn.Params), len(parameters)))
	}

	if stubFunc, ok := intr.stubs[fn.Name]; ok {
		return stubFunc.Call(parameters)
	}

	panic(fmt.Sprintf("stub function %s not registered", fn.Name))
}

type externalFunction struct {
	md     compiler.ExportedFunction
	fn     reflect.Value
	fnType reflect.Type
}

func (ef *externalFunction) verifySignature(stub parser.FnDef) error {
	if ef.fnType.NumIn() != len(stub.Params) {
		return fmt.Errorf("expected %d parameters, got %d", ef.fnType.NumIn(), len(stub.Params))
	}

	for i := 0; i < ef.fnType.NumIn(); i++ {
		if err := verifyEqualType(stub.Params[i].Type, ef.fnType.In(i)); err != nil {
			return fmt.Errorf("parameter %d: %w", i+1, err)
		}
	}

	if ef.fnType.NumOut() == 1 {
		if stub.ReturnType == types.Void {
			return fmt.Errorf("expected no return value, got %s", stub.ReturnType)
		}

		if err := verifyEqualType(stub.ReturnType, ef.fnType.Out(0)); err != nil {
			return fmt.Errorf("return value: %w", err)
		}
	}

	return nil
}

func (ef *externalFunction) verifySignature2(stub compiler.ExportedFunction) error {
	if ef.fnType.NumIn() != len(stub.Params) {
		return fmt.Errorf("expected %d parameters, got %d", ef.fnType.NumIn(), len(stub.Params))
	}

	for i := range ef.fnType.NumIn() {
		if err := verifyEqualType(stub.Params[i].Type, ef.fnType.In(i)); err != nil {
			return fmt.Errorf("parameter %d: %w", i+1, err)
		}
	}

	if ef.fnType.NumOut() > 1 {
		return fmt.Errorf("expected at most one return value, got %d", ef.fnType.NumOut())
	}

	if ef.fnType.NumOut() == 0 && stub.ReturnType != types.Void {
		return fmt.Errorf("expected a return value of type %s, got void", stub.ReturnType)
	}

	if ef.fnType.NumOut() == 1 {
		if stub.ReturnType == types.Void {
			return fmt.Errorf("expected no return value, got %s", stub.ReturnType)
		}

		if err := verifyEqualType(stub.ReturnType, ef.fnType.Out(0)); err != nil {
			return fmt.Errorf("return value: %w", err)
		}
	}

	return nil
}

func verifyEqualType(lsType types.Type, goType reflect.Type) error {
	switch goType.Kind() { // TODO: Swap this to switch on lsType once TypeKind is implemented
	case reflect.Int:
		if lsType != types.Int {
			return fmt.Errorf("expected %s, got int", lsType)
		}
	case reflect.Bool:
		if lsType != types.Bool {
			return fmt.Errorf("expected %s, got bool", lsType)
		}
	default:
		return fmt.Errorf("unsupported type: %s", goType)
	}
	return nil
}

func (ef *externalFunction) Call(args []runtimeVal) runtimeVal {
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
	case nil:
		return nil
	default:
		panic(fmt.Sprintf("unsupported return type from stub function: %T", v))
	}
}

func (ef *externalFunction) call2(args []uint64) uint64 {
	if len(args) != len(ef.md.Params) {
		panic(fmt.Sprintf("expected %d arguments, got %d", len(ef.md.Params), len(args))) // This should never happen
	}

	inVals := make([]reflect.Value, len(args))
	for i, arg := range args {
		switch ef.md.Params[i].Type.(types.BasicType) {
		case types.Int:
			inVals[i] = reflect.ValueOf(int(arg))
		case types.Bool:
			inVals[i] = reflect.ValueOf(arg != 0)
		default:
			panic(fmt.Sprintf("unsupported argument type for stub function: %T", ef.md.Params[i].Type))
		}
	}

	out := ef.fn.Call(inVals)
	if len(out) == 0 {
		return 0
	}

	// Only use the first return value, more than one return value is not supported
	result := out[0].Interface()
	switch v := result.(type) {
	case int:
		return uint64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		panic(fmt.Sprintf("unsupported return type from stub function: %T", v))
	}
}
