package compiler

import (
	"encoding/binary"
	"leoscript/parser"
	"strconv"
	"strings"
)

func Compile(p *parser.Program) Executable {
	var code []byte

	for _, stmt := range p.FnDefs {
		code = append(code, CompileStatement(stmt)...)
	}

	return Executable{
		code: code,
	}
}

type Executable struct {
	code []byte
}

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
		default:
			panic("unknown opcode" + strconv.Itoa(int(op)))
		}

		b.WriteRune('\n')
	}

	return b.String()
}

func CompileStatement(stmt parser.Statement) []byte {
	switch stmt := stmt.(type) {
	case parser.IntegerLiteral:
		// return []byte{OpPush, byte(stmt.Value)} // Example bytecode for integer literal
		bytes := make([]byte, 1, 1+8)
		bytes[0] = OpPush
		bytes = binary.BigEndian.AppendUint64(bytes, uint64(stmt.Value))
		return bytes
	case parser.UnaryExpression:
		switch stmt.Op {
		case "-":
			var bytes []byte
			bytes = append(bytes, OpPush)
			bytes = binary.BigEndian.AppendUint64(bytes, 0)

			bytes = append(bytes, OpPush)
			expr := CompileStatement(stmt.Expression)
			bytes = append(bytes, expr...)

			bytes = append(bytes, OpSub)
			return bytes
		default:
			panic("unsupported unary operator: " + stmt.Op)
		}
	case parser.BinaryExpression:
		bytes := make([]byte, 0, 1+8+1+8)

		left := CompileStatement(stmt.Left)
		bytes = append(bytes, left...)

		right := CompileStatement(stmt.Right)
		bytes = append(bytes, right...)
		switch stmt.Op {
		case "+":
			bytes = append(bytes, OpAdd)
		case "-":
			bytes = append(bytes, OpSub)
		case "*":
			bytes = append(bytes, OpMul)
		case "/":
			bytes = append(bytes, OpDiv)
		default:
			panic("unsupported binary operator: " + stmt.Op)
		}
		return bytes
	default:
		panic("unsupported statement type")
	}
}

const (
	OpPush byte = iota + 1
	OpMul
	OpAdd
	OpSub
	OpDiv
)
