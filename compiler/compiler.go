package compiler

import (
	"encoding/binary"
	"fmt"
	"leoscript/parser"
)

func Compile(p *parser.Program) Executable {
	var code []byte

	sc := &scopeContext{
		stackAllocs:        make(map[string]variable),
		currentStackOffset: 0,
	}

	for _, stmt := range p.VarDecls {
		code = append(code, CompileStatement(stmt, sc)...)
	}

	for _, stmt := range p.FnDefs {
		code = append(code, CompileStatement(stmt, sc)...)
	}

	return Executable{
		code: code,
	}
}

type Executable struct {
	code []byte
}

func (e *Executable) Raw() []byte {
	return e.code
}

type scopeContext struct {
	// The stack offsets for each variable allocated on the stack
	stackAllocs map[string]variable
	// The current stack offset for the next variable to be allocated
	currentStackOffset uint64
}

type variable struct {
	stackOffset uint64
	size        uint64
}

func CompileStatement(stmt parser.Statement, sc *scopeContext) []byte {
	if sc == nil {
		sc = &scopeContext{
			stackAllocs:        make(map[string]variable),
			currentStackOffset: 0,
		}
	}

	var bytes []byte

	switch stmt := stmt.(type) {
	case parser.IntegerLiteral:
		bytes = append(bytes, OpPush)
		bytes = binary.BigEndian.AppendUint64(bytes, uint64(stmt.Value))
	case parser.UnaryExpression:
		switch stmt.Op {
		case "-":
			bytes = append(bytes, OpPush)
			bytes = binary.BigEndian.AppendUint64(bytes, 0)

			bytes = append(bytes, OpPush)
			expr := CompileStatement(stmt.Expression, sc)
			bytes = append(bytes, expr...)

			bytes = append(bytes, OpSub)
		default:
			panic("unsupported unary operator: " + stmt.Op)
		}
	case parser.BinaryExpression:
		left := CompileStatement(stmt.Left, sc)
		right := CompileStatement(stmt.Right, sc)

		bytes = append(bytes, left...)
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
	case parser.VarDecl:
		// The full size required by the variable
		varSize := stmt.Type.Size()
		sc.stackAllocs[stmt.Name] = variable{
			stackOffset: sc.currentStackOffset,
			size:        varSize,
		}
		sc.currentStackOffset += varSize

		// The code required to compute the value of the variable.
		val := CompileStatement(stmt.Value, sc)

		bytes = append(bytes, val...)
		bytes = append(bytes, OpStore)
		// bytes = binary.BigEndian.AppendUint64(bytes, varSize)
	case parser.VarIdentifier:
		v := sc.stackAllocs[stmt.Name]

		bytes = append(bytes, OpLoad)
		// bytes = binary.BigEndian.AppendUint64(bytes, v.size)
		bytes = binary.BigEndian.AppendUint64(bytes, v.stackOffset)
	default:
		panic(fmt.Sprintf("unsupported statement type: %T", stmt))
	}

	return bytes
}

const (
	OpPush byte = iota + 1
	OpMul
	OpAdd
	OpSub
	OpDiv
	OpStore
	OpLoad
)
