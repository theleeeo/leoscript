package runtime

import (
	"encoding/binary"
	"fmt"
	"leoscript/compiler"
)

type VM struct {
	cStack        []uint64
	variableStack []byte

	program []byte
	pc      int
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

func (vm *VM) Run() (int, error) {
	for vm.pc < len(vm.program) {
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
			vm.cStack = append(vm.cStack, a+b)
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
			value := vm.pop()
			// Store the value in the variable stack
			vm.variableStack = binary.BigEndian.AppendUint64(vm.variableStack, value)
		case compiler.OpLoad:
			varOffset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			value := binary.BigEndian.Uint64(vm.variableStack[varOffset : varOffset+8])
			vm.cStack = append(vm.cStack, value)
			vm.pc += 8
		case compiler.OpReturn:
			if len(vm.cStack) == 0 {
				return 0, nil // No value to return
			}
			return int(vm.pop()), nil
		case compiler.OpJump:
			offset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
			vm.pc = int(offset) // Jump to the specified offset
			continue            // Skip the increment of pc below
		case compiler.OpJumpIfFalse:
			condition := vm.pop()
			if condition == 0 {
				// If the condition is false, jump to the specified offset
				offset := binary.BigEndian.Uint64(vm.program[vm.pc+1 : vm.pc+1+8])
				vm.pc = int(offset)
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
		case compiler.OpGte:
			a := vm.pop()
			b := vm.pop()
			if b >= a {
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
