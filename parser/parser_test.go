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

	t.Run("Negation of integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("-123;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.UnaryExpression{
					Expression: parser.IntegerLiteral{Value: 123},
					Op:         "-",
				},
			},
		}, prog)
	})

	t.Run("Double negation of integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("--123;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.UnaryExpression{
					Expression: parser.UnaryExpression{
						Expression: parser.IntegerLiteral{Value: 123},
						Op:         "-"},
					Op: "-",
				},
			},
		}, prog)
	})

	t.Run("Unneccessary plus sign", func(t *testing.T) {
		lx, err := lexer.Tokenize("+123 - 45;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.UnaryExpression{
						Expression: parser.IntegerLiteral{Value: 123},
						Op:         "+"},
					Right: parser.IntegerLiteral{Value: 45},
					Op:    "-",
				},
			},
		}, prog)
	})

	t.Run("Negation of integer in operation", func(t *testing.T) {
		lx, err := lexer.Tokenize("4 + -123;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		// Maybe in the future this can be made into a subtraction
		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.IntegerLiteral{Value: 4},
					Right: parser.UnaryExpression{
						Expression: parser.IntegerLiteral{Value: 123},
						Op:         "-"},
					Op: "+",
				},
			},
		}, prog)
	})

	t.Run("Unneccessary parentheses on integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("(123);")
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

	t.Run("Parentheses first in top-level with more afterwards", func(t *testing.T) {
		lx, err := lexer.Tokenize("(1 + 2) + 10;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.BinaryExpression{
						Left:     parser.IntegerLiteral{Value: 1},
						Right:    parser.IntegerLiteral{Value: 2},
						Op:       "+",
						Priority: 100,
					},
					Right:    parser.IntegerLiteral{Value: 10},
					Op:       "+",
					Priority: 0,
				},
			},
		}, prog)
	})

	t.Run("Parentheses first in top-level with more afterwards, order changed", func(t *testing.T) {
		lx, err := lexer.Tokenize("(1 + 2) * 10;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.BinaryExpression{
						Left:     parser.IntegerLiteral{Value: 1},
						Right:    parser.IntegerLiteral{Value: 2},
						Op:       "+",
						Priority: 100,
					},
					Right:    parser.IntegerLiteral{Value: 10},
					Op:       "*",
					Priority: 1,
				},
			},
		}, prog)
	})

	t.Run("Parentheses on binary expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("(123 + 456);")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left:     parser.IntegerLiteral{Value: 123},
					Right:    parser.IntegerLiteral{Value: 456},
					Op:       "+",
					Priority: 100,
				},
			},
		}, prog)
	})

	t.Run("Parentheses inside binary expression, not changing order", func(t *testing.T) {
		lx, err := lexer.Tokenize("(67 + 123) + 456 - 70;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.BinaryExpression{
						Left: parser.BinaryExpression{
							Left:     parser.IntegerLiteral{Value: 67},
							Right:    parser.IntegerLiteral{Value: 123},
							Op:       "+",
							Priority: 100,
						},
						Right: parser.IntegerLiteral{Value: 456},
						Op:    "+",
					},
					Right: parser.IntegerLiteral{Value: 70},
					Op:    "-",
				},
			},
		}, prog)
	})

	t.Run("Parentheses inside binary expression, changing order", func(t *testing.T) {
		lx, err := lexer.Tokenize("67 * (123 - 456) - 70;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.BinaryExpression{
						Left: parser.IntegerLiteral{Value: 67},
						Right: parser.BinaryExpression{
							Left:     parser.IntegerLiteral{Value: 123},
							Right:    parser.IntegerLiteral{Value: 456},
							Op:       "-",
							Priority: 100,
						},
						Op:       "*",
						Priority: 1,
					},
					Right: parser.IntegerLiteral{Value: 70},
					Op:    "-",
				},
			},
		}, prog)
	})

	t.Run("another case", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + (2 + 10) * 5;")
		assert.NoError(t, err)

		p := parser.NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.Equal(t, parser.Program{
			Body: []parser.Expression{
				parser.BinaryExpression{
					Left: parser.IntegerLiteral{Value: 1},
					Right: parser.BinaryExpression{
						Left: parser.BinaryExpression{
							Left:     parser.IntegerLiteral{Value: 2},
							Right:    parser.IntegerLiteral{Value: 10},
							Op:       "+",
							Priority: 100,
						},
						Right:    parser.IntegerLiteral{Value: 5},
						Op:       "*",
						Priority: 1,
					},
					Op: "+",
				},
			},
		}, prog)
	})
}
