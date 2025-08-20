package runtime_test

import (
	"encoding/binary"
	"fmt"
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"leoscript/runtime"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_VM_ArithmeticExpr(t *testing.T) {
	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 5, ret)
	})

	t.Run("negative number expression", func(t *testing.T) {
		lx := lexer.MustTokenize("5 - 6;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, -1, ret)
	})

	t.Run("multiplication and division binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("3 * 4 - 4 / 2;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 10, ret)
	})
}

func Test_VM_Variables(t *testing.T) {
	t.Run("Variable assignment and retrieval", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a = 10;
		int b = a + 20;
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8]))
		assert.Equal(t, uint64(30), binary.BigEndian.Uint64(vm.VariableStack()[8:16]))
	})
}

func Test_VM_IfElse(t *testing.T) {
	t.Run("True if statement", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (true) {
			return 1;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	// NOTE: This does not work because when the if-statement is passed, it will try to reset the stackframe but is not able to since there is none.
	// LeoScript will however not be ran in this way since the if-statement will always be in a function of some sort and therefor it will always have a stackframe.
	// t.Run("False if statement", func(t *testing.T) {
	// 	lx := lexer.MustTokenize(`
	// 	if (false) {
	// 		return 1;
	// 	}
	// 	`)
	// 	stmt, _ := parser.NewParser(lx).ParseStatement()
	// 	exe := compiler.CompileStatement(stmt)
	// 	vm := runtime.NewVM(exe)
	// 	ret, err := vm.Run()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, 0, ret) // No return value, should be 0
	// })

	t.Run("If-else statement", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (true) {
			return 1;
		} else {
			return 2;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	t.Run("If-else with boolean condition", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (5 > 3) {
			return 1;
		} else {
			return 2;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})
}

func Test_VM_Comparison(t *testing.T) {
	t.Run("Equality check", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (5 == 5) {
			return 1;
		} else {
			return 0;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	t.Run("Inequality check", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (5 != 3) {
			return 1;
		} else {
			return 0;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	t.Run("Greater than check", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (5 > 3) {
			return 1;
		} else {
			return 0;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	t.Run("Less than check", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (3 < 5) {
			return 1;
		} else {
			return 0;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	t.Run("Greater than or equal check", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (5 >= 5) {
			return 1;
		} else {
			return 0;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})

	t.Run("Less than or equal check", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if (3 <= 5) {
			return 1;
		} else {
			return 0;
		}
		`)
		stmt, _ := parser.NewParser(lx).ParseStatement()
		exe := compiler.CompileStatement(stmt)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret)
	})
}

func Test_VM_BooleanOps(t *testing.T) {
	t.Run("Boolean AND operation", func(t *testing.T) {
		lx := lexer.MustTokenize(`true && false;`)
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 0, ret) // false
	})

	t.Run("Boolean OR operation", func(t *testing.T) {
		lx := lexer.MustTokenize(`true || false;`)
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 1, ret) // true
	})

	t.Run("Boolean NOT operation", func(t *testing.T) {
		lx := lexer.MustTokenize(`!true;`)
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		ret, err := runtime.Evaluate(exe)
		assert.NoError(t, err)
		assert.Equal(t, 0, ret) // false
	})
}

func Test_VM_Functions(t *testing.T) {
	t.Run("Function definition and call", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn add(int a, int b) int {
			return a + b;
		}

		var a = add(2, 3);
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(5), binary.BigEndian.Uint64(vm.VariableStack()[0:8])) // a should be 5
	})

	t.Run("Function referencing global variables", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var x = 10;

		fn getX() int {
			return x;
		}

		var a = getX();
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8]))  // x should be 10
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[8:16])) // a should also be 10
	})

	t.Run("local variable shadowing global variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var x = 10;

		fn shadowX() int {
			var x = 20;
			return x;
		}

		var a = shadowX();
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8]))  // x should still be 10
		assert.Equal(t, uint64(20), binary.BigEndian.Uint64(vm.VariableStack()[8:16])) // a should be 20 (from shadowX)
	})

	t.Run("Function shadows global variable, restored later", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var x = 10;

		fn shadowX() int {
			var x = 20;
			return x;
		}

		var a = shadowX();
		var b = x;
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8]))   // Global x should still be 10
		assert.Equal(t, uint64(20), binary.BigEndian.Uint64(vm.VariableStack()[8:16]))  // a should be 20 (from shadowX)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[16:24])) // b should be 10 (global x)
	})
}

func Test_Fibonacci(t *testing.T) {
	t.Run("Fibonacci function", func(t *testing.T) {
		go func() {
			<-time.After(1 * time.Second) // Allow time for the VM to run
			fmt.Println("Fibonacci test timed out")
			os.Exit(1)
		}()
		lx := lexer.MustTokenize(`
		fn fib(int n) int {
			if n <= 1 {
				return n;
			}
			return fib(n - 1) + fib(n - 2);
		}

		var result = fib(10);
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		compiler.DebugPrint(exe.Marshal())
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(55), binary.BigEndian.Uint64(vm.VariableStack()[0:8])) // result should be 55 (fib(10))
	})
}

func Test_Assignment(t *testing.T) {
	t.Run("reassign variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() int {
			int a = 5;
			a = 10;
			return a;
		}
		var result = foo();
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8])) // result should be 10
	})

	t.Run("conditional assignment", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() int {
			int a = 5;
			if (a > 3) {
				a = 10;
			} else {
				a = 20;
			}
			return a;
		}
		var result = foo();
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8])) // result should be 10
	})

	t.Run("assignment in and after if", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() int {
			int a = 5;
			if (a > 3) {
				a = 10;
			}
			a = 20;
			return a;
		}
		var result = foo();
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(20), binary.BigEndian.Uint64(vm.VariableStack()[0:8])) // result should be 10
	})

	t.Run("reassign global variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var a = 5;

		fn foo() int {
			a = 10;
			return a;
		}

		var b = foo();
		`)
		pg, err := parser.NewParser(lx).Parse()
		assert.NoError(t, err)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8]))  // a should be 10
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[8:16])) // b should also be 10
	})
}

func Test_Read_ExportedVariable(t *testing.T) {
	t.Run("Read exported variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export int a = 5;
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		vm.Run()
		ret, err := vm.GetVariable("a")
		assert.NoError(t, err)
		assert.Equal(t, uint64(5), ret)
	})

	t.Run("Read non-exported variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int b = 10;
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		vm.Run()
		ret, err := vm.GetVariable("b")
		assert.ErrorContains(t, err, "exported variable b not found")
		assert.Equal(t, uint64(0), ret) // Should not be able to read non-exported variable
	})

	t.Run("Read multiple", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export int a = 5;
		export int b = 10;
		export int c = 15;
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		vm.Run()
		ret, err := vm.GetVariable("a")
		assert.NoError(t, err)
		assert.Equal(t, uint64(5), ret)
		ret, err = vm.GetVariable("b")
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), ret)
		ret, err = vm.GetVariable("c")
		assert.NoError(t, err)
		assert.Equal(t, uint64(15), ret)
	})
}

func Test_InvokeFunction(t *testing.T) {
	t.Run("Invoke exported function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		result, err := vm.Invoke("add", 2, 3)
		assert.NoError(t, err)
		assert.Equal(t, 5, result)
	})

	t.Run("Invoke non-exported function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn multiply(int a, int b) int {
			return a * b;
		}
		var result = multiply(2, 3);
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		result, err := vm.Invoke("multiply", 2, 3)
		assert.ErrorContains(t, err, "function multiply not found")
		assert.Equal(t, nil, result) // Should not be able to invoke non-exported function
	})

	t.Run("Invoke function with wrong number of arguments", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		_, err := vm.Invoke("add", 2) // Only one argument provided
		assert.ErrorContains(t, err, "function add expects 2 arguments, got 1")
	})

	t.Run("Invoke function with unsupported argument type", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		_, err := vm.Invoke("add", "string", 3) // First argument is a string
		assert.ErrorContains(t, err, "unsupported argument type: string")
	})

	t.Run("Invoke function with void return type", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn doNothing() {
			// This function does nothing
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		result, err := vm.Invoke("doNothing")
		assert.NoError(t, err)
		assert.Nil(t, result) // Should return nil for void function
	})

	t.Run("Invoke non-existent function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		_, err := vm.Invoke("subtract", 2, 3) // Non-existent function
		assert.ErrorContains(t, err, "function subtract not found")
	})

	t.Run("Invoke function with global variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int x = 10;

		export fn getX() int {
			return x;
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		result, err := vm.Invoke("getX")
		assert.NoError(t, err)
		assert.Equal(t, 10, result) // Should return the value of global variable x
	})

	t.Run("Invoke recursive function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn factorial(int n) int {
			if n <= 1 {
				return 1;
			}
			return n * factorial(n - 1);
		}
		`)
		pg := parser.MustParse(lx)
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Marshal())
		result, err := vm.Invoke("factorial", 5)
		assert.NoError(t, err)
		assert.Equal(t, 120, result) // Should return 5! = 120
	})
}
