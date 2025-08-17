package compiler

import (
	"encoding/binary"
	"strconv"
	"strings"
)

func DumpOpcode(exe []byte) string {
	b := strings.Builder{}

	i := 0
	for i < len(exe) {
		op := exe[i]

		// b.WriteString("\033[90m") // Set text color to grey
		// b.WriteString(strconv.Itoa(i))
		// b.WriteString(": ")
		// b.WriteString("\033[0m") // Reset color

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
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		case OpLoad:
			b.WriteString("LOAD")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe[i:i+8])), 10))
			i = i + 7
		case OpLoadGlobal:
			b.WriteString("LOAD_GLOBAL")
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
		case OpJump:
			b.WriteString("JUMP")
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
		case OpEq:
			b.WriteString("EQ")
		case OpLt:
			b.WriteString("LT")
		case OpGt:
			b.WriteString("GT")
		case OpLte:
			b.WriteString("LTE")
		case OpGte:
			b.WriteString("GTE")
		case OpAnd:
			b.WriteString("AND")
		case OpOr:
			b.WriteString("OR")
		case OpStoreGlobal:
			b.WriteString("STORE_GLOBAL")
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
