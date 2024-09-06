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
			{Type: lexer.Integer, Literal: "1"},
			{Type: lexer.Binary, Literal: "+"},
			{Type: lexer.Integer, Literal: "2"},
			{Type: lexer.Binary, Literal: "-"},
			{Type: lexer.Integer, Literal: "3"},
			{Type: lexer.Binary, Literal: "*"},
			{Type: lexer.Integer, Literal: "4"},
			{Type: lexer.Binary, Literal: "/"},
			{Type: lexer.Integer, Literal: "5"},
		}, lx)
	})
}
