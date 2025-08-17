package compiler_test

import (
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"strings"
	"testing"
)

func equalProgram(t *testing.T, p []byte, expected string) {
	ps := compiler.DumpOpcode(p)

	if ps == "" {
		t.Error("Compiled program is empty")
		return
	}

	ps = strings.TrimSpace(ps)
	pOpsRaw := strings.Split(ps, "\n")
	pOps := make([]string, len(pOpsRaw))
	for i, op := range pOpsRaw {
		pOps[i] = strings.TrimSpace(op)
	}

	expected = strings.TrimSpace(expected)
	exOpsRaw := strings.Split(expected, "\n")
	exOps := make([]string, len(exOpsRaw))
	for i, op := range exOpsRaw {
		exOps[i] = strings.TrimSpace(op)
	}

	if len(pOps) != len(exOps) {
		t.Errorf("Expected %d operations, got %d\nExpected operations:\n%v\nGot operations:\n%v",
			len(exOps), len(pOps), exOps, pOps)
		return
	}

	for i := range pOps {
		if pOps[i] != exOps[i] {
			t.Errorf("Operation %d mismatch: expected '%s', got '%s'\nExpected:\n%v\nGot:\n%v",
				i, exOps[i], pOps[i], exOps, pOps)
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
			STORE 0
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
			STORE 0
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
			`
			PUSH 10
			STORE_GLOBAL 0
			LOAD_GLOBAL 0
			PUSH 20
			ADD
			STORE_GLOBAL 8
			RETURN
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
			`
			RETURN
			STORE 0
			STORE 8
			LOAD 0
			LOAD 8
			ADD
			RETURN
			`,
		)
	})

	t.Run("void return", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			RETURN
			RETURN`,
		)
	})

	t.Run("call function with no arguments", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {
			var a = 5;
		}

		fn main() {
			foo();
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			RETURN
			PUSH 5
			STORE 0
			RETURN
			CALL 1
			RETURN
			`,
		)
	})

	t.Run("call function with arguments", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn add(int a, int b) int {
			return a + b;
		}

		fn main() {
			add(5, 10);
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			RETURN
			STORE 0
			STORE 8
			LOAD 0
			LOAD 8
			ADD
			RETURN
			PUSH 5
			PUSH 10
			CALL 1
			RETURN
			`,
		)
	})

	t.Run("call function before its compilation", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a = foo();

		fn foo() int {
			return 42;
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			CALL 19
			STORE_GLOBAL 0
			RETURN
			PUSH 42
			RETURN
			`,
		)
	})
}

func Test_IfElse(t *testing.T) {
	t.Run("if statement with condition", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (true) {
			int a = 5;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		equalProgram(t,
			compiler.CompileStatement(stmt),
			`
			PUSH 1
			JUMP_IF_FALSE 36
			PUSH 5
			STORE 0
			`,
		)
	})

	t.Run("if-else statement", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (true) {
			int a = 5;
		} else {
			int a = 10;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		equalProgram(t,
			compiler.CompileStatement(stmt),
			`
			PUSH 1
			JUMP_IF_FALSE 36
			PUSH 5
			STORE 0
			JUMP 63
			PUSH 10
			STORE 0
			`,
		)
	})
}

func Test_Assignment(t *testing.T) {
	t.Run("reassign variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {
			int a = 5;
			a = 10;
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			RETURN
			PUSH 5
			STORE 0
			PUSH 10
			STORE 0
			RETURN
			`,
		)
	})

	t.Run("conditional assignment", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {
			int a = 5;
			if (a > 3) {
				a = 10;
			} else {
				a = 20;
			}
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			RETURN
			PUSH 5
			STORE 0
			LOAD 0
			PUSH 3
			GT
			JUMP_IF_FALSE 65
			PUSH 10
			STORE 0
			JUMP 92
			PUSH 20
			STORE 0
			RETURN
			`,
		)
	})

	t.Run("assignment in and after if", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {
			int a = 5;
			if (a > 3) {
				a = 10;
			}
			a = 20;
		}
		`)
		pg, _ := parser.NewParser(lx).Parse()
		equalProgram(t,
			compiler.Compile(pg).Raw(),
			`
			RETURN
			PUSH 5
			STORE 0
			LOAD 0
			PUSH 3
			GT
			JUMP_IF_FALSE 65
			PUSH 10
			STORE 0
			PUSH 20
			STORE 0
			RETURN
			`,
		)
	})
}
