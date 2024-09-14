package runtime

import (
	"leoscript/lexer"
	"leoscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Interpret_Expr(t *testing.T) {
	tokens, err := lexer.Tokenize("1 + 2 * 3;")
	assert.NoError(t, err)

	program, err := parser.NewParser(tokens).Parse()
	assert.NoError(t, err)

	result := Run(program)
	assert.Equal(t, 7, result)
}
