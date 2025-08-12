package runtime

import (
	"encoding/binary"
	"fmt"
	"leoscript/compiler"
)

type VM struct {
	stack []uint64

	program []byte
	pc      int
}

func NewVM(program []byte) *VM {
	return &VM{
		program: program,
		pc:      0,
		stack:   make([]uint64, 0),
	}
}

func (vm *VM) Run() (int, error) {
	for vm.pc < len(vm.program) {
		op := vm.program[vm.pc]

		switch op {
		case compiler.OpPush:
			if vm.pc+8 > len(vm.program) {
				return 0, fmt.Errorf("not enough bytes for PUSH operation")
			}
			vm.pc++

			value := binary.BigEndian.Uint64(vm.program[vm.pc : vm.pc+8])
			vm.stack = append(vm.stack, value)
			vm.pc += 7
		case compiler.OpAdd:
			if len(vm.stack) < 2 {
				return 0, fmt.Errorf("not enough values on stack for ADD operation")
			}

			// Extract the values from the stack
			a := vm.stack[len(vm.stack)-1]
			b := vm.stack[len(vm.stack)-2]
			// Actually remove them from the stack
			vm.stack = vm.stack[:len(vm.stack)-2]
			// Add their sum back to the stack
			vm.stack = append(vm.stack, a+b)
		case compiler.OpSub:
			if len(vm.stack) < 2 {
				return 0, fmt.Errorf("not enough values on stack for SUB operation")
			}
			a := vm.stack[len(vm.stack)-1]
			b := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]
			vm.stack = append(vm.stack, b-a)

		case compiler.OpMul:
			if len(vm.stack) < 2 {
				return 0, fmt.Errorf("not enough values on stack for SUB operation")
			}
			a := vm.stack[len(vm.stack)-1]
			b := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]
			vm.stack = append(vm.stack, b*a)
		case compiler.OpDiv:
			if len(vm.stack) < 2 {
				return 0, fmt.Errorf("not enough values on stack for SUB operation")
			}
			a := vm.stack[len(vm.stack)-1]
			b := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]
			vm.stack = append(vm.stack, b/a)

		default:
			return 0, fmt.Errorf("unknown opcode %d", op)
		}

		vm.pc++
	}

	return int(vm.stack[len(vm.stack)-1]), nil
}
