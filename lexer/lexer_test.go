package lexer_test

import (
	"leoscript/lexer"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExpr(t *testing.T) {
	t.Run("Binary ops with whitespace", func(t *testing.T) {
		lx, err := lexer.Tokenize("1+ 2- 3 *4/ 5")
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
		lx, err := lexer.Tokenize("123+456789-987 7898 / 898989")
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

	t.Run("Invalid character", func(t *testing.T) {
		_, err := lexer.Tokenize("1+2$3")
		assert.ErrorContains(t, err, "invalid character: $")
	})

	t.Run("Parentheses", func(t *testing.T) {
		lx, err := lexer.Tokenize("((1+2)*3);")
		assert.NoError(t, err)
		assert.Equal(t, []lexer.Token{
			lexer.OpenParenToken{},
			lexer.OpenParenToken{},
			lexer.IntegerToken{Value: 1},
			lexer.BinaryToken{Operation: "+"},
			lexer.IntegerToken{Value: 2},
			lexer.CloseParenToken{},
			lexer.BinaryToken{Operation: "*"},
			lexer.IntegerToken{Value: 3},
			lexer.CloseParenToken{},
			lexer.SemicolonToken{},
		}, lx)
	})
}
