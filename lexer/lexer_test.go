package lexer_test

import (
	"leoscript/lexer"
	"leoscript/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Expression(t *testing.T) {
	t.Run("Single digit", func(t *testing.T) {
		lx, err := lexer.Tokenize("1")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Integer{Value: 1},
		}, lx)
	})

	t.Run("Single number", func(t *testing.T) {
		lx, err := lexer.Tokenize("12345")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Integer{Value: 12345},
		}, lx)
	})

	t.Run("Binary ops with whitespace", func(t *testing.T) {
		lx, err := lexer.Tokenize("1+ 2- 3 *4/ 5")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Integer{Value: 1},
			token.MathOp{Operation: "+"},
			token.Integer{Value: 2},
			token.MathOp{Operation: "-"},
			token.Integer{Value: 3},
			token.MathOp{Operation: "*"},
			token.Integer{Value: 4},
			token.MathOp{Operation: "/"},
			token.Integer{Value: 5},
		}, lx)
	})

	t.Run("Multiple digit numbers", func(t *testing.T) {
		lx, err := lexer.Tokenize("123+456789-987 7898 / 898989")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Integer{Value: 123},
			token.MathOp{Operation: "+"},
			token.Integer{Value: 456789},
			token.MathOp{Operation: "-"},
			token.Integer{Value: 987},
			token.Integer{Value: 7898},
			token.MathOp{Operation: "/"},
			token.Integer{Value: 898989},
		}, lx)
	})

	t.Run("Invalid character", func(t *testing.T) {
		_, err := lexer.Tokenize("1+2$3")
		assert.ErrorContains(t, err, "invalid character: $")
	})

	t.Run("Parentheses", func(t *testing.T) {
		lx, err := lexer.Tokenize("((1+2)*3);")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.OpenParen{},
			token.OpenParen{},
			token.Integer{Value: 1},
			token.MathOp{Operation: "+"},
			token.Integer{Value: 2},
			token.CloseParen{},
			token.MathOp{Operation: "*"},
			token.Integer{Value: 3},
			token.CloseParen{},
			token.Semicolon{},
		}, lx)
	})

}

func Test_Identifiers(t *testing.T) {
	t.Run("Identifiers", func(t *testing.T) {
		lx, err := lexer.Tokenize("foo + bar-baz")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Identifier{Value: "foo"},
			token.MathOp{Operation: "+"},
			token.Identifier{Value: "bar"},
			token.MathOp{Operation: "-"},
			token.Identifier{Value: "baz"},
		}, lx)
	})

	t.Run("Reserved keywords", func(t *testing.T) {
		lx, err := lexer.Tokenize("true false")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Boolean{Value: true},
			token.Boolean{Value: false},
		}, lx)
	})

	t.Run("Combined keywords, fine", func(t *testing.T) {
		lx, err := lexer.Tokenize("truefalse")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Identifier{Value: "truefalse"},
		}, lx)
	})

	t.Run("Mixed identifiers and keywords", func(t *testing.T) {
		lx, err := lexer.Tokenize("true foo false bar")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Boolean{Value: true},
			token.Identifier{Value: "foo"},
			token.Boolean{Value: false},
			token.Identifier{Value: "bar"},
		}, lx)
	})
}
