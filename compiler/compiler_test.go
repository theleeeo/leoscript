package compiler_test

import (
	"fmt"
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"strings"
	"testing"
)

func equalProgram(t *testing.T, expected, p string) {
	pOps := strings.Split(p, "\n")
	exOps := strings.Split(expected, "\n")

	if len(pOps) != len(exOps) {
		t.Errorf("Expected %d operations, got %d", len(exOps), len(pOps))
		return
	}

	for i, op := range pOps {
		if strings.TrimSpace(op) != strings.TrimSpace(exOps[i]) {
			t.Errorf("Operation %d mismatch: expected '%s', got '%s'", i, strings.TrimSpace(exOps[i]), strings.TrimSpace(op))
			return
		}
	}
}

func Test_ArithmeticExpr(t *testing.T) {
	t.Run("penis", func(t *testing.T) {
		a := -1
		b := uint(a)
		b--
		fmt.Println(int(b))
	})

	t.Run("Single integer", func(t *testing.T) {
		lx := lexer.MustTokenize("257;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			"PUSH 257\n",
		)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			`PUSH 2
			PUSH 3
			ADD
			`)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + 2 - 3 * 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			`PUSH 1
			PUSH 2
			ADD
			PUSH 3
			PUSH 4
			MUL
			SUB
			`)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + (2 - 3) + 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			`PUSH 1
			PUSH 2
			PUSH 3
			SUB
			ADD
			PUSH 4
			ADD
			`)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + (2 - 3) * 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			`PUSH 1
			PUSH 2
			PUSH 3
			SUB
			PUSH 4
			MUL
			ADD
			`)
	})
}
