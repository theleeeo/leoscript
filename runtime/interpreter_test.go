package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ArithmeticExpr(t *testing.T) {
	t.Run("Single integer", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("123;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 123, resp.(numberVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("2 + 3;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 5, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("1 + 2 - 3 * 4;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, -9, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("1 + (2 - 3) + 4;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 4, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("1 + (2 - 3) * 4;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, -3, resp.(numberVal).value)
	})
}

func Test_BooleanExpr(t *testing.T) {
	t.Run("Single boolean", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("true;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("true && false;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("true && false || true;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("true && (false || true);")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("true || false && true;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_Arithmetic_UnaryExpr(t *testing.T) {
	t.Run("Single unary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("-1;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, -1, resp.(numberVal).value)
	})

	t.Run("Multiple unary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("-1 + +2;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.(numberVal).value)
	})

	t.Run("Multiple unary expression with parentheses", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("-1 + (+2);")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.(numberVal).value)
	})
}

func Test_Boolean_UnaryExpr(t *testing.T) {
	t.Run("Single unary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("!true;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple unary expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("!true && !!false;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple unary expression with parentheses", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("!true && (!(!false));")
		assert.NoError(t, err)
		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, false, resp.(booleanVal).value)
	})
}

func Test_Boolean_Comparisons(t *testing.T) {
	t.Run("Single comparison", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("1 == 1;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("1 == 1 && 2 != 1;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison with parentheses", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("false == true && 2 != 1;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple comparison with parentheses, order changed", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("1 == 1 && 2 != 1 || 3 > 1;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}

func Test_VariableDeclarations(t *testing.T) {
	t.Run("Variable declaration", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("var foo = 123;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 123, resp.(numberVal).value)

		// Check if variable is declared
		val, ok := i.scope.GetVar("foo")
		assert.True(t, ok)
		assert.Equal(t, 123, val.(numberVal).value)
	})

	t.Run("Variable declaration with expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("var foo = 1 + 2 * 3;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 7, resp.(numberVal).value)

		val, ok := i.scope.GetVar("foo")
		assert.True(t, ok)
		assert.Equal(t, 7, val.(numberVal).value)
	})

	t.Run("Variable declaration with boolean expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("var bar = true && false || true;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)

		val, ok := i.scope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, true, val.(booleanVal).value)
	})
}

func Test_Identifiers(t *testing.T) {
	t.Run("Variable declaration with identifier", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("var foo = 123; var bar = foo;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 123, resp.(numberVal).value)

		val, ok := i.scope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, 123, val.(numberVal).value)
	})

	t.Run("Variable declaration with identifier and expression", func(t *testing.T) {
		i := New()
		err := i.LoadRaw("var foo = 123; var bar = foo + 1;")
		assert.NoError(t, err)

		resp, err := i.Run()
		assert.NoError(t, err)
		assert.Equal(t, 124, resp.(numberVal).value)

		val, ok := i.scope.GetVar("bar")
		assert.True(t, ok)
		assert.Equal(t, 124, val.(numberVal).value)
	})
}
