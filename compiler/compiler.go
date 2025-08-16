package compiler

import (
	"encoding/binary"
	"fmt"
	"leoscript/parser"
)

func Compile(p *parser.Program) Executable {
	c := &compiler{
		code:      make([]byte, 0),
		functions: make(map[string]uint64),
	}

	sc := &scopeContext{
		stackAllocs:        make(map[string]variable),
		currentStackOffset: 0,
	}

	for _, stmt := range p.VarDecls {
		c.code = append(c.code, c.compileStatement(stmt, sc)...)
	}

	for _, stmt := range p.FnDefs {
		c.code = append(c.code, c.compileStatement(stmt, sc)...)
		c.functions[stmt.Name] = uint64(len(c.code))
	}

	return Executable{
		code: c.code,
	}
}

type Executable struct {
	code []byte
}

func (e Executable) Raw() []byte {
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

type compiler struct {
	code []byte

	// A mapping for the start-instruction of each function
	functions map[string]uint64
}

func CompileStatement(stmt parser.Statement) []byte {
	c := &compiler{
		code:      make([]byte, 0),
		functions: make(map[string]uint64),
	}
	c.code = append(c.code, c.compileStatement(stmt, &scopeContext{
		stackAllocs:        make(map[string]variable),
		currentStackOffset: 0,
	})...)
	return c.code
}

func (c *compiler) compileStatement(stmt parser.Statement, sc *scopeContext) []byte {
	var bytes []byte // TODO: Possible optimization to pre-allocate the size of bytes based on the expected size of the type of stmt

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
			expr := c.compileStatement(stmt.Expression, sc)
			bytes = append(bytes, expr...)

			bytes = append(bytes, OpSub)
		default:
			panic("unsupported unary operator: " + stmt.Op)
		}
	case parser.BinaryExpression:
		left := c.compileStatement(stmt.Left, sc)
		right := c.compileStatement(stmt.Right, sc)

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
		sc.stackAllocs[stmt.Name] = variable{
			stackOffset: sc.currentStackOffset,
			size:        stmt.Type.Size(),
		}
		sc.currentStackOffset += stmt.Type.Size()

		// The code required to compute the value of the variable.
		val := c.compileStatement(stmt.Value, sc)

		bytes = append(bytes, val...)
		bytes = append(bytes, OpStore)
		// bytes = binary.BigEndian.AppendUint64(bytes, varSize)
	case parser.VarIdentifier:
		v := sc.stackAllocs[stmt.Name]

		bytes = append(bytes, OpLoad)
		// bytes = binary.BigEndian.AppendUint64(bytes, v.size)
		bytes = binary.BigEndian.AppendUint64(bytes, v.stackOffset)
	case parser.FnDef:
		if stmt.Stub {
			panic("stubs not implemented")
		}

		for _, arg := range stmt.Args {
			// Allocate space for each parameter in the stack
			sc.stackAllocs[arg.Name] = variable{
				stackOffset: sc.currentStackOffset,
				size:        arg.Type.Size(),
			}
			sc.currentStackOffset += arg.Type.Size()
		}

		// Compile the function body
		for _, bodyStmt := range stmt.Body {
			bytes = append(bytes, c.compileStatement(bodyStmt, sc)...) // TODO: Nested scopeContexts
		}
	case parser.Return:
		if stmt.Value != nil {
			val := c.compileStatement(stmt.Value, sc)
			bytes = append(bytes, val...)
		}
		bytes = append(bytes, OpReturn)
	case parser.VoidLiteral:
		// No operation needed for void literals
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
	OpReturn
)
