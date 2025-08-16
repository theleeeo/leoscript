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

	// Compile the initialization code required for the program
	for _, stmt := range p.VarDecls {
		c.code = append(c.code, c.compileStatement(stmt, sc)...)
	}
	// The initialization code is followed by a return to have the VM break out of its execution, allowing it to start acception external invokations.
	c.code = append(c.code, OpReturn) // Add a return at the end of the program

	for _, stmt := range p.FnDefs {
		c.functions[stmt.Name] = uint64(len(c.code))
		c.code = append(c.code, c.compileStatement(stmt, sc)...)
	}

	for _, fn := range c.unresolvedFunctions {
		start, ok := c.functions[fn.name]
		if !ok {
			panic(fmt.Sprintf("unresolved function %s not found in functions map", fn.name))
		}
		// Replace the placeholder offset with the actual start instruction of the function
		binary.BigEndian.PutUint64(c.code[fn.placeholderOffset:], start)
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

	unresolvedFunctions []unresolvedFunction
}

// unresolvedFunction is used to track a function call that has not yet been compiled and therefore does not have a start instruction.
type unresolvedFunction struct {
	name              string
	placeholderOffset uint64 // The offset in the code where the function call is made
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
		case "!":
			expr := c.compileStatement(stmt.Expression, sc)
			bytes = append(bytes, expr...)
			bytes = append(bytes, OpPush)
			bytes = binary.BigEndian.AppendUint64(bytes, 0) // Push false
			bytes = append(bytes, OpEq)                     // Negate the result of equality
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
		case "==":
			bytes = append(bytes, OpEq)
		case "<":
			bytes = append(bytes, OpLt)
		case ">":
			bytes = append(bytes, OpGt)
		case "<=":
			bytes = append(bytes, OpLte)
		case ">=":
			bytes = append(bytes, OpGte)
		case "!=":
			// For !=, we can use the equality operator and negate the result
			bytes = append(bytes, OpEq)
			bytes = append(bytes, OpPush)
			bytes = binary.BigEndian.AppendUint64(bytes, 0) // Push false
			bytes = append(bytes, OpEq)                     // Negate the result of equality
		case "&&": // TODO: Short-circuit evaluation
			bytes = append(bytes, OpAnd)
		case "||": // TODO: Short-circuit evaluation
			bytes = append(bytes, OpOr)
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
		v, ok := sc.stackAllocs[stmt.Name]
		if !ok {
			panic(fmt.Sprintf("variable %s not found in scope", stmt.Name)) // Should not happen on a correct ast
		}

		bytes = append(bytes, OpLoad)
		// bytes = binary.BigEndian.AppendUint64(bytes, v.size)
		bytes = binary.BigEndian.AppendUint64(bytes, v.stackOffset)
	case parser.FnDef:
		if stmt.Stub {
			panic("stubs not implemented")
		}

		// TODO: Functions should likely have to first store their parameters in the stack
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
	case parser.Call:
		// Compile the function call arguments
		for _, arg := range stmt.Args {
			argBytes := c.compileStatement(arg, sc)
			bytes = append(bytes, argBytes...)
		}

		bytes = append(bytes, OpCall)

		// Call the function
		start, ok := c.functions[stmt.Name]
		if ok {
			bytes = binary.BigEndian.AppendUint64(bytes, start)
		} else {
			// If the function is not defined yet, we need to add a placeholder for it
			c.unresolvedFunctions = append(c.unresolvedFunctions, unresolvedFunction{
				name:              stmt.Name,
				placeholderOffset: uint64(len(c.code) + len(bytes)),
			})
			bytes = binary.BigEndian.AppendUint64(bytes, 0) // Placeholder for the function call
		}
	case parser.If:
		// Compile the condition
		cond := c.compileStatement(stmt.Cond, sc)
		bytes = append(bytes, cond...)

		// Compile the true branch
		var trueBranch []byte
		for _, trueStmt := range stmt.Then {
			trueBranch = append(trueBranch, c.compileStatement(trueStmt, sc)...)
		}

		var elseBranch []byte
		for _, elseStmt := range stmt.Else {
			elseBranch = append(elseBranch, c.compileStatement(elseStmt, sc)...)
		}

		bytes = append(bytes, OpJumpIfFalse)

		// The offset to jump to if the condition is false
		var falseOffset uint64
		if len(elseBranch) > 0 {
			// If there is an else branch, we need to jump to it if the condition is false
			falseOffset = uint64(len(c.code) + len(bytes) + len(trueBranch) + 8) // +8 for the size of the jump location itself
			bytes = binary.BigEndian.AppendUint64(bytes, falseOffset)
		} else {
			// If there is no else branch, we just jump over the true branch
			falseOffset = uint64(len(c.code) + len(bytes) + len(trueBranch) + 8) // +8 for the size of the jump location itself
			bytes = binary.BigEndian.AppendUint64(bytes, falseOffset)
		}

		bytes = append(bytes, trueBranch...)

		if len(elseBranch) > 0 {
			// If there is an else branch, we need to jump over it after the true branch
			bytes = append(bytes, OpJump)
			// The offset to jump to after the true branch
			bytes = binary.BigEndian.AppendUint64(bytes, uint64(len(c.code)+len(bytes)+len(elseBranch)+8)) // +8 for the size of the jump location itself
		}

		bytes = append(bytes, elseBranch...)
	case parser.BooleanLiteral:
		bytes = append(bytes, OpPush)
		if stmt.Value {
			bytes = binary.BigEndian.AppendUint64(bytes, 1) // true
		} else {
			bytes = binary.BigEndian.AppendUint64(bytes, 0) // false
		}
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
	OpCall
	OpJump
	OpJumpIfFalse
	OpEq
	OpLt
	OpGt
	OpLte
	OpGte
	OpAnd
	OpOr
)
