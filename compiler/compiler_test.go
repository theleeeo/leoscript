package compiler_test

import (
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"strings"
	"testing"
)

func equalProgram(t *testing.T, p []byte, expected string) {
	ps := compiler.DebugPrint(p)

	if ps == "" {
		t.Error("Compiled program is empty")
		return
	}

	ps = strings.TrimSpace(ps)
	pOps := strings.Split(ps, "\n")

	expected = strings.TrimSpace(expected)
	exOps := strings.Split(expected, "\n")

	if len(pOps) != len(exOps) {
		t.Errorf("Expected %d operations, got %d\nExpected operations: %v\nGot operations: %v", len(exOps), len(pOps), exOps, pOps)
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
		equalProgram(t,
			compiler.CompileStatement(expr),
			"PUSH 257\n",
		)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		equalProgram(t,
			compiler.CompileStatement(expr),
			`PUSH 2
			PUSH 3
			ADD
			`)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + 2 - 3 * 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		equalProgram(t,
			compiler.CompileStatement(expr),
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
		equalProgram(t,
			compiler.CompileStatement(expr),
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
		equalProgram(t,
			compiler.CompileStatement(expr),
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
		equalProgram(t,
			compiler.CompileStatement(stmt),
			`PUSH 10
			STORE
			`,
		)
	})

	t.Run("store integer result of expression", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = 10 + 20;")
		stmt, _ := parser.NewParser(lx).ParseStatement()
		equalProgram(t,
			compiler.CompileStatement(stmt),
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
		equalProgram(t,
			compiler.Compile(pg).Raw(),
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

func Test_Function(t *testing.T) {
	t.Run("simple function definition", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`LOAD 0
			LOAD 8
			ADD
			RETURN
			`,
		)
	})

	// TODO: Parser should add void return
	// t.Run("void return", func(t *testing.T) {
	// 	lx := lexer.MustTokenize(`
	// 	fn foo() {}
	// 	`)
	// 	pg, _ := parser.NewParser(lx).Parse()
	// 	equalProgram(t,
	// 		compiler.Compile(pg).Raw(),
	// 		`RETURN`,
	// 	)
	// })
}
