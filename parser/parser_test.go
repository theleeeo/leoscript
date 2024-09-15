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

func Test_Stmnt_VarDef(t *testing.T) {
	t.Run("Simple integer definition", func(t *testing.T) {
		lx, err := lexer.Tokenize("int a = 123;")
		assert.NoError(t, err)

		prog, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				VarDef{
					Name:  "a",
					Type:  types.Int,
					Value: IntegerLiteral{Value: 123},
				},
			},
		}, prog)
	})
}
