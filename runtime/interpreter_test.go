package runtime

import (
	"leoscript/lexer"
	"leoscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ArithmeticExpr(t *testing.T) {
	t.Run("Single integer", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("123;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 123, resp.(numberVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 5, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 + 2 - 3 * 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, -9, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 + (2 - 3) + 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 4, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 + (2 - 3) * 4;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, -3, resp.(numberVal).value)
	})
}

func Test_BooleanExpr(t *testing.T) {
	t.Run("Single boolean", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true && false;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true && false || true;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true && (false || true);")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true || false && true;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_Arithmetic_UnaryExpr(t *testing.T) {
	t.Run("Single unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("-1;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, -1, resp.(numberVal).value)
	})

	t.Run("Multiple unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("-1 + +2;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 1, resp.(numberVal).value)
	})

	t.Run("Multiple unary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("-1 + (+2);")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 1, resp.(numberVal).value)
	})
}

func Test_Boolean_UnaryExpr(t *testing.T) {
	t.Run("Single unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("!true;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("!true && !!false;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple unary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("!true && (!(!false));")
		expr, _ := parser.NewParser(lx).ParseExpr()
		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})
}

func Test_Boolean_Comparisons(t *testing.T) {
	t.Run("Single comparison", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 == 1;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 == 1 && 2 != 1;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("false == true && 2 != 1;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison with parentheses, order changed", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 == 1 && 2 != 1 || 3 > 1;")
		expr, _ := parser.NewParser(lx).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_VariableDeclarations(t *testing.T) {
	t.Run("Variable declaration", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("var foo = 123;")
		stmt, _ := parser.NewParser(lx).ParseStatement()

		i.evaluateStatement(stmt)

		// Check if variable is declared
		val, ok := i.activeScope.GetVar("foo")
		assert.True(t, ok)
		assert.Equal(t, 123, val.(numberVal).value)
	})

	t.Run("Variable declaration with expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("var foo = 1 + 2 * 3;")
		stmt, _ := parser.NewParser(lx).ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("foo")
		assert.True(t, ok)
		assert.Equal(t, 7, val.(numberVal).value)
	})

	t.Run("Variable declaration with boolean expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("var bar = true && false || true;")
		stmt, _ := parser.NewParser(lx).ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, true, val.(booleanVal).value)
	})
}

func Test_Identifiers(t *testing.T) {
	t.Run("Variable declaration with identifier", func(t *testing.T) {
		i := New()
		p := parser.NewParser(lexer.MustTokenize("var foo = 123;"))
		stmt, _ := p.ParseStatement()

		i.evaluateStatement(stmt)

		p = parser.NewParser(lexer.MustTokenize("var bar = foo;"))
		stmt, _ = p.ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, 123, val.(numberVal).value)
	})

	t.Run("Variable declaration with identifier and expression", func(t *testing.T) {
		i := New()
		p := parser.NewParser(lexer.MustTokenize("var foo = 123;"))
		stmt, _ := p.ParseStatement()

		i.evaluateStatement(stmt)

		p = parser.NewParser(lexer.MustTokenize("var bar = foo + 1;"))
		stmt, _ = p.ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, 124, val.(numberVal).value)
	})
}

func Test_RunCompleteFile(t *testing.T) {
	t.Run("Simple main function", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				return 1;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.(numberVal).value)
	})

	t.Run("Functioncall", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				return foo() + 1;
			}

			fn foo() int {
				return 1 + 2;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 4, resp.(numberVal).value)
	})

	t.Run("variable declaration", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				var a = 1;
				return a + 10;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 11, resp.(numberVal).value)
	})
}

func Test_Scope(t *testing.T) {
	t.Run("global and local scope", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			var a = 10;

			fn main() {
				var b = 11;
				return a + b;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 21, resp.(numberVal).value)
	})

	t.Run("local overrides global scope", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			var a = 10;

			fn main() {
				var a = 11;
				return a;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 11, resp.(numberVal).value)
	})

	t.Run("nested scopes", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			var a = 10;

			fn main() {
				var b = 11;
				return foo();
			}

			fn foo() int {
				var c = 12;
				return a + b + c;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 33, resp.(numberVal).value)
	})
}

func Test_IfStatements(t *testing.T) {
	t.Run("constant if, true", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				if true {
					return 1;
				}
				return 0;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.(numberVal).value)
	})

	t.Run("constant if, false", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				if false {
					return 1;
				}
				return 0;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.(numberVal).value)
	})

	t.Run("Simple if else statement", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				if false {
					return 1;
				} else {
					return 0;
				}
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.(numberVal).value)
	})

	t.Run("Simple if else if else statement", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				if false {
					return 1;
				} else if true {
					return 0;
				} else {
					return -1;
				}
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.(numberVal).value)
	})
}

func Test_VariableAssignment(t *testing.T) {
	t.Run("Simple variable assignment", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				var a = 1;
				a = a + 1;
				return a;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.(numberVal).value)
	})

	t.Run("Variable assignment with expression", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				var a = 1;
				var b = 2;
				a = a + b;
				return a;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 3, resp.(numberVal).value)
	})

	t.Run("Assign variable to another variable", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				var a = 1;
				var b = 2;
				b = a;
				return b;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.(numberVal).value)
	})
}

func Test_WhileStatements(t *testing.T) {
	t.Run("Simple while loop", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				var a = 0;
				while a < 5 {
					a = a + 1;
				}
				return a;
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 5, resp.(numberVal).value)
	})

	// t.Run("While loop with break", func(t *testing.T) {
	// 	i := New()

	// 	err := i.LoadRaw(`
	// 		fn main() int {
	// 			var a = 0;
	// 			while (a < 10) {
	// 				if (a == 5) {
	// 					break;
	// 				}
	// 				a = a + 1;
	// 			}
	// 			return a;
	// 		}
	// 	`)
	// 	assert.NoError(t, err)

	// 	resp, err := i.Run()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, 5, resp.(numberVal).value)
	// })
}

func Test_FunctionCall(t *testing.T) {
	t.Run("Unknown function call", func(t *testing.T) {
		i := New()

		err := i.LoadRaw(`
			fn main() int {
				return foo();
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.ErrorContains(t, err, "function foo not defined")
		assert.Nil(t, resp)
	})
}

func Test_Stub(t *testing.T) {
	t.Run("Call registered stub function", func(t *testing.T) {
		i := New()
		i.RegisterStub("foo", func() int {
			return 42
		})

		err := i.LoadRaw(`
			stub foo() int;
			
			fn main() int {
				return foo();
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 42, resp.(numberVal).value)
	})

	t.Run("Call registered stub function with arguments", func(t *testing.T) {
		i := New()
		i.RegisterStub("foo", func(arg int) int {
			return arg + 1
		})

		err := i.LoadRaw(`
				stub foo(int x) int;

				fn main() int {
					return foo(41);
				}
			`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 42, resp.(numberVal).value)
	})

	t.Run("Call registered stub function with multiple arguments", func(t *testing.T) {
		i := New()
		i.RegisterStub("foo", func(a1, a2 int) int {
			return a1 + a2
		})

		err := i.LoadRaw(`
			stub foo(int x, int y) int;

			fn main() int {
				return foo(41, 1);
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 42, resp.(numberVal).value)
	})

	t.Run("Boolean stub function", func(t *testing.T) {
		i := New()
		i.RegisterStub("isEven", func(n int) bool {
			return n%2 == 0
		})

		err := i.LoadRaw(`
			stub isEven(int x) bool;

			fn main() bool {
				return isEven(42);
			}
		`)
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_StubVerification(t *testing.T) {
	t.Run("Verify stub with correct signature", func(t *testing.T) {
		i := New()
		i.RegisterStub("foo", func(arg int) int {
			return arg + 1
		})

		err := i.LoadRaw(`
			stub foo(int x) int;

			fn main() int {
				return foo(41);
			}
		`)
		assert.NoError(t, err)

		err = i.verifyStubs()
		assert.NoError(t, err)
	})

	t.Run("Verify stub with incorrect argument type", func(t *testing.T) {
		i := New()
		i.RegisterStub("foo", func(arg bool) int {
			return 1
		})

		err := i.LoadRaw(`
			stub foo(int x) int;

			fn main() int {
				return foo(41);
			}
		`)
		assert.NoError(t, err)

		err = i.verifyStubs()
		assert.ErrorContains(t, err, "argument 1: expected bool, got Int")
	})

	t.Run("Verify stub with incorrect return type", func(t *testing.T) {
		i := New()
		i.RegisterStub("foo", func(arg int) bool {
			return arg%2 == 0
		})

		err := i.LoadRaw(`
			stub foo(int x) int;

			fn main() int {
				return foo(42);
			}
		`)
		assert.NoError(t, err)

		err = i.verifyStubs()
		assert.ErrorContains(t, err, "return value: expected bool, got Int")
	})

	t.Run("register non-function stub", func(t *testing.T) {
		i := New()

		defer func() {
			if r := recover(); r != nil {
				assert.Contains(t, r, "expected a function, got string")
			} else {
				t.Errorf("expected panic but did not occur")
			}
		}()

		i.RegisterStub("foo", "not a function")
	})
}
