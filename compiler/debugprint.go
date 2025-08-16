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
			i = i + 8
		case OpAdd:
			b.WriteString("ADD")
			i++
		case OpSub:
			b.WriteString("SUB")
			i++
		case OpMul:
			b.WriteString("MUL")
			i++
		case OpDiv:
			b.WriteString("DIV")
			i++
		case OpStore:
			b.WriteString("STORE")
			// b.WriteRune(' ')
			i++

			// b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			// i = i + 8
		case OpLoad:
			b.WriteString("LOAD")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 8
		case OpReturn:
			b.WriteString("RETURN")
			i++
		default:
			panic("unknown opcode: " + strconv.Itoa(int(op)))
		}

		b.WriteRune('\n')
	}

	return b.String()
}
