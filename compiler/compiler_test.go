package compiler_test

import (
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"strings"
	"testing"
)

func equalProgram(t *testing.T, p, expected string) {
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
	t.Run("Single integer", func(t *testing.T) {
		lx := lexer.MustTokenize("257;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr, nil)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			"PUSH 257\n",
		)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := compiler.CompileStatement(expr, nil)
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
		i := compiler.CompileStatement(expr, nil)
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
		i := compiler.CompileStatement(expr, nil)
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
		i := compiler.CompileStatement(expr, nil)
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

func Test_StackVariables(t *testing.T) {
	t.Run("store a simple integer variable", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = 10;")
		stmt, _ := parser.NewParser(lx).ParseStatement()
		i := compiler.CompileStatement(stmt, nil)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			`PUSH 10
			STORE
			`,
		)
	})

	t.Run("store integer result of expression", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = 10 + 20;")
		stmt, _ := parser.NewParser(lx).ParseStatement()
		i := compiler.CompileStatement(stmt, nil)
		s := compiler.DebugPrint(i)
		equalProgram(t,
			s,
			`PUSH 10
			PUSH 20
			ADD
			STORE
			`,
		)
	})

	t.Run("save in var and use it", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a = 10;
		int b = a + 20;
		`)
		pg, _ := parser.NewParser(lx).Parse()
		i := compiler.Compile(pg)
		s := compiler.DebugPrint(i.Raw())
		equalProgram(t,
			s,
			`PUSH 10
			STORE
			LOAD 0
			PUSH 20
			ADD
			STORE
			`,
		)
	})
}
