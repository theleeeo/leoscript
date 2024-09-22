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
		lx := lexer.MustTokenize("123;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, IntegerLiteral{Value: 123}, prog)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("123 + 456;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  IntegerLiteral{Value: 123},
			Right: IntegerLiteral{Value: 456},
			Op:    "+",
		}, prog)
	})

	t.Run("Multiple binary expression, no order", func(t *testing.T) {
		lx := lexer.MustTokenize("123 + 2 - 789 + 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("Multiple binary expression, order", func(t *testing.T) {
		lx := lexer.MustTokenize("123 + 2 * 789 / 4 - 9 * 1 / 2;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("Negation of integer", func(t *testing.T) {
		lx := lexer.MustTokenize("-123;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, UnaryExpression{
			Expression: IntegerLiteral{Value: 123},
			Op:         "-",
		}, prog)
	})

	t.Run("Double negation of integer", func(t *testing.T) {
		lx := lexer.MustTokenize("--123;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, UnaryExpression{
			Expression: UnaryExpression{
				Expression: IntegerLiteral{Value: 123},
				Op:         "-"},
			Op: "-",
		}, prog)
	})

	t.Run("Unneccessary plus sign", func(t *testing.T) {
		lx := lexer.MustTokenize("+123 - 45;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: UnaryExpression{
				Expression: IntegerLiteral{Value: 123},
				Op:         "+"},
			Right: IntegerLiteral{Value: 45},
			Op:    "-",
		}, prog)
	})

	t.Run("Negation of integer in operation", func(t *testing.T) {
		lx := lexer.MustTokenize("4 + -123;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		// Maybe in the future this can be made into a subtraction
		assert.EqualExportedValues(t, BinaryExpression{
			Left: IntegerLiteral{Value: 4},
			Right: UnaryExpression{
				Expression: IntegerLiteral{Value: 123},
				Op:         "-"},
			Op: "+",
		}, prog)
	})

	t.Run("Unneccessary parentheses on integer", func(t *testing.T) {
		lx := lexer.MustTokenize("(123);")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, IntegerLiteral{Value: 123}, prog)
	})

	t.Run("Parentheses first in top-level with more afterwards", func(t *testing.T) {
		lx := lexer.MustTokenize("(1 + 2) + 10;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 2},
				Op:    "+",
			},
			Right: IntegerLiteral{Value: 10},
			Op:    "+",
		}, prog)
	})

	t.Run("Parentheses first in top-level with more afterwards, order changed", func(t *testing.T) {
		lx := lexer.MustTokenize("(1 + 2) * 10;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 2},
				Op:    "+",
			},
			Right: IntegerLiteral{Value: 10},
			Op:    "*",
		}, prog)
	})

	t.Run("Parentheses on binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("(123 + 456);")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  IntegerLiteral{Value: 123},
			Right: IntegerLiteral{Value: 456},
			Op:    "+",
		}, prog)
	})

	t.Run("Parentheses inside binary expression, not changing order", func(t *testing.T) {
		lx := lexer.MustTokenize("(67 + 123) + 456 - 70;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("Parentheses inside binary expression, changing order", func(t *testing.T) {
		lx := lexer.MustTokenize("67 * (123 - 456) - 70;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("another case", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + (2 + 10) * 5;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("unary with parentheses", func(t *testing.T) {
		lx := lexer.MustTokenize("-(1 + 2);")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, UnaryExpression{
			Expression: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 2},
				Op:    "+",
			},
			Op: "-",
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
		lx := lexer.MustTokenize("true && false || true;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: BinaryExpression{
				Left:  BooleanLiteral{Value: true},
				Right: BooleanLiteral{Value: false},
				Op:    "&&",
			},
			Right: BooleanLiteral{Value: true},
			Op:    "||",
		}, prog)
	})

	t.Run("Boolean expression, changing order", func(t *testing.T) {
		lx := lexer.MustTokenize("true || false && true;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: BooleanLiteral{Value: true},
			Right: BinaryExpression{
				Left:  BooleanLiteral{Value: false},
				Right: BooleanLiteral{Value: true},
				Op:    "&&",
			},
			Op: "||",
		}, prog)
	})

	t.Run("Mixed boolean and arithmetic expression", func(t *testing.T) {
		lx := lexer.MustTokenize("true && 1 + 2 || false;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("Multiple boolean unary expressions", func(t *testing.T) {
		lx := lexer.MustTokenize("!true && !!false;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})
}

func Test_Expr_Comparisons(t *testing.T) {
	t.Run("simple equality", func(t *testing.T) {
		lx := lexer.MustTokenize("1 == 2;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  IntegerLiteral{Value: 1},
			Right: IntegerLiteral{Value: 2},
			Op:    "==",
		}, prog)
	})

	t.Run("equality with arithmetic", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + 2 == 3 * 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("equality with parentheses", func(t *testing.T) {
		lx := lexer.MustTokenize("(1 + 2) == 3 * 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("equality with parentheses, order changed", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + (2 == 3) * 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("simple comparison", func(t *testing.T) {
		lx := lexer.MustTokenize("1 < 2;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  IntegerLiteral{Value: 1},
			Right: IntegerLiteral{Value: 2},
			Op:    "<",
		}, prog)
	})

	t.Run("comparison with arithmetic", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + 2 <= 3 * 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("comparison with parentheses", func(t *testing.T) {
		lx := lexer.MustTokenize("(1 >= 2) < 3 * 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})

	t.Run("comparison with parentheses, order changed", func(t *testing.T) {
		lx := lexer.MustTokenize("1 + (2 < 3) * 4;")
		p := Parser{tokens: lx}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
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
		}, prog)
	})
}

func Test_Stmnt_VarDecl(t *testing.T) {
	t.Run("Simple integer declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = 123;")
		p := Parser{
			tokens: lx,
			scope:  NewScope(nil),
		}
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:  "a",
			Type:  types.Int,
			Value: IntegerLiteral{Value: 123},
		}, prog)
	})

	t.Run("Simple boolean declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("bool a = true;")
		p := Parser{tokens: lx, scope: NewScope(nil)}
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:  "a",
			Type:  types.Bool,
			Value: BooleanLiteral{Value: true},
		}, prog)
	})

	t.Run("Integer declaration with expression", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = 1 + 2 * 3;")
		p := Parser{tokens: lx, scope: NewScope(nil)}
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
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
		}, prog)
	})

	t.Run("Type-free var declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("var a = 1 < 2 && true;")
		p := Parser{tokens: lx, scope: NewScope(nil)}
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
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
		}, prog)
	})

	t.Run("Identifier declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = abc;")
		p := Parser{tokens: lx, scope: &Scope{varDecls: map[string]VarDecl{"abc": {Name: "a", Type: types.Int}}}}
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:  "a",
			Type:  types.Int,
			Value: Identifier{Name: "abc"},
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
		lx := lexer.MustTokenize("1 + a;")
		p := Parser{tokens: lx, scope: &Scope{varDecls: map[string]VarDecl{"a": {Name: "a", Type: types.Int}}}}
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  IntegerLiteral{Value: 1},
			Right: Identifier{Name: "a"},
			Op:    "+",
		}, prog)
	})
}

func Test_FunctionDefinitions(t *testing.T) {
	t.Run("Simple function definition", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() {}")
		p := Parser{tokens: lx}
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(new(Scope))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Args:       []Argument{},
			Body:       []Statement{},
		}, fn)
	})

	t.Run("Simple function definition with body", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() {return 1 + 2;}")
		p := Parser{tokens: lx}
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(new(Scope))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Args:       []Argument{},
			Body: []Statement{
				Return{
					Value: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
				},
			},
		}, fn)
	})

	t.Run("Function definition with return type", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() int {}")
		p := Parser{tokens: lx}
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(new(Scope))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			Args:       []Argument{},
			ReturnType: types.Int,
			Body:       []Statement{},
		}, fn)
	})

	t.Run("Function definition with one argument", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo(int a) {}")
		p := Parser{tokens: lx}
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(new(Scope))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Args: []Argument{
				{Name: "a", Type: types.Int},
			},
			Body: []Statement{},
		}, fn)
	})

	t.Run("Function definition with arguments", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo(int a, bool b, bool c) {}")
		p := Parser{tokens: lx}
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(new(Scope))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Args: []Argument{
				{Name: "a", Type: types.Int},
				{Name: "b", Type: types.Bool},
				{Name: "c", Type: types.Bool},
			},
			Body: []Statement{},
		}, fn)
	})

	t.Run("Function definition with arguments and return type", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo(bool a, int b) bool {}")
		p := Parser{tokens: lx}
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(new(Scope))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Bool,
			Args: []Argument{
				{Name: "a", Type: types.Bool},
				{Name: "b", Type: types.Int},
			},
			Body: []Statement{},
		}, fn)
	})

	t.Run("function using local scope", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn foo() {
				var a = 123;
				return 1 + a;
			}
		`)

		p := NewParser(lx, nil)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		fn, err := fnDef.parseBody(NewScope(nil))
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Args:       []Argument{},
			Body: []Statement{
				VarDecl{
					Name:  "a",
					Type:  types.Int,
					Value: IntegerLiteral{Value: 123},
				},
				Return{
					Value: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: Identifier{Name: "a"},
						Op:    "+",
					},
				},
			},
		}, fn)
	})

	t.Run("function using global and local scope", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				var b = 11;
				return a + b;
			}
		`)

		p := NewParser(lx, nil)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				VarDecl{
					Name:  "a",
					Type:  types.Int,
					Value: IntegerLiteral{Value: 10},
				},
				FnDef{
					Name:       "main",
					ReturnType: types.Void,
					Args:       []Argument{},
					Body: []Statement{
						VarDecl{
							Name:  "b",
							Type:  types.Int,
							Value: IntegerLiteral{Value: 11},
						},
						Return{
							Value: BinaryExpression{
								Left:  Identifier{Name: "a"},
								Right: Identifier{Name: "b"},
								Op:    "+",
							},
						},
					},
				},
			},
		}, prog)
	})
}

func Test_ParseFile(t *testing.T) {
	t.Run("Simple file", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn foo() int {
				bar();
				return 123;
			}

			fn bar() {
				return;
			}

			fn main() int {
				return foo() + 1;
			}
		`)

		p := Parser{tokens: lx}
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Program{
			Body: []Statement{
				FnDef{
					Name:       "foo",
					ReturnType: types.Int,
					Args:       []Argument{},
					Body: []Statement{
						Call{
							Name: "bar",
							Args: []Expression{},
						},
						Return{
							Value: IntegerLiteral{Value: 123},
						},
					},
				},
				FnDef{
					Name:       "bar",
					ReturnType: types.Void,
					Args:       []Argument{},
					Body: []Statement{
						Return{},
					},
				},
				FnDef{
					Name:       "main",
					ReturnType: types.Int,
					Args:       []Argument{},
					Body: []Statement{
						Return{
							Value: BinaryExpression{
								Left: Call{
									Name: "foo",
									Args: []Expression{},
								},
								Right: IntegerLiteral{Value: 1},
								Op:    "+",
							},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("no main function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn bar() {}
		`)
		p := Parser{tokens: lx}
		prog, err := p.ParseFile()
		assert.ErrorContains(t, err, "no main function")
		assert.Empty(t, prog)
	})
}
