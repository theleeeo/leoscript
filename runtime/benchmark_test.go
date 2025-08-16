package runtime

import (
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"testing"
)

func Benchmark_Arithmetic(b *testing.B) {
	b.Run("ast-walking interpreter", func(b *testing.B) {
		lx := lexer.MustTokenize(`
			1 + 2 * 3 - 4 / 2;
		`)
		expr, _ := parser.NewParser(lx).ParseExpr()
		i := NewInterpreter(nil)

		for b.Loop() {
			i.evaluateExpression(expr)
		}
	})

	b.Run("bytecode VM", func(b *testing.B) {
		lx := lexer.MustTokenize(`
		1 + 2 * 3 - 4 / 2;
		`)
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		vm := NewVM(exe)

		for b.Loop() {
			vm.Run()
		}
	})
}

func Benchmark_Fibonacci(b *testing.B) {
	b.Run("ast-walking interpreter", func(b *testing.B) {

		lx := lexer.MustTokenize(`
		fn fib(int n) int {
			if n <= 1 {
				return n;
			}
			return fib(n - 1) + fib(n - 2);
		}`)
		pg, err := parser.NewParser(lx).Parse()
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
		i, err := NewInterpreter(pg).Initialize()
		if err != nil {
			b.Fatalf("Interpreter init error: %v", err)
		}

		for b.Loop() {
			i.Invoke("fib", 20)
		}
	})
}
