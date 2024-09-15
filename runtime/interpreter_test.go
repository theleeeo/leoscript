package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ArithmeticExpr(t *testing.T) {
	t.Run("Single integer", func(t *testing.T) {
		resp, err := RunRaw("123;")
		assert.NoError(t, err)
		assert.Equal(t, 123, resp.(numberVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		resp, err := RunRaw("2 + 3;")
		assert.NoError(t, err)
		assert.Equal(t, 5, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		resp, err := RunRaw("1 + 2 - 3 * 4;")
		assert.NoError(t, err)
		assert.Equal(t, -9, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		resp, err := RunRaw("1 + (2 - 3) + 4;")
		assert.NoError(t, err)
		assert.Equal(t, 4, resp.(numberVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		resp, err := RunRaw("1 + (2 - 3) * 4;")
		assert.NoError(t, err)
		assert.Equal(t, -3, resp.(numberVal).value)
	})
}

func Test_BooleanExpr(t *testing.T) {
	t.Run("Single boolean", func(t *testing.T) {
		resp, err := RunRaw("true;")
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		resp, err := RunRaw("true && false;")
		assert.NoError(t, err)
		assert.Equal(t, false, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		resp, err := RunRaw("true && false || true;")
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		resp, err := RunRaw("true && (false || true);")
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		resp, err := RunRaw("true || false && true;")
		assert.NoError(t, err)
		assert.Equal(t, true, resp.(booleanVal).value)
	})
}
