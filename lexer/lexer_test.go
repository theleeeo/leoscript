package lexer_test

import (
	"leoscript/lexer"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExr(t *testing.T) {
	t.Run("Binary ops with whitespace", func(t *testing.T) {
		lx, err := lexer.Parse("1+ 2- 3 *4/ 5")
		assert.NoError(t, err)
		assert.Equal(t, []lexer.Token{
			lexer.IntegerToken{Value: 1},
			lexer.BinaryToken{Operation: "+"},
			lexer.IntegerToken{Value: 2},
			lexer.BinaryToken{Operation: "-"},
			lexer.IntegerToken{Value: 3},
			lexer.BinaryToken{Operation: "*"},
			lexer.IntegerToken{Value: 4},
			lexer.BinaryToken{Operation: "/"},
			lexer.IntegerToken{Value: 5},
		}, lx)
	})

	t.Run("Multiple digit numbers", func(t *testing.T) {
		lx, err := lexer.Parse("123+456789-987 7898 / 898989")
		assert.NoError(t, err)
		assert.Equal(t, []lexer.Token{
			lexer.IntegerToken{Value: 123},
			lexer.BinaryToken{Operation: "+"},
			lexer.IntegerToken{Value: 456789},
			lexer.BinaryToken{Operation: "-"},
			lexer.IntegerToken{Value: 987},
			lexer.IntegerToken{Value: 7898},
			lexer.BinaryToken{Operation: "/"},
			lexer.IntegerToken{Value: 898989},
		}, lx)
	})
}
