package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Interpret_Expr(t *testing.T) {
	t.Run("Single integer", func(t *testing.T) {
		resp, err := RunRaw("123;")
		assert.NoError(t, err)
		assert.Equal(t, 123, resp)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		resp, err := RunRaw("2 + 3;")
		assert.NoError(t, err)
		assert.Equal(t, 5, resp)
	})

	t.Run("Multiple binary expression", func(t *testing.T) {
		resp, err := RunRaw("1 + 2 - 3 * 4;")
		assert.NoError(t, err)
		assert.Equal(t, -9, resp)
	})

	t.Run("Multiple binary expression with parentheses", func(t *testing.T) {
		resp, err := RunRaw("1 + (2 - 3) + 4;")
		assert.NoError(t, err)
		assert.Equal(t, 4, resp)
	})

	t.Run("Multiple binary expression with parentheses, order changed", func(t *testing.T) {
		resp, err := RunRaw("1 + (2 - 3) * 4;")
		assert.NoError(t, err)
		assert.Equal(t, -3, resp)
	})
}
