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
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 123, resp.(numberVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 5, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 + 2 - 3 * 4;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, -9, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 + (2 - 3) + 4;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 4, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 + (2 - 3) * 4;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, -3, resp.(numberVal).value)
	})
}

func Test_BooleanExpr(t *testing.T) {
	t.Run("Single boolean", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true && false;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true && false || true;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true && (false || true);")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("true || false && true;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_Arithmetic_UnaryExpr(t *testing.T) {
	t.Run("Single unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("-1;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, -1, resp.(numberVal).value)
	})

	t.Run("Multiple unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("-1 + +2;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 1, resp.(numberVal).value)
	})

	t.Run("Multiple unary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("-1 + (+2);")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, 1, resp.(numberVal).value)
	})
}

func Test_Boolean_UnaryExpr(t *testing.T) {
	t.Run("Single unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("!true;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple unary expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("!true && !!false;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple unary expression with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("!true && (!(!false));")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()
		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})
}

func Test_Boolean_Comparisons(t *testing.T) {
	t.Run("Single comparison", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 == 1;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 == 1 && 2 != 1;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison with parentheses", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("false == true && 2 != 1;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison with parentheses, order changed", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("1 == 1 && 2 != 1 || 3 > 1;")
		expr, _ := parser.NewParser(lx, nil).ParseExpr()

		resp := i.evaluateExpression(expr)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_VariableDeclarations(t *testing.T) {
	t.Run("Variable declaration", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("var foo = 123;")
		stmt, _ := parser.NewParser(lx, nil).ParseStatement()

		i.evaluateStatement(stmt)

		// Check if variable is declared
		val, ok := i.activeScope.GetVar("foo")
		assert.True(t, ok)
		assert.Equal(t, 123, val.(numberVal).value)
	})

	t.Run("Variable declaration with expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("var foo = 1 + 2 * 3;")
		stmt, _ := parser.NewParser(lx, nil).ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("foo")
		assert.True(t, ok)
		assert.Equal(t, 7, val.(numberVal).value)
	})

	t.Run("Variable declaration with boolean expression", func(t *testing.T) {
		i := New()
		lx := lexer.MustTokenize("var bar = true && false || true;")
		stmt, _ := parser.NewParser(lx, nil).ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, true, val.(booleanVal).value)
	})
}

func Test_Identifiers(t *testing.T) {
	t.Run("Variable declaration with identifier", func(t *testing.T) {
		i := New()
		p := parser.NewParser(lexer.MustTokenize("var foo = 123;"), nil)
		stmt, _ := p.ParseStatement()

		i.evaluateStatement(stmt)

		scp := parser.NewScope(nil)
		scp.RegisterVar(stmt.(parser.VarDecl))

		p = parser.NewParser(lexer.MustTokenize("var bar = foo;"), scp)
		stmt, _ = p.ParseStatement()

		i.evaluateStatement(stmt)

		val, ok := i.activeScope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, 123, val.(numberVal).value)
	})

	t.Run("Variable declaration with identifier and expression", func(t *testing.T) {
		i := New()
		p := parser.NewParser(lexer.MustTokenize("var foo = 123;"), nil)
		stmt, _ := p.ParseStatement()

		i.evaluateStatement(stmt)

		scp := parser.NewScope(nil)
		scp.RegisterVar(stmt.(parser.VarDecl))

		p = parser.NewParser(lexer.MustTokenize("var bar = foo + 1;"), scp)
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
}
