package runtime

import (
	"encoding/binary"
	"fmt"
	"leoscript/compiler"
)

type VM struct {
	cStack        []uint64
	variableStack []byte
	callStack     []stackFrame

	program []byte
	pc      uint64
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
	return &VM{
		program:       program,
		pc:            0,
		cStack:        make([]uint64, 0),
		variableStack: make([]byte, 0),
	}
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

func (vm *VM) Run() (int, error) {
	for vm.pc < uint64(len(vm.program)) {
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
			// If there is nothing on the call stack, we are returning execution from the main program
			if len(vm.callStack) == 0 {
				if len(vm.cStack) == 0 {
					return 0, nil // No value to return
				}
				return int(vm.pop()), nil
			}

			sf := vm.callStack[len(vm.callStack)-1]
			vm.pc = sf.returnAddress                           // Set pc to the return address
			vm.variableStack = vm.variableStack[:sf.stackBase] // Restore the variable stack to the base of the current function call
			vm.callStack = vm.callStack[:len(vm.callStack)-1]

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

	}

	if len(vm.cStack) == 0 {
		return 0, nil
	}

	// TODO: Should not be encountered here. The program should always end with a return.
	// Or maybe, if we allow it to run incomplete-programs, just raw bytecode, it should be possible. But that should be handled explicitly either way.
	return int(vm.pop()), nil
}

func (vm *VM) VariableStack() []byte {
	return vm.variableStack
}
