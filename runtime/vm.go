package runtime

import (
	"encoding/binary"
	"fmt"
	"leoscript/compiler"
	"leoscript/types"
)

type VM struct {
	cStack        []uint64
	variableStack []byte
	callStack     []stackFrame

	metadata *compiler.Metadata // Metadata for the program, if available
	program  []byte
	pc       uint64
}

type stackFrame struct {
	// The return address to jump back to after the function call
	// This is the address of the instruction after the one that made the call
	returnAddress uint64
	// The base of the stack for this function call
	// This is used to calculate the offsets of local variables
	// relative to the base of the stack
	stackBase uint64
}

func NewVM(program []byte) *VM {
	exe := new(compiler.Executable)
	if err := exe.Unmarshal(program); err != nil {
		panic(fmt.Sprintf("unmarshalling executable: %v", err))
	}

	vm := &VM{
		metadata:      exe.Metadata(),
		program:       exe.Code(),
		pc:            0,
		cStack:        make([]uint64, 0),
		variableStack: make([]byte, 0),
	}

	_, err := vm.invokeRaw(0) // Initialize the VM by invoking the _init function
	if err != nil {
		panic(fmt.Sprintf("initializing VM: %v", err)) // TODO: No panics
	}

	return vm
}

func Evaluate(program []byte) (int, error) {
	vm := &VM{
		program:       program,
		pc:            0,
		cStack:        make([]uint64, 0),
		variableStack: make([]byte, 0),
	}

	r, err := vm.Run()
	return int(r), err
}

func (vm *VM) pop() uint64 {
	value := vm.cStack[len(vm.cStack)-1]
	vm.cStack = vm.cStack[:len(vm.cStack)-1]
	return value
}

func (vm *VM) Reset() {
	clear(vm.cStack)
	clear(vm.variableStack)
	clear(vm.callStack)
	vm.pc = 0
}

func (vm *VM) storeVariable(varOffset uint64, value uint64) {
	if int(varOffset) > len(vm.variableStack) {
		panic(fmt.Sprintf("Variable offset %d is out of bounds for variable stack of length %d", varOffset, len(vm.variableStack))) // Should not be able to happen
	}

	// If the variable is supposed to be stored at the end of the variable stack,
	// we need to ensure that the variable stack is large enough
	if int(varOffset) == len(vm.variableStack) {
		vm.variableStack = binary.BigEndian.AppendUint64(vm.variableStack, value)
		return
	}
	// Store the value in the variable stack
	binary.BigEndian.PutUint64(vm.variableStack[varOffset:varOffset+8], value)
}

func (vm *VM) loadVariable(varOffset uint64) uint64 {
	if int(varOffset) > len(vm.variableStack) {
		panic(fmt.Sprintf("Variable offset %d is out of bounds for variable stack of length %d", varOffset, len(vm.variableStack))) // Should not be able to happen
	}

	return binary.BigEndian.Uint64(vm.variableStack[varOffset : varOffset+8])
}

func (vm *VM) Run() (uint64, error) {
	for {
		op := vm.program[vm.pc]

		// PROPOSAL: The binary (and boolean ops) could be changed to be one byte for a bin-op and then one byte (or merge them together using bit-magic) for the specific operation.
		// It is done in a similar way but only for boolean operations in the python bytecode.
		// Test this to check for performance benefits.
		switch op {
		case compiler.OpPush:
			value := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.cStack = append(vm.cStack, value)
			vm.pc += 8
		case compiler.OpAdd:
			a := vm.pop()
			b := vm.pop()
			vm.cStack = append(vm.cStack, b+a)
		case compiler.OpSub:
			a := vm.pop()
			b := vm.pop()
			vm.cStack = append(vm.cStack, b-a)
		case compiler.OpMul:
			a := vm.pop()
			b := vm.pop()
			vm.cStack = append(vm.cStack, b*a)
		case compiler.OpDiv:
			a := vm.pop()
			b := vm.pop()
			vm.cStack = append(vm.cStack, b/a)
		case compiler.OpStore:
			varOffset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc += 8

			sf := vm.callStack[len(vm.callStack)-1] // Get the current stack frame

			value := vm.pop()
			vm.storeVariable(sf.stackBase+varOffset, value)
		case compiler.OpStoreGlobal:
			varOffset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc += 8

			value := vm.pop()
			vm.storeVariable(varOffset, value)
		case compiler.OpLoad: // Load relative to the stackframe base
			varOffset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc += 8

			sf := vm.callStack[len(vm.callStack)-1] // Get the current stack frame

			value := vm.loadVariable(sf.stackBase + varOffset)
			vm.cStack = append(vm.cStack, value)
		case compiler.OpLoadGlobal: // Load relative to the absolute base of the variable stack
			varOffset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc += 8
			value := vm.loadVariable(varOffset)
			vm.cStack = append(vm.cStack, value)
		case compiler.OpReturn:
			sf := vm.callStack[len(vm.callStack)-1]
			vm.pc = sf.returnAddress                           // Set pc to the return address
			vm.variableStack = vm.variableStack[:sf.stackBase] // Restore the variable stack to the base of the current function call
			vm.callStack = vm.callStack[:len(vm.callStack)-1]

			// If there is nothing on the call stack, we are returning execution from the main program
			if len(vm.callStack) == 0 {
				if len(vm.cStack) == 0 {
					return 0, nil // No value to return (void function)
				}
				return vm.pop(), nil
			}

			continue // Skip the increment of pc below, we have already set it to the return address
		case compiler.OpCall:
			fnStart := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc += 8 // Move past the call instruction

			// Push the pc of the next instruction onto the call stack
			vm.callStack = append(vm.callStack, stackFrame{
				returnAddress: vm.pc + 1,
				stackBase:     uint64(len(vm.variableStack)),
			})
			vm.pc = fnStart // Jump to the function start

			continue // Skip the increment of pc below
		case compiler.OpJump:
			offset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc = offset // Jump to the specified offset

			continue // Skip the increment of pc below
		case compiler.OpJumpIfFalse:
			condition := vm.pop()
			if condition == 0 {
				// If the condition is false, jump to the specified offset
				offset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
				vm.pc = offset
				continue // Skip the increment of pc below
			}
			// If the condition is true, just continue to the next instruction
			vm.pc += 8 // Move past the jump instruction
		case compiler.OpEq:
			a := vm.pop()
			b := vm.pop()
			if a == b {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		case compiler.OpLt:
			a := vm.pop()
			b := vm.pop()
			if b < a {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		case compiler.OpGt:
			a := vm.pop()
			b := vm.pop()
			if b > a {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		case compiler.OpLte:
			a := vm.pop()
			b := vm.pop()
			if b <= a {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		case compiler.OpGte:
			a := vm.pop()
			b := vm.pop()
			if b >= a {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		case compiler.OpAnd:
			a := vm.pop()
			b := vm.pop()
			if a != 0 && b != 0 {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		case compiler.OpOr:
			a := vm.pop()
			b := vm.pop()
			if a != 0 || b != 0 {
				vm.cStack = append(vm.cStack, 1)
			} else {
				vm.cStack = append(vm.cStack, 0)
			}
		default:
			return 0, fmt.Errorf("unknown opcode %d", op)
		}

		vm.pc++

		if vm.pc >= uint64(len(vm.program)) {
			return 0, fmt.Errorf("program counter %d is out of bounds for program length %d", vm.pc, len(vm.program))
		}
	}
}

func (vm *VM) VariableStack() []byte {
	return vm.variableStack
}

func (vm *VM) GetVariable(name string) (uint64, error) {
	for _, variable := range vm.metadata.Variables() {
		if variable.Name == name {
			return vm.loadVariable(variable.Offset), nil
		}
	}

	return 0, fmt.Errorf("exported variable %s not found", name)
}

func (vm *VM) invokeRaw(pc uint64, args ...uint64) (uint64, error) {
	vm.cStack = append(vm.cStack, args...)

	// Call the function
	vm.callStack = append(vm.callStack, stackFrame{
		returnAddress: vm.pc + 1,
		stackBase:     uint64(len(vm.variableStack)),
	})
	vm.pc = pc

	ret, err := vm.Run()
	if err != nil {
		return 0, err
	}

	return ret, nil
}

func (vm *VM) Invoke(name string, args ...any) (any, error) {
	var fn *compiler.ExportedFunction
	for _, function := range vm.metadata.Functions() {
		if function.Name == name {
			fn = &function
			break
		}
	}

	if fn == nil {
		return nil, fmt.Errorf("function %s not found", name)
	}

	if len(args) != len(fn.Args) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d", name, len(fn.Args), len(args))
	}

	// Convert args to uint64
	uint64Args := make([]uint64, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case int:
			uint64Args[i] = uint64(v)
		case uint64:
			uint64Args[i] = v
		default:
			return nil, fmt.Errorf("unsupported argument type: %T", arg)
		}
	}

	result, err := vm.invokeRaw(fn.StartOffset, uint64Args...)
	if err != nil {
		return nil, err
	}

	switch fn.ReturnType {
	case types.Void:
		return nil, nil // Void function, no return value
	case types.Int:
		return int(result), nil // Return as int
	case types.Bool:
		return result != 0, nil // Return as bool
	default:
		return nil, fmt.Errorf("unsupported return type: %s", fn.ReturnType)
	}
}
