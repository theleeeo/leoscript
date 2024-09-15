package parser

import (
	"leoscript/lexer"
	"leoscript/token"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Expr_Parse(t *testing.T) {
	t.Run("Single integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("123;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				IntegerLiteral{Value: 123},
			},
		}, prog)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("123 + 456;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left:  IntegerLiteral{Value: 123},
					Right: IntegerLiteral{Value: 456},
					Op:    "+",
				},
			},
		}, prog)
	})

	t.Run("Multiple binary expression, no order", func(t *testing.T) {
		lx, err := lexer.Tokenize("123 + 2 - 789 + 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 123},
							Right: IntegerLiteral{Value: 2},
							Op:    "+",
						},
						Right: IntegerLiteral{Value: 789},
						Op:    "-",
					},
					Right: IntegerLiteral{Value: 4},
					Op:    "+",
				},
			},
		}, prog)
	})

	t.Run("Multiple binary expression, order", func(t *testing.T) {
		lx, err := lexer.Tokenize("123 + 2 * 789 / 4 - 9 * 1 / 2;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left: IntegerLiteral{Value: 123},
						Right: BinaryExpression{
							Left: BinaryExpression{
								Left:  IntegerLiteral{Value: 2},
								Right: IntegerLiteral{Value: 789},
								Op:    "*",
							},
							Right: IntegerLiteral{Value: 4},
							Op:    "/",
						},
						Op: "+",
					},
					Right: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 9},
							Right: IntegerLiteral{Value: 1},
							Op:    "*",
						},
						Right: IntegerLiteral{Value: 2},
						Op:    "/",
					},
					Op: "-",
				},
			},
		}, prog)
	})

	t.Run("Negation of integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("-123;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				UnaryExpression{
					Expression: IntegerLiteral{Value: 123},
					Op:         "-",
				},
			},
		}, prog)
	})

	t.Run("Double negation of integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("--123;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				UnaryExpression{
					Expression: UnaryExpression{
						Expression: IntegerLiteral{Value: 123},
						Op:         "-"},
					Op: "-",
				},
			},
		}, prog)
	})

	t.Run("Unneccessary plus sign", func(t *testing.T) {
		lx, err := lexer.Tokenize("+123 - 45;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: UnaryExpression{
						Expression: IntegerLiteral{Value: 123},
						Op:         "+"},
					Right: IntegerLiteral{Value: 45},
					Op:    "-",
				},
			},
		}, prog)
	})

	t.Run("Negation of integer in operation", func(t *testing.T) {
		lx, err := lexer.Tokenize("4 + -123;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		// Maybe in the future this can be made into a subtraction
		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: IntegerLiteral{Value: 4},
					Right: UnaryExpression{
						Expression: IntegerLiteral{Value: 123},
						Op:         "-"},
					Op: "+",
				},
			},
		}, prog)
	})

	t.Run("Unneccessary parentheses on integer", func(t *testing.T) {
		lx, err := lexer.Tokenize("(123);")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				IntegerLiteral{Value: 123},
			},
		}, prog)
	})

	t.Run("Parentheses first in top-level with more afterwards", func(t *testing.T) {
		lx, err := lexer.Tokenize("(1 + 2) + 10;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
					Right: IntegerLiteral{Value: 10},
					Op:    "+",
				},
			},
		}, prog)
	})

	t.Run("Parentheses first in top-level with more afterwards, order changed", func(t *testing.T) {
		lx, err := lexer.Tokenize("(1 + 2) * 10;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
					Right: IntegerLiteral{Value: 10},
					Op:    "*",
				},
			},
		}, prog)
	})

	t.Run("Parentheses on binary expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("(123 + 456);")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left:  IntegerLiteral{Value: 123},
					Right: IntegerLiteral{Value: 456},
					Op:    "+",
				},
			},
		}, prog)
	})

	t.Run("Parentheses inside binary expression, not changing order", func(t *testing.T) {
		lx, err := lexer.Tokenize("(67 + 123) + 456 - 70;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 67},
							Right: IntegerLiteral{Value: 123},
							Op:    "+",
						},
						Right: IntegerLiteral{Value: 456},
						Op:    "+",
					},
					Right: IntegerLiteral{Value: 70},
					Op:    "-",
				},
			},
		}, prog)
	})

	t.Run("Parentheses inside binary expression, changing order", func(t *testing.T) {
		lx, err := lexer.Tokenize("67 * (123 - 456) - 70;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left: IntegerLiteral{Value: 67},
						Right: BinaryExpression{
							Left:  IntegerLiteral{Value: 123},
							Right: IntegerLiteral{Value: 456},
							Op:    "-",
						},
						Op: "*",
					},
					Right: IntegerLiteral{Value: 70},
					Op:    "-",
				},
			},
		}, prog)
	})

	t.Run("another case", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + (2 + 10) * 5;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: IntegerLiteral{Value: 1},
					Right: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 2},
							Right: IntegerLiteral{Value: 10},
							Op:    "+",
						},
						Right: IntegerLiteral{Value: 5},
						Op:    "*",
					},
					Op: "+",
				},
			},
		}, prog)
	})

	t.Run("unary with parentheses", func(t *testing.T) {
		lx, err := lexer.Tokenize("-(1 + 2);")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				UnaryExpression{
					Expression: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
					Op: "-",
				},
			},
		}, prog)
	})
}

func Test_Expr_PriorityMerge(t *testing.T) {
	t.Run("Merge with same priority", func(t *testing.T) {
		left := BinaryExpression{
			Left:     IntegerLiteral{Value: 1},
			Right:    IntegerLiteral{Value: 2},
			Op:       "+",
			priority: token.PRIO_SUM,
		}

		right := IntegerLiteral{Value: 3}

		newExpr := left.PriorityMerge(token.Operator{Op: "+"}, right)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  left,
			Right: right,
			Op:    "+",
		}, newExpr)
	})

	t.Run("Merge with lower priority", func(t *testing.T) {
		left := BinaryExpression{
			Left:     IntegerLiteral{Value: 1},
			Right:    IntegerLiteral{Value: 2},
			Op:       "*",
			priority: token.PRIO_SUM,
		}

		right := IntegerLiteral{Value: 3}

		newExpr := left.PriorityMerge(token.Operator{Op: "-"}, right)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:     left,
			Right:    right,
			Op:       "-",
			priority: token.PRIO_SUM,
		}, newExpr)
	})

	t.Run("Merge with higher priority", func(t *testing.T) {
		left := BinaryExpression{
			Left:     IntegerLiteral{Value: 1},
			Right:    IntegerLiteral{Value: 2},
			Op:       "+",
			priority: token.PRIO_SUM,
		}

		right := IntegerLiteral{Value: 3}

		newExpr := left.PriorityMerge(token.Operator{Op: "*"}, right)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: IntegerLiteral{Value: 1},
			Right: BinaryExpression{
				Left:  IntegerLiteral{Value: 2},
				Right: right,
				Op:    "*",
			},
			Op: "+",
		}, newExpr)
	})

	t.Run("Merge with higher priority, multiple layers", func(t *testing.T) {
		left := BinaryExpression{
			Left: IntegerLiteral{Value: 1},
			Right: BinaryExpression{
				Left: BooleanLiteral{Value: false},
				Right: BinaryExpression{
					Left:     BooleanLiteral{Value: true},
					Right:    BooleanLiteral{Value: true},
					Op:       "||",
					priority: token.PRIO_OR,
				},
				Op:       "&&",
				priority: token.PRIO_AND,
			},
			Op: "+",
		}

		right := IntegerLiteral{Value: 10}

		newExpr := left.PriorityMerge(token.Operator{Op: "*"}, right)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: IntegerLiteral{Value: 1},
			Right: BinaryExpression{
				Left: BooleanLiteral{Value: false},
				Right: BinaryExpression{
					Left: BooleanLiteral{Value: true},
					Right: BinaryExpression{
						Left:  BooleanLiteral{Value: true},
						Right: IntegerLiteral{Value: 10},
						Op:    "*",
					},
					Op:       "||",
					priority: token.PRIO_OR,
				},
				Op:       "&&",
				priority: token.PRIO_AND,
			},
			Op: "+",
		}, newExpr)
	})
}

func Test_Expr_Boolean(t *testing.T) {
	t.Run("Boolean expression, no change of order", func(t *testing.T) {
		lx, err := lexer.Tokenize("true && false || true;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  BooleanLiteral{Value: true},
						Right: BooleanLiteral{Value: false},
						Op:    "&&",
					},
					Right: BooleanLiteral{Value: true},
					Op:    "||",
				},
			},
		}, prog)
	})

	t.Run("Boolean expression, changing order", func(t *testing.T) {
		lx, err := lexer.Tokenize("true || false && true;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BooleanLiteral{Value: true},
					Right: BinaryExpression{
						Left:  BooleanLiteral{Value: false},
						Right: BooleanLiteral{Value: true},
						Op:    "&&",
					},
					Op: "||",
				},
			},
		}, prog)
	})

	t.Run("Mixed boolean and arithmetic expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("true && 1 + 2 || false;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left: BooleanLiteral{Value: true},
						Right: BinaryExpression{
							Left:  IntegerLiteral{Value: 1},
							Right: IntegerLiteral{Value: 2},
							Op:    "+",
						},
						Op: "&&",
					},
					Right: BooleanLiteral{Value: false},
					Op:    "||",
				},
			},
		}, prog)
	})

	t.Run("Multiple boolean unary expressions", func(t *testing.T) {
		lx, err := lexer.Tokenize("!true && !!false;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: UnaryExpression{
						Expression: BooleanLiteral{Value: true},
						Op:         "!",
					},
					Right: UnaryExpression{
						Expression: UnaryExpression{
							Expression: BooleanLiteral{Value: false},
							Op:         "!",
						},
						Op: "!",
					},
					Op: "&&",
				},
			},
		}, prog)
	})
}

func Test_Expr_Comparisons(t *testing.T) {
	t.Run("simple equality", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 == 2;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left:  IntegerLiteral{Value: 1},
					Right: IntegerLiteral{Value: 2},
					Op:    "==",
				},
			},
		}, prog)
	})

	t.Run("equality with arithmetic", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + 2 == 3 * 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
					Right: BinaryExpression{
						Left:  IntegerLiteral{Value: 3},
						Right: IntegerLiteral{Value: 4},
						Op:    "*",
					},
					Op: "==",
				},
			},
		}, prog)
	})

	t.Run("equality with parentheses", func(t *testing.T) {
		lx, err := lexer.Tokenize("(1 + 2) == 3 * 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
					Right: BinaryExpression{
						Left:  IntegerLiteral{Value: 3},
						Right: IntegerLiteral{Value: 4},
						Op:    "*",
					},
					Op: "==",
				},
			},
		}, prog)
	})

	t.Run("equality with parentheses, order changed", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + (2 == 3) * 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: IntegerLiteral{Value: 1},
					Right: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 2},
							Right: IntegerLiteral{Value: 3},
							Op:    "==",
						},
						Right: IntegerLiteral{Value: 4},
						Op:    "*",
					},
					Op: "+",
				},
			},
		}, prog)
	})

	t.Run("simple comparison", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 < 2;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left:  IntegerLiteral{Value: 1},
					Right: IntegerLiteral{Value: 2},
					Op:    "<",
				},
			},
		}, prog)
	})

	t.Run("comparison with arithmetic", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + 2 <= 3 * 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
					Right: BinaryExpression{
						Left:  IntegerLiteral{Value: 3},
						Right: IntegerLiteral{Value: 4},
						Op:    "*",
					},
					Op: "<=",
				},
			},
		}, prog)
	})

	t.Run("comparison with parentheses", func(t *testing.T) {
		lx, err := lexer.Tokenize("(1 >= 2) < 3 * 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    ">=",
					},
					Right: BinaryExpression{
						Left:  IntegerLiteral{Value: 3},
						Right: IntegerLiteral{Value: 4},
						Op:    "*",
					},
					Op: "<",
				},
			},
		}, prog)
	})

	t.Run("comparison with parentheses, order changed", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + (2 < 3) * 4;")
		assert.NoError(t, err)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left: IntegerLiteral{Value: 1},
					Right: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 2},
							Right: IntegerLiteral{Value: 3},
							Op:    "<",
						},
						Right: IntegerLiteral{Value: 4},
						Op:    "*",
					},
					Op: "+",
				},
			},
		}, prog)
	})
}

func Test_Stmnt_VarDecl(t *testing.T) {
	t.Run("Simple integer declaration", func(t *testing.T) {
		lx, err := lexer.Tokenize("int a = 123;")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				VarDecl{
					Name:  "a",
					Type:  types.Int,
					Value: IntegerLiteral{Value: 123},
				},
			},
		}, prog)
	})

	t.Run("Simple boolean declaration", func(t *testing.T) {
		lx, err := lexer.Tokenize("bool a = true;")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				VarDecl{
					Name:  "a",
					Type:  types.Bool,
					Value: BooleanLiteral{Value: true},
				},
			},
		}, prog)
	})

	t.Run("Integer declaration with expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("int a = 1 + 2 * 3;")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				VarDecl{
					Name: "a",
					Type: types.Int,
					Value: BinaryExpression{
						Left: IntegerLiteral{Value: 1},
						Right: BinaryExpression{
							Left:  IntegerLiteral{Value: 2},
							Right: IntegerLiteral{Value: 3},
							Op:    "*",
						},
						Op: "+",
					},
				},
			},
		}, prog)
	})

	t.Run("Type-free var declaration", func(t *testing.T) {
		lx, err := lexer.Tokenize("var a = 1 < 2 && true;")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				VarDecl{
					Name: "a",
					Type: types.Bool,
					Value: BinaryExpression{
						Left: BinaryExpression{
							Left:  IntegerLiteral{Value: 1},
							Right: IntegerLiteral{Value: 2},
							Op:    "<",
						},
						Right: BooleanLiteral{Value: true},
						Op:    "&&",
					},
				},
			},
		}, prog)
	})
}

func Test_ReturnTypes(t *testing.T) {
	t.Run("Literals", func(t *testing.T) {
		assert.Equal(t, types.Int, IntegerLiteral{Value: 123}.ReturnType())
		assert.Equal(t, types.Bool, BooleanLiteral{Value: true}.ReturnType())
	})

	t.Run("Unary expressions", func(t *testing.T) {
		assert.Equal(t, types.Int, UnaryExpression{
			Expression: IntegerLiteral{Value: 123},
			Op:         "-",
		}.ReturnType())

		assert.Equal(t, types.Bool, UnaryExpression{
			Expression: BooleanLiteral{Value: true},
			Op:         "!",
		}.ReturnType())
	})

	t.Run("Binary expressions", func(t *testing.T) {
		assert.Equal(t, types.Int, BinaryExpression{
			Left:  IntegerLiteral{Value: 1},
			Right: IntegerLiteral{Value: 2},
			Op:    "+",
		}.ReturnType())

		assert.Equal(t, types.Bool, BinaryExpression{
			Left:  BooleanLiteral{Value: true},
			Right: BooleanLiteral{Value: false},
			Op:    "&&",
		}.ReturnType())
	})

	t.Run("Binary expressions with different types", func(t *testing.T) {
		assert.Equal(t, types.Bool, BinaryExpression{
			Left:  IntegerLiteral{Value: 1},
			Right: IntegerLiteral{Value: 2},
			Op:    "<",
		}.ReturnType())

		assert.Equal(t, types.Bool, BinaryExpression{
			Left:  BooleanLiteral{Value: true},
			Right: IntegerLiteral{Value: 2},
			Op:    "==",
		}.ReturnType())
	})
}

func Test_Identifiers(t *testing.T) {
	t.Run("Simple identifier in binary expr", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 + a;")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				BinaryExpression{
					Left:  IntegerLiteral{Value: 1},
					Right: Identifier{Name: "a"},
					Op:    "+",
				},
			},
		}, prog)
	})
}

func Test_FunctionDefinitions(t *testing.T) {
	t.Run("Simple function definition", func(t *testing.T) {
		lx, err := lexer.Tokenize("fn foo() {}")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				FnDef{
					Name: "foo",
					Args: nil,
					Body: []Statement{},
				},
			},
		}, prog)
	})

	t.Run("Simple function definition with body", func(t *testing.T) {
		lx, err := lexer.Tokenize("fn foo() {return 1 + 2;}")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				FnDef{
					Name: "foo",
					Args: nil,
					Body: []Statement{
						Return{
							Value: BinaryExpression{
								Left:  IntegerLiteral{Value: 1},
								Right: IntegerLiteral{Value: 2},
								Op:    "+",
							},
						},
					},
				},
			},
		}, prog)
	})

	// t.Run("Function definition with arguments", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo(a, b, c) {}")
	// 	assert.NoError(t, err)

	// 	prog, err := NewParser(lx).Parse()
	// 	assert.NoError(t, err)

	// 	assert.EqualExportedValues(t, Program{
	// 		Body: []Statement{
	// 			FnDef{
	// 				Name: "foo",
	// 				Args: []Argument{
	// 					{Name: "a", Type: types.Any},
	// 					{Name: "b", Type: types.Any},
	// 					{Name: "c", Type: types.Any},
	// 				},
	// 				Body: []Statement{},
	// 			},
	// 		},
	// 	}, prog)
	// })

	// t.Run("Function definition with return type", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo() int {}")
	// 	assert.NoError(t, err)

	// 	prog, err := NewParser(lx).Parse()
	// 	assert.NoError(t, err)

	// 	assert.EqualExportedValues(t, Program{
	// 		Body: []Statement{
	// 			FnDef{
	// 				Name:       "foo",
	// 				Args:       nil,
	// 				ReturnType: types.Int,
	// 				Body:       []Statement{},
	// 			},
	// 		},
	// 	}, prog)
	// })

	// t.Run("Function definition with arguments and return type", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo(a, b, c) bool {}")
	// 	assert.NoError(t, err)

	// 	prog, err := NewParser(lx).Parse()
	// 	assert.NoError(t, err)

	// 	assert.EqualExportedValues(t, Program{
	// 		Body: []Statement{
	// 			FnDef{
	// 				Name: "foo",
	// 				Args: []Argument{
	// 					{Name: "a", Type: types.Any},
	// 					{Name: "b", Type: types.Any},
	// 					{Name: "c", Type: types.Any},
	// 				},
	// 				ReturnType: types.Bool,
	// 				Body:       []Statement{},
	// 			},
	// 		},
	// 	}, prog)
	// })
}
