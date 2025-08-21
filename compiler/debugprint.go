package compiler

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func DebugPrint(rawExe []byte) {
	exe := new(Executable)
	if err := exe.Unmarshal(rawExe); err != nil {
		panic(fmt.Sprintf("unmarshalling executable: %v", err))
	}

	md := exe.Metadata()
	fnDefs := md.functions

	// Add a dummy function definition for initialization part of the program
	fnDefs = append(fnDefs, ExportedFunction{
		Name:        "_init",
		StartOffset: 0,
	})

	b := strings.Builder{}

	i := 0
	for i < len(exe.code) {
		startIdx := i

		op := exe.code[i]

		b.WriteString("\033[90m") // Set text color to grey
		b.WriteString(strconv.Itoa(i))
		b.WriteString(": ")
		b.WriteString("\033[0m") // Reset color

		switch op {
		case OpPush:
			b.WriteString("PUSH")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
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

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
			i = i + 7
		case OpLoad:
			b.WriteString("LOAD")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
			i = i + 7
		case OpLoadGlobal:
			b.WriteString("LOAD_GLOBAL")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
			i = i + 7
		case OpReturn:
			b.WriteString("RETURN")

		case OpCall:
			b.WriteString("CALL")
			b.WriteRune(' ')
			i++

			fnOffset := binary.BigEndian.Uint64(exe.code[i : i+8])
			b.WriteString(strconv.FormatInt(int64(fnOffset), 10))
			i = i + 7

			// Find the function definition by its offset
			fnIndex := slices.IndexFunc(fnDefs, func(fn ExportedFunction) bool {
				return fn.StartOffset == fnOffset && !fn.stub
			})

			fnDef := fnDefs[fnIndex]
			b.WriteString(" \033[1;34m") // Set text color to blue
			b.WriteRune('(')
			b.WriteString(fnDef.Name)
			b.WriteRune(')')
			b.WriteString("\033[0m") // Reset color
		case OpJump:
			b.WriteString("JUMP")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
			i = i + 7
		case OpJumpIfFalse:
			b.WriteString("JUMP_IF_FALSE")
			b.WriteRune(' ')
			i++

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
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

			b.WriteString(strconv.FormatInt(int64(binary.BigEndian.Uint64(exe.code[i:i+8])), 10))
			i = i + 7
		case OpInvokeStub:
			b.WriteString("INVOKE_STUB")
			b.WriteRune(' ')
			i++

			stubNumber := binary.BigEndian.Uint64(exe.code[i : i+8])
			b.WriteString(strconv.FormatInt(int64(stubNumber), 10))
			i = i + 7

			// Find the stub definition by its number
			stubDef := md.Stubs()[stubNumber]
			b.WriteString(" \033[1;34m") // Set text color to blue
			b.WriteRune('(')
			b.WriteString(stubDef.Name)
			b.WriteRune(')')
			b.WriteString("\033[0m") // Reset color
		default:
			panic("unknown opcode: " + strconv.Itoa(int(op)))
		}

		fnIndex := slices.IndexFunc(fnDefs, func(fn ExportedFunction) bool {
			return fn.StartOffset == uint64(startIdx) && !fn.stub
		})
		if fnIndex != -1 {
			fnDef := fnDefs[fnIndex]
			b.WriteString(" \033[1;34m") // Set text color to blue
			b.WriteString(":")
			b.WriteString(fnDef.Name)
			b.WriteString("\033[0m") // Reset color
		}

		b.WriteRune('\n')
		i++
	}

	fmt.Println(b.String())
}

func DumpOpcode(exe []byte) string {
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
		case OpInvokeStub:
			b.WriteString("INVOKE_STUB")
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
