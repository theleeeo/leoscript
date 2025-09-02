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
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, IntegerLiteral{Value: 123}, prog)
	})

	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("123 + 456;")
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, UnaryExpression{
			Expression: IntegerLiteral{Value: 123},
			Op:         "-",
		}, prog)
	})

	t.Run("Double negation of integer", func(t *testing.T) {
		lx := lexer.MustTokenize("--123;")
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, IntegerLiteral{Value: 123}, prog)
	})

	t.Run("Parentheses first in top-level with more afterwards", func(t *testing.T) {
		lx := lexer.MustTokenize("(1 + 2) + 10;")
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
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
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name: "a",
			Value: BinaryExpression{
				Left: BinaryExpression{
					Left:  IntegerLiteral{Value: 1},
					Right: IntegerLiteral{Value: 2},
					Op:    "<",
				},
				Right: BooleanLiteral{Value: true},
				Op:    "&&",
			},
			Type: types.Unspecified,
		}, prog)
	})

	t.Run("Identifier declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("int a = abc;")
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:  "a",
			Type:  types.Int,
			Value: VarIdentifier{Name: "abc"},
		}, prog)
	})

	t.Run("Declare with return of function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() int {return 0;}

		var a = foo();
		`)
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Int,
			Params:     []Parameter{},
			Body: []Statement{
				Return{
					Value: IntegerLiteral{Value: 0},
				},
			},
		}, fnDef)

		p.next()

		vardef, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name: "a",
			Value: Call{
				Name: "foo",
				Args: []Expression{},
			},
			Type: types.Unspecified,
		}, vardef)
	})

	t.Run("Self-recursive global var", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var a = a;
		`)
		_, err := Parse(lx)
		assert.ErrorContains(t, err, "circular dependency detected: a")
	})

	t.Run("Out of order var declarations", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var a = b;
		var b = 123;
		`)
		pg := MustParse(lx)

		assert.EqualExportedValues(t, &Program{
			VarDecls: []VarDecl{
				{
					Name:  "b",
					Type:  types.Int,
					Value: IntegerLiteral{Value: 123},
				},
				{
					Name:  "a",
					Type:  types.Int,
					Value: VarIdentifier{Name: "b"},
				},
			},
		}, pg)
	})

	t.Run("Uninitialized typed variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a;
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:  "a",
			Type:  types.Int,
			Value: nil,
		}, prog)
	})

	t.Run("Var declared with struct type", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		MyStruct foo;
		}`)
		p := NewParser(lx)
		decl, err := p.parseVarDecl()
		assert.NoError(t, err)
		assert.Equal(t, "foo", decl.Name)
		assert.Equal(t, unresolvedTypeIdentifier{
			Name: "MyStruct",
		}, decl.Type.(unresolvedTypeIdentifier))
	})

	t.Run("typed with implicit struct literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		MyStruct foo = { x: 10, y: true, z: "hello" };
		}`)
		p := NewParser(lx)
		decl, err := p.parseVarDecl()
		assert.NoError(t, err)
		assert.Equal(t, "foo", decl.Name)
		assert.Equal(t, unresolvedTypeIdentifier{
			Name: "MyStruct",
		}, decl.Type.(unresolvedTypeIdentifier))
		assert.EqualExportedValues(t, StructLiteral{
			Type: types.Unspecified,
			Fields: []FieldLiteral{
				{Name: "x", Value: IntegerLiteral{Value: 10}},
				{Name: "y", Value: BooleanLiteral{Value: true}},
				{Name: "z", Value: StringLiteral{Value: "hello"}},
			},
		}, decl.Value)
	})

	t.Run("implicit var with typed struct literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var foo = MyStruct{ x: 10, y: true, z: "hello" };
		}`)
		p := NewParser(lx)
		decl, err := p.parseVarDecl()
		assert.NoError(t, err)
		assert.Equal(t, "foo", decl.Name)
		assert.Equal(t, types.Unspecified, decl.Type)
		assert.EqualExportedValues(t, StructLiteral{
			Type: unresolvedTypeIdentifier{Name: "MyStruct"},
			Fields: []FieldLiteral{
				{Name: "x", Value: IntegerLiteral{Value: 10}},
				{Name: "y", Value: BooleanLiteral{Value: true}},
				{Name: "z", Value: StringLiteral{Value: "hello"}},
			},
		}, decl.Value)
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
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  IntegerLiteral{Value: 1},
			Right: VarIdentifier{Name: "a"},
			Op:    "+",
		}, prog)
	})
}

func Test_FunctionDefinitions(t *testing.T) {
	t.Run("Simple function definition", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() {}")
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params:     []Parameter{},
			Body: []Statement{
				Return{Value: VoidLiteral{}},
			},
		}, fnDef)
	})

	t.Run("Simple function definition with body", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() {return 1 + 2;}")
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params:     []Parameter{},
			Body: []Statement{
				Return{
					Value: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "+",
					},
				},
			},
		}, fnDef)
	})

	t.Run("Function definition with return type", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() int {}")
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			Params:     []Parameter{},
			ReturnType: types.Int,
			Body:       []Statement{},
		}, fnDef)
	})

	t.Run("Function definition with one parameter", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo(int a) {}")
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params: []Parameter{
				{Name: "a", Type: types.Int},
			},
			Body: []Statement{
				Return{Value: VoidLiteral{}},
			},
		}, fnDef)
	})

	t.Run("Function definition with parameters", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo(int a, bool b, bool c) {}")
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params: []Parameter{
				{Name: "a", Type: types.Int},
				{Name: "b", Type: types.Bool},
				{Name: "c", Type: types.Bool},
			},
			Body: []Statement{
				Return{Value: VoidLiteral{}},
			},
		}, fnDef)
	})

	t.Run("Function definition with parameters and return type", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo(bool a, int b) bool {}")
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Bool,
			Params: []Parameter{
				{Name: "a", Type: types.Bool},
				{Name: "b", Type: types.Int},
			},
			Body: []Statement{},
		}, fnDef)
	})

	t.Run("function using local scope", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn foo() {
				var a = 123;
				return 1 + a;
			}
		`)

		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params:     []Parameter{},
			Body: []Statement{
				VarDecl{
					Name:  "a",
					Value: IntegerLiteral{Value: 123},
					Type:  types.Unspecified,
				},
				Return{
					Value: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: VarIdentifier{Name: "a"},
						Op:    "+",
					},
				},
			},
		}, fnDef)
	})

	t.Run("function using global and local scope", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() int {
				var b = 11;
				return a + b;
			}
		`)

		pg := MustParse(lx)

		assert.EqualExportedValues(t, &Program{
			VarDecls: []VarDecl{
				{
					Name:  "a",
					Value: IntegerLiteral{Value: 10},
					Type:  types.Int,
				},
			},
			FnDefs: []FnDef{
				{

					Name:       "main",
					ReturnType: types.Int,
					Params:     []Parameter{},
					Body: []Statement{
						VarDecl{
							Name:  "b",
							Value: IntegerLiteral{Value: 11},
							Type:  types.Int,
						},
						Return{
							Value: BinaryExpression{
								Left:  VarIdentifier{Name: "a"},
								Right: VarIdentifier{Name: "b"},
								Op:    "+",
							},
						},
					},
				},
			},
		}, pg)
	})

	t.Run("struct parameter", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			bool a;
			bool b;
		}

		fn bar(Foo p) bool {
			return p.a && p.b;
		}
		`)
		pg, err := NewParser(lx).Parse()
		assert.NoError(t, err)

		fooType := &types.Struct{
			Name:   "Foo",
			Fields: []types.Field{{Name: "a", Type: types.Bool}, {Name: "b", Type: types.Bool}},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				{
					Name:   "Foo",
					Fields: []types.Field{{Name: "a", Type: types.Bool}, {Name: "b", Type: types.Bool}},
				},
			},
			VarDecls: []VarDecl{},
			FnDefs: []FnDef{
				{
					Name:       "bar",
					ReturnType: types.Bool,
					Params:     []Parameter{{Name: "p", Type: fooType}},
					Body: []Statement{
						Return{
							Value: BinaryExpression{
								Left:  VarIdentifier{Name: "p.a"},
								Right: VarIdentifier{Name: "p.b"},
								Op:    "&&",
							},
						},
					},
				},
			},
		}, pg)
	})

	t.Run("struct return type", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Point {
			int x;
			int y;
		}

		fn getPoint() Point {
			return Point{x: 5, y: 10};
		}
		`)
		pg := MustParse(lx)

		pointType := &types.Struct{
			Name:   "Point",
			Fields: []types.Field{{Name: "x", Type: types.Int}, {Name: "y", Type: types.Int}},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				pointType,
			},
			VarDecls: []VarDecl{},
			FnDefs: []FnDef{
				{
					Name:       "getPoint",
					ReturnType: pointType,
					Params:     []Parameter{},
					Body: []Statement{
						Return{
							Value: StructLiteral{
								Type:   pointType,
								Fields: []FieldLiteral{{Name: "x", Value: IntegerLiteral{Value: 5}}, {Name: "y", Value: IntegerLiteral{Value: 10}}},
							},
						},
					},
				},
			},
		}, pg)
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

		pg := MustParse(lx)

		assert.EqualExportedValues(t, &Program{
			FnDefs: []FnDef{
				{
					Name:       "foo",
					ReturnType: types.Int,
					Params:     []Parameter{},
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
				{
					Name:       "bar",
					ReturnType: types.Void,
					Params:     []Parameter{},
					Body: []Statement{
						Return{Value: VoidLiteral{}},
					},
				},
				{
					Name:       "main",
					ReturnType: types.Int,
					Params:     []Parameter{},
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
			VarDecls: []VarDecl{},
		}, pg)
	})

	t.Run("no main function", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn bar() {}
		`)
		pg := MustParse(lx)
		assert.EqualExportedValues(t, &Program{
			FnDefs: []FnDef{
				{
					Name:       "bar",
					ReturnType: types.Void,
					Params:     []Parameter{},
					Body: []Statement{
						Return{Value: VoidLiteral{}},
					},
				},
			},
			VarDecls: []VarDecl{},
		}, pg)
	})
}

func Test_If(t *testing.T) {
	t.Run("if, compare static ints", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if 1 > 0 {
			foo = 10;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, If{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 0},
				Op:    ">",
			},
			Then: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 10},
				},
			},
		}, prog)

	})

	t.Run("if in func", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn main() int {
			if true {
				return 1;
			}
			return 0;
		}
		`)
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "main",
			ReturnType: types.Int,
			Params:     []Parameter{},
			Body: []Statement{
				If{
					Cond: BooleanLiteral{Value: true},
					Then: []Statement{
						Return{Value: IntegerLiteral{Value: 1}},
					},
				},
				Return{Value: IntegerLiteral{Value: 0}},
			},
		}, fnDef)
	})

	t.Run("if with else", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if 1 > 0 {
			foo = 10;
		} else {
			foo = 20;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, If{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 0},
				Op:    ">",
			},
			Then: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 10},
				},
			},
			Else: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 20},
				},
			},
		}, prog)
	})

	t.Run("if with else if", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if 1 > 0 {
			foo = 10;
		} else if 2 < 3 {
			foo = 20;
		} else {
			foo = 30;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, If{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 0},
				Op:    ">",
			},
			Then: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 10},
				},
			},
			Else: []Statement{
				If{
					Cond: BinaryExpression{
						Left:  IntegerLiteral{Value: 2},
						Right: IntegerLiteral{Value: 3},
						Op:    "<",
					},
					Then: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 20},
						},
					},
					Else: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 30},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("if with multiple else ifs", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if 1 > 0 {
			foo = 10;
		} else if 2 < 3 {
			foo = 20;
		} else if 3 > 4 {
			foo = 30;
		} else {
			foo = 40;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, If{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 0},
				Op:    ">",
			},
			Then: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 10},
				},
			},
			Else: []Statement{
				If{
					Cond: BinaryExpression{
						Left:  IntegerLiteral{Value: 2},
						Right: IntegerLiteral{Value: 3},
						Op:    "<",
					},
					Then: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 20},
						},
					},
					Else: []Statement{
						If{

							Cond: BinaryExpression{
								Left:  IntegerLiteral{Value: 3},
								Right: IntegerLiteral{Value: 4},
								Op:    ">",
							},
							Then: []Statement{
								Assignment{
									Name:  "foo",
									Value: IntegerLiteral{Value: 30},
								},
							},
							Else: []Statement{
								Assignment{
									Name:  "foo",
									Value: IntegerLiteral{Value: 40},
								},
							},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("if with nested if", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if 1 > 0 {
			if 2 < 3 {
				foo = 10;
			} else {
				foo = 20;
			}
		} else {
			foo = 30;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, If{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 0},
				Op:    ">",
			},
			Then: []Statement{
				If{
					Cond: BinaryExpression{
						Left:  IntegerLiteral{Value: 2},
						Right: IntegerLiteral{Value: 3},
						Op:    "<",
					},
					Then: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 10},
						},
					},
					Else: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 20},
						},
					},
				},
			},
			Else: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 30},
				},
			},
		}, prog)
	})
}

func Test_FunctionCall(t *testing.T) {
	t.Run("Function call, 0 args", func(t *testing.T) {
		lx := lexer.MustTokenize("foo();")
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Call{
			Name: "foo",
			Args: []Expression{},
		}, prog)
	})

	t.Run("Function call, 1 arg", func(t *testing.T) {
		lx := lexer.MustTokenize("foo(1);")
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)
		assert.EqualExportedValues(t, Call{
			Name: "foo",
			Args: []Expression{
				IntegerLiteral{Value: 1},
			},
		}, prog)
	})

	t.Run("Function call, 2 args", func(t *testing.T) {
		lx := lexer.MustTokenize("foo(1, 2);")
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Call{
			Name: "foo",
			Args: []Expression{
				IntegerLiteral{Value: 1},
				IntegerLiteral{Value: 2},
			},
		}, prog)
	})

	t.Run("Function call with mixed arguments", func(t *testing.T) {
		lx := lexer.MustTokenize("foo(1 + 2, true);")
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, Call{
			Name: "foo",
			Args: []Expression{
				BinaryExpression{
					Left:  IntegerLiteral{Value: 1},
					Right: IntegerLiteral{Value: 2},
					Op:    "+",
				},
				BooleanLiteral{Value: true},
			},
		}, prog)
	})
}

func Test_WhileStatements(t *testing.T) {
	t.Run("Simple while loop", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		while true {
			foo = 10;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, While{
			Cond: BooleanLiteral{Value: true},
			Body: []Statement{
				Assignment{
					Name:  "foo",
					Value: IntegerLiteral{Value: 10},
				},
			},
		}, prog)
	})

	t.Run("While loop with condition", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		while 1 < 2 {
			bar();
			baz = 20;
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, While{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 2},
				Op:    "<",
			},
			Body: []Statement{
				Call{
					Name: "bar",
					Args: []Expression{},
				},
				Assignment{
					Name:  "baz",
					Value: IntegerLiteral{Value: 20},
				},
			},
		}, prog)
	})

	t.Run("Function with while loop", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn main() {
			while 1 < 2 {
				bar();
				baz = 20;
			}
		}
		`)
		p := NewParser(lx)
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "main",
			ReturnType: types.Void,
			Params:     []Parameter{},
			Body: []Statement{
				While{
					Cond: BinaryExpression{
						Left:  IntegerLiteral{Value: 1},
						Right: IntegerLiteral{Value: 2},
						Op:    "<",
					},
					Body: []Statement{
						Call{
							Name: "bar",
							Args: []Expression{},
						},
						Assignment{
							Name:  "baz",
							Value: IntegerLiteral{Value: 20},
						},
					},
				},
				Return{Value: VoidLiteral{}},
			},
		}, fnDef)
	})

	t.Run("While loop with nested if", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		while 1 < 2 {
			if true {
				foo = 10;
			} else {
				foo = 20;
			}
		}
		`)
		p := NewParser(lx)
		prog, err := p.ParseStatement()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, While{
			Cond: BinaryExpression{
				Left:  IntegerLiteral{Value: 1},
				Right: IntegerLiteral{Value: 2},
				Op:    "<",
			},
			Body: []Statement{
				If{
					Cond: BooleanLiteral{Value: true},
					Then: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 10},
						},
					},
					Else: []Statement{
						Assignment{
							Name:  "foo",
							Value: IntegerLiteral{Value: 20},
						},
					},
				},
			},
		}, prog)
	})
}

func Test_Stub(t *testing.T) {
	t.Run("Simple stub", func(t *testing.T) {
		lx := lexer.MustTokenize("stub foo();")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			ReturnType: types.Void,
			Name:       "foo",
			Params:     []Parameter{},
			Body:       nil,
			Stub:       true,
		}, prog)
	})

	t.Run("Stub with parameters", func(t *testing.T) {
		lx := lexer.MustTokenize("stub foo(int a, bool b);")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params: []Parameter{
				{Name: "a", Type: types.Int},
				{Name: "b", Type: types.Bool},
			},
			Body: nil,
			Stub: true,
		}, prog)
	})

	t.Run("Stub with return type", func(t *testing.T) {
		lx := lexer.MustTokenize("stub foo() int;")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Int,
			Params:     []Parameter{},
			Body:       nil,
			Stub:       true,
		}, prog)
	})

	t.Run("Stub with parameters and return type", func(t *testing.T) {
		lx := lexer.MustTokenize("stub foo(int a, bool b) int;")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Int,
			Params: []Parameter{
				{Name: "a", Type: types.Int},
				{Name: "b", Type: types.Bool},
			},
			Body: nil,
			Stub: true,
		}, prog)
	})
}

func Test_ExportedFunction(t *testing.T) {
	t.Run("Simple exported function", func(t *testing.T) {
		lx := lexer.MustTokenize("export fn foo() {};")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params:     []Parameter{},
			Body: []Statement{
				Return{Value: VoidLiteral{}},
			},
			Exported: true,
		}, prog)
	})

	t.Run("Exported function with parameters", func(t *testing.T) {
		lx := lexer.MustTokenize("export fn foo(int a, bool b) {};")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Void,
			Params: []Parameter{
				{Name: "a", Type: types.Int},
				{Name: "b", Type: types.Bool},
			},
			Body: []Statement{
				Return{Value: VoidLiteral{}},
			},
			Exported: true,
		}, prog)
	})

	t.Run("Exported function with return type", func(t *testing.T) {
		lx := lexer.MustTokenize("export fn foo() int {};")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Int,
			Params:     []Parameter{},
			Body:       []Statement{},
			Exported:   true,
		}, prog)
	})

	t.Run("Exported function with parameters and return type", func(t *testing.T) {
		lx := lexer.MustTokenize("export fn foo(int a, bool b) int {};")
		p := NewParser(lx)
		prog, err := p.parseFnDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, FnDef{
			Name:       "foo",
			ReturnType: types.Int,
			Params: []Parameter{
				{Name: "a", Type: types.Int},
				{Name: "b", Type: types.Bool},
			},
			Body:     []Statement{},
			Exported: true,
		}, prog)
	})
}

func Test_ExportedVariable(t *testing.T) {
	t.Run("Exported variable, inferred type", func(t *testing.T) {
		lx := lexer.MustTokenize("export var foo = 10;")
		p := NewParser(lx)
		prog, err := p.parseVarDecl()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:     "foo",
			Type:     types.Unspecified,
			Value:    IntegerLiteral{Value: 10},
			Exported: true,
		}, prog)
	})

	t.Run("Exported variable", func(t *testing.T) {
		lx := lexer.MustTokenize("export int foo = 10;")
		p := NewParser(lx)
		prog, err := p.parseVarDecl()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:     "foo",
			Type:     types.Int,
			Value:    IntegerLiteral{Value: 10},
			Exported: true,
		}, prog)
	})
}

func Test_StringLiteral(t *testing.T) {
	t.Run("Simple string literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`"hello world";`)
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, StringLiteral{Value: "hello world"}, prog)
	})

	t.Run("String literal with escape sequences", func(t *testing.T) {
		lx := lexer.MustTokenize(`"hello\nworld\t!";`)
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, StringLiteral{Value: "hello\nworld\t!"}, prog)
	})

	t.Run("String literal in expression", func(t *testing.T) {
		lx := lexer.MustTokenize(`"foo" + "bar";`)
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left:  StringLiteral{Value: "foo"},
			Right: StringLiteral{Value: "bar"},
			Op:    "+",
		}, prog)
	})

	t.Run("String literal in complex expression", func(t *testing.T) {
		lx := lexer.MustTokenize(`"foo" + "bar" + (1 < 4) * 5;`)
		p := NewParser(lx)
		prog, err := p.ParseExpr()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, BinaryExpression{
			Left: BinaryExpression{
				Left:  StringLiteral{Value: "foo"},
				Right: StringLiteral{Value: "bar"},
				Op:    "+",
			},
			Right: BinaryExpression{
				Left: BinaryExpression{
					Left:  IntegerLiteral{Value: 1},
					Right: IntegerLiteral{Value: 4},
					Op:    "<",
				},
				Right: IntegerLiteral{Value: 5},
				Op:    "*",
			},
			Op: "+",
		}, prog)
	})

	t.Run("String variable", func(t *testing.T) {
		lx := lexer.MustTokenize(`string foo = "hello";`)
		p := NewParser(lx)
		prog, err := p.parseVarDecl()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, VarDecl{
			Name:  "foo",
			Type:  types.String,
			Value: StringLiteral{Value: "hello"},
		}, prog)
	})
}

func Test_StructDef(t *testing.T) {
	t.Run("Simple struct", func(t *testing.T) {
		lx := lexer.MustTokenize(`struct Point { int x; int y; }`)
		p := NewParser(lx)
		prog, err := p.parseStructDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, types.Struct{
			Name: "Point",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
				{Name: "y", Type: types.Int},
			},
		}, prog)
	})

	t.Run("Struct with no fields", func(t *testing.T) {
		lx := lexer.MustTokenize(`struct Point {}`)
		p := NewParser(lx)
		prog, err := p.parseStructDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, types.Struct{
			Name:   "Point",
			Fields: []types.Field{},
		}, prog)
	})

	t.Run("Struct with various field types", func(t *testing.T) {
		lx := lexer.MustTokenize(`struct Point { int x; bool y; string z; }`)
		p := NewParser(lx)
		prog, err := p.parseStructDef()
		assert.NoError(t, err)

		assert.EqualExportedValues(t, types.Struct{
			Name: "Point",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
				{Name: "y", Type: types.Bool},
				{Name: "z", Type: types.String},
			},
		}, prog)
	})

	t.Run("Typed and uninitialized variable of struct type", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
			bool y;
			string z;
		}
		Foo p;
		`)
		prog, err := Parse(lx)
		assert.NoError(t, err)

		fooStruct := &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
				{Name: "y", Type: types.Bool},
				{Name: "z", Type: types.String},
			},
		}
		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				fooStruct,
			},
			VarDecls: []VarDecl{
				{
					Name: "p",
					Type: fooStruct,
					Value: StructLiteral{
						Type: fooStruct,
						Fields: []FieldLiteral{
							{Name: "x", Value: IntegerLiteral{Value: 0}},
							{Name: "y", Value: BooleanLiteral{Value: false}},
							{Name: "z", Value: StringLiteral{Value: ""}},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("typed variable with implicit struct literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
			bool y;
			string z;
		}
		Foo p = { x: 10, y: true, z: "hello" };
		`)
		prog, err := Parse(lx)
		assert.NoError(t, err)

		fooStruct := &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
				{Name: "y", Type: types.Bool},
				{Name: "z", Type: types.String},
			},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				fooStruct,
			},
			VarDecls: []VarDecl{
				{
					Name: "p",
					Type: fooStruct,
					Value: StructLiteral{
						Type: fooStruct,
						Fields: []FieldLiteral{
							{Name: "x", Value: IntegerLiteral{Value: 10}},
							{Name: "y", Value: BooleanLiteral{Value: true}},
							{Name: "z", Value: StringLiteral{Value: "hello"}},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("implicit variable with implicit struct literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
		}
		var b = { x: 10 };
		`)
		_, err := Parse(lx)
		assert.ErrorContains(t, err, "cannot use an implicit struct literal without an explicitly typed variable")
	})

	t.Run("typed variable with explicit struct literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
		}
		Foo p = Foo{ x: 10 };
		`)
		prog, err := Parse(lx)
		assert.NoError(t, err)

		fooType := &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
			},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				fooType,
			},
			VarDecls: []VarDecl{
				{
					Name: "p",
					Type: fooType,
					Value: StructLiteral{
						Type: fooType,
						Fields: []FieldLiteral{
							{Name: "x", Value: IntegerLiteral{Value: 10}},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("implicit variable with explicit struct literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
		}
		var p = Foo{ x: 10 };
		`)
		prog, err := Parse(lx)
		assert.NoError(t, err)

		fooType := &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
			},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				fooType,
			},
			VarDecls: []VarDecl{
				{
					Name: "p",
					Type: fooType,
					Value: StructLiteral{
						Type: fooType,
						Fields: []FieldLiteral{
							{Name: "x", Value: IntegerLiteral{Value: 10}},
						},
					},
				},
			},
		}, prog)
	})
}

func Test_StructFieldAccess(t *testing.T) {
	t.Run("field access", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
		}
		Foo p;
		`)
		prog, err := Parse(lx)
		assert.NoError(t, err)

		fooType := &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
			},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				fooType,
			},
			VarDecls: []VarDecl{
				{
					Name: "p",
					Type: fooType,
					Value: StructLiteral{
						Type: fooType,
						Fields: []FieldLiteral{
							{Name: "x", Value: IntegerLiteral{Value: 0}},
						},
					},
				},
			},
		}, prog)
	})

	t.Run("assign to field", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Foo {
			int x;
		}
		fn main() int {
			Foo p;
			p.x = 5;
			return p.x;
		}
		`)
		prog, err := Parse(lx)
		assert.NoError(t, err)

		fooType := &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
			},
		}

		assert.EqualExportedValues(t, &Program{
			Structs: []*types.Struct{
				fooType,
			},
			VarDecls: []VarDecl{},
			FnDefs: []FnDef{
				{
					Name:       "main",
					Params:     []Parameter{},
					ReturnType: types.Int,
					Body: []Statement{
						VarDecl{
							Name: "p",
							Type: fooType,
							Value: StructLiteral{
								Type: fooType,
								Fields: []FieldLiteral{
									{Name: "x", Value: IntegerLiteral{Value: 0}},
								},
							},
						},
						Assignment{
							Name:  "p.x",
							Value: IntegerLiteral{Value: 5},
						},
						Return{
							Value: VarIdentifier{Name: "p.x"},
						},
					},
				},
			},
		}, prog)
	})
}
