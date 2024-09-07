package parser_test

import (
	"leoscript/lexer"
	"leoscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseExpr(t *testing.T) {
	t.Run("Single integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("123;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.IntegerLiteral{Value: 123},
			},
		}, prog)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("123 + 456;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left:  parser.IntegerLiteral{Value: 123},
					Right: parser.IntegerLiteral{Value: 456},
					Op:    "+",
				},
			},
		}, prog)
	})

	t.Run("Multiple binary expression, no order", func(t *testing.T) {
		lx, err := lexer.Tokenize("123 + 2 - 789 + 4;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.BinaryExpression{
						Left: parser.BinaryExpression{
							Left:  parser.IntegerLiteral{Value: 123},
							Right: parser.IntegerLiteral{Value: 2},
							Op:    "+",
						},
						Right: parser.IntegerLiteral{Value: 789},
						Op:    "-",
					},
					Right: parser.IntegerLiteral{Value: 4},
					Op:    "+",
				},
			},
		}, prog)
	})

	t.Run("Multiple binary expression, order", func(t *testing.T) {
		lx, err := lexer.Tokenize("123 + 2 * 789 / 4 - 9 * 1 / 2;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.BinaryExpression{
						Left: parser.IntegerLiteral{Value: 123},
						Right: parser.BinaryExpression{
							Left: parser.BinaryExpression{
								Left:     parser.IntegerLiteral{Value: 2},
								Right:    parser.IntegerLiteral{Value: 789},
								Op:       "*",
								Priority: 1,
							},
							Right:    parser.IntegerLiteral{Value: 4},
							Op:       "/",
							Priority: 1,
						},
						Op: "+",
					},
					Right: parser.BinaryExpression{
						Left: parser.BinaryExpression{
							Left:     parser.IntegerLiteral{Value: 9},
							Right:    parser.IntegerLiteral{Value: 1},
							Op:       "*",
							Priority: 1,
						},
						Right:    parser.IntegerLiteral{Value: 2},
						Op:       "/",
						Priority: 1,
					},
					Op: "-",
				},
			},
		}, prog)
	})
}

// func Test_ParseExpr_InvalidSyntax(t *testing.T) {
// t.Run("Multiple integers", func(t *testing.T) {
// 	lx, err := lexer.Tokenize("123 456 789")
// 	assert.NoError(t, err)

// 	p := parser.NewParser(lx)
// 	prog, err := p.Parse()
// 	assert.NoError(t, err)

// 	assert.Equal(t, parser.Program{
// 		Body: []parser.Expression{
// 			parser.IntegerLiteral{Value: 123},
// 			parser.IntegerLiteral{Value: 456},
// 			parser.IntegerLiteral{Value: 789},
// 		},
// 	}, prog)
// })
// }
