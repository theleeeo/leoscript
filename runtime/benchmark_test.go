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
		export fn main() int {
			return 1 + 2 * 3 - 4 / 2;
		}`)
		expr, _ := parser.NewParser(lx).Parse()
		exe := compiler.Compile(expr)
		vm, _ := NewVM(exe.Marshal())
		vm.Init()

		for b.Loop() {
			vm.Reset()        // Reset VM state before each iteration
			vm.Invoke("main") // Invoke the main function
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
		pg := parser.MustParse(lx)
		i, err := NewInterpreter(pg).Initialize()
		if err != nil {
			b.Fatalf("Interpreter init error: %v", err)
		}

		for b.Loop() {
			i.Invoke("fib", 20)
		}
	})

	b.Run("bytecode VM", func(b *testing.B) {
		lx := lexer.MustTokenize(`
		export fn fib(int n) int {
			if n <= 1 {
				return n;
			}
			return fib(n - 1) + fib(n - 2);
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm, _ := NewVM(exe.Marshal())

		for b.Loop() {
			vm.Reset() // Reset VM state after each iteration
			ret, err := vm.Invoke("fib", 20)
			if err != nil {
				b.Fatalf("VM invoke error: %v", err)
			}
			if ret != 6765 {
				b.Fatalf("Expected 6765, got %d", ret)
			}
		}
	})

	b.Run("native", func(b *testing.B) {
		var fibonacci func(n int) int
		fibonacci = func(n int) int {
			if n <= 1 {
				return n
			}
			return fibonacci(n-1) + fibonacci(n-2)
		}

		for b.Loop() {
			fibonacci(20)
		}
	})
}
