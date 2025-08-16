package compiler

import (
	"encoding/binary"
	"strconv"
	"strings"
)

func DebugPrint(exe []byte) string {
	b := strings.Builder{}

	i := 0
	for i < len(exe) {
		op := exe[i]
		switch op {
		case OpPush:
			b.WriteString("PUSH")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		case OpAdd:
			b.WriteString("ADD")

		case OpSub:
			b.WriteString("SUB")

		case OpMul:
			b.WriteString("MUL")

		case OpDiv:
			b.WriteString("DIV")

		case OpStore:
			b.WriteString("STORE")
			// b.WriteRune(' ')

			// b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			// i = i + 8
		case OpLoad:
			b.WriteString("LOAD")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		case OpReturn:
			b.WriteString("RETURN")

		case OpCall:
			b.WriteString("CALL")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		case OpJumpIfFalse:
			b.WriteString("JUMP_IF_FALSE")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		case OpJump:
			b.WriteString("JUMP")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		default:
			panic("unknown opcode: " + strconv.Itoa(int(op)))
		}

		b.WriteRune('\n')
		i++
	}

	return b.String()
}
