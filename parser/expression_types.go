package parser

import (
	"fmt"
	"leoscript/token"
	"leoscript/types"
	"strconv"
	"strings"
)

type Expression interface {
	ReturnType() types.Type
	String() string
}

type IntegerLiteral struct {
	Value int
}

func (IntegerLiteral) ReturnType() types.Type { return types.Int }

func (i IntegerLiteral) String() string {
	return strconv.Itoa(i.Value)
}

type VoidLiteral struct{}

func (VoidLiteral) ReturnType() types.Type { return types.Void }

func (VoidLiteral) String() string {
	return "void"
}

type BooleanLiteral struct {
	Value bool
}

func (BooleanLiteral) ReturnType() types.Type { return types.Bool }

func (b BooleanLiteral) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Op       string
	priority token.Priority
}

// Note: Will not support other number types than int with the current setup.
func (e BinaryExpression) ReturnType() types.Type {
	if e.Op == "&&" || e.Op == "||" {
		return types.Bool
	}

	if e.Op == "<" || e.Op == ">" || e.Op == "<=" || e.Op == ">=" {
		return types.Bool
	}

	if e.Op == "==" || e.Op == "!=" {
		return types.Bool
	}

	if e.Op == "+" || e.Op == "-" || e.Op == "*" || e.Op == "/" {
		return types.Int
	}

	panic("unknown binary expression type")
}

// PriorityMerge will merge the current binary expression with a new expression based on the priorities of the operators
// A new expression tree will be returned with the order of operations handled correctly.
func (root BinaryExpression) PriorityMerge(binTk token.Operator, newExpr Expression) Expression {
	priority := binTk.Priority()

	// If the new priority is lower, it should be higher in the expression tree to be evaluated later.
	// If it is the same, it should also be higher to preserve left-to-right evaluation.
	if priority <= root.priority {
		// No priority swap needed, create a new root expression
		return BinaryExpression{
			Left:     root,
			Right:    newExpr,
			Op:       binTk.Op,
			priority: priority,
		}
	}

	// The right side of the root binary expression is also a binary expression
	// We need to also do a priority merge on that to support multiple layers of priority
	if rbin, ok := root.Right.(BinaryExpression); ok {
		root.Right = rbin.PriorityMerge(binTk, newExpr)
		return root
	}

	// We have a higher priority operator, so we need to swap the root right side
	newRight := BinaryExpression{
		Left:     root.Right,
		Right:    newExpr,
		Op:       binTk.Op,
		priority: priority,
	}
	root.Right = newRight

	return root
}

func (e BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left.String(), e.Op, e.Right.String())
}

type UnaryExpression struct {
	Expression Expression
	Op         string
}

func (e UnaryExpression) ReturnType() types.Type { return e.Expression.ReturnType() }

func (e UnaryExpression) String() string {
	return e.Op + e.Expression.String()
}

type VarIdentifier struct {
	Name       string
	returnType types.Type
}

func (i VarIdentifier) ReturnType() types.Type {
	if i.returnType == nil {
		return types.Unspecified
	}

	return i.returnType
}

func (i VarIdentifier) String() string {
	return i.Name
}

type Call struct {
	Name       string
	Args       []Expression
	returnType types.Type
}

func (c Call) ReturnType() types.Type {
	if c.returnType == nil {
		return types.Unspecified
	}

	return c.returnType
}

func (c Call) String() string {
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", c.Name, strings.Join(args, ", "))
}

type StringLiteral struct {
	Value string
}

func (StringLiteral) ReturnType() types.Type { return types.String }

func (s StringLiteral) String() string {
	return fmt.Sprintf("%q", s.Value)
}

type StructLiteral struct {
	Type   types.Type
	Fields []FieldLiteral
}

type FieldLiteral struct {
	Name  string
	Value Expression
}

func (s StructLiteral) ReturnType() types.Type {
	return s.Type
}

func (s StructLiteral) String() string {
	var fields []string
	for _, field := range s.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s", field.Name, field.Value.String()))
	}
	return fmt.Sprintf("struct %v { %s }", s.Type, strings.Join(fields, ", "))
}
