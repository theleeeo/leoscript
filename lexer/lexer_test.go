package lexer_test

import (
	"leoscript/lexer"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MathExpression(t *testing.T) {
	t.Run("Single digit", func(t *testing.T) {
		lx := lexer.MustTokenize("1")
		assert.Equal(t, []lexer.Token{
			lexer.Integer{Value: 1},
		}, lx)
	})

	t.Run("Single number", func(t *testing.T) {
		lx := lexer.MustTokenize("12345")
		assert.Equal(t, []lexer.Token{
			lexer.Integer{Value: 12345},
		}, lx)
	})

	t.Run("Binary ops with whitespace", func(t *testing.T) {
		lx := lexer.MustTokenize("1+ 2- 3 *4/ 5")
		assert.Equal(t, []lexer.Token{
			lexer.Integer{Value: 1},
			lexer.Operator{Op: "+"},
			lexer.Integer{Value: 2},
			lexer.Operator{Op: "-"},
			lexer.Integer{Value: 3},
			lexer.Operator{Op: "*"},
			lexer.Integer{Value: 4},
			lexer.Operator{Op: "/"},
			lexer.Integer{Value: 5},
		}, lx)
	})

	t.Run("Multiple digit numbers", func(t *testing.T) {
		lx := lexer.MustTokenize("123+456789-987 7898 / 898989")
		assert.Equal(t, []lexer.Token{
			lexer.Integer{Value: 123},
			lexer.Operator{Op: "+"},
			lexer.Integer{Value: 456789},
			lexer.Operator{Op: "-"},
			lexer.Integer{Value: 987},
			lexer.Integer{Value: 7898},
			lexer.Operator{Op: "/"},
			lexer.Integer{Value: 898989},
		}, lx)
	})

	t.Run("Invalid character", func(t *testing.T) {
		_, err := lexer.Tokenize("1+2$3")
		assert.ErrorContains(t, err, "invalid character: $")
	})

	t.Run("Parentheses", func(t *testing.T) {
		lx := lexer.MustTokenize("((1+2)*3);")
		assert.Equal(t, []lexer.Token{
			lexer.OpenParen{},
			lexer.OpenParen{},
			lexer.Integer{Value: 1},
			lexer.Operator{Op: "+"},
			lexer.Integer{Value: 2},
			lexer.CloseParen{},
			lexer.Operator{Op: "*"},
			lexer.Integer{Value: 3},
			lexer.CloseParen{},
			lexer.Semicolon{},
		}, lx)
	})

}

func Test_Identifiers(t *testing.T) {
	t.Run("Identifiers", func(t *testing.T) {
		lx := lexer.MustTokenize("foo + bar-baz")
		assert.Equal(t, []lexer.Token{
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "+"},
			lexer.Identifier{Value: "bar"},
			lexer.Operator{Op: "-"},
			lexer.Identifier{Value: "baz"},
		}, lx)
	})

	t.Run("Reserved keywords", func(t *testing.T) {
		lx := lexer.MustTokenize("true false")
		assert.Equal(t, []lexer.Token{
			lexer.Boolean{Value: true},
			lexer.Boolean{Value: false},
		}, lx)
	})

	t.Run("Combined keywords, fine", func(t *testing.T) {
		lx := lexer.MustTokenize("truefalse")
		assert.Equal(t, []lexer.Token{
			lexer.Identifier{Value: "truefalse"},
		}, lx)
	})

	t.Run("Mixed identifiers and keywords", func(t *testing.T) {
		lx := lexer.MustTokenize("true foo false bar")
		assert.Equal(t, []lexer.Token{
			lexer.Boolean{Value: true},
			lexer.Identifier{Value: "foo"},
			lexer.Boolean{Value: false},
			lexer.Identifier{Value: "bar"},
		}, lx)
	})
}

func Test_LogicalExpressions(t *testing.T) {
	t.Run("Logical operators", func(t *testing.T) {
		lx := lexer.MustTokenize("true && false || true")
		assert.Equal(t, []lexer.Token{
			lexer.Boolean{Value: true},
			lexer.Operator{Op: "&&"},
			lexer.Boolean{Value: false},
			lexer.Operator{Op: "||"},
			lexer.Boolean{Value: true},
		}, lx)
	})

	t.Run("Parentheses", func(t *testing.T) {
		lx := lexer.MustTokenize("(true && false) || true")
		assert.Equal(t, []lexer.Token{
			lexer.OpenParen{},
			lexer.Boolean{Value: true},
			lexer.Operator{Op: "&&"},
			lexer.Boolean{Value: false},
			lexer.CloseParen{},
			lexer.Operator{Op: "||"},
			lexer.Boolean{Value: true},
		}, lx)
	})

	t.Run("With identifiers", func(t *testing.T) {
		lx := lexer.MustTokenize("true && bar || baz")
		assert.Equal(t, []lexer.Token{
			lexer.Boolean{Value: true},
			lexer.Operator{Op: "&&"},
			lexer.Identifier{Value: "bar"},
			lexer.Operator{Op: "||"},
			lexer.Identifier{Value: "baz"},
		}, lx)
	})

	t.Run("Single character ops invalid, for now", func(t *testing.T) {
		_, err := lexer.Tokenize("true & false | true")
		assert.ErrorContains(t, err, "invalid character: &")
	})

	t.Run("Comparison operators", func(t *testing.T) {
		lx := lexer.MustTokenize("1 < 2 > 3 <= 4 >= 5 == 6 != 7")
		assert.Equal(t, []lexer.Token{
			lexer.Integer{Value: 1},
			lexer.Operator{Op: "<"},
			lexer.Integer{Value: 2},
			lexer.Operator{Op: ">"},
			lexer.Integer{Value: 3},
			lexer.Operator{Op: "<="},
			lexer.Integer{Value: 4},
			lexer.Operator{Op: ">="},
			lexer.Integer{Value: 5},
			lexer.Operator{Op: "=="},
			lexer.Integer{Value: 6},
			lexer.Operator{Op: "!="},
			lexer.Integer{Value: 7},
		}, lx)
	})
}

func Test_VariableDeclaration(t *testing.T) {
	t.Run("Variable declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("var foo = 123;")
		assert.Equal(t, []lexer.Token{
			lexer.VarDecl{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 123},
			lexer.Semicolon{},
		}, lx)
	})

	t.Run("Variable declaration with expression", func(t *testing.T) {
		lx := lexer.MustTokenize("var foo = 1 + 2 * 3;")
		assert.Equal(t, []lexer.Token{
			lexer.VarDecl{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 1},
			lexer.Operator{Op: "+"},
			lexer.Integer{Value: 2},
			lexer.Operator{Op: "*"},
			lexer.Integer{Value: 3},
			lexer.Semicolon{},
		}, lx)
	})

	t.Run("Integer variable declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("int foo = 123;")
		assert.Equal(t, []lexer.Token{
			lexer.Identifier{Value: "int"},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 123},
			lexer.Semicolon{},
		}, lx)
	})

	t.Run("Boolean variable declaration", func(t *testing.T) {
		lx := lexer.MustTokenize("bool foo = true;")
		assert.Equal(t, []lexer.Token{
			lexer.Identifier{Value: "bool"},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Boolean{Value: true},
			lexer.Semicolon{},
		}, lx)
	})

	// t.Run("String variable declaration", func(t *testing.T) {
	// err := lexer.MustTokenize("string foo = \"bar\";"
	// 	assert.Equal(t, []lexer.Token{
	// 		lexer.StringDecl{},
	// 		lexer.Identifier{Value: "foo"},
	// 		lexer.Operator{Op: "="},
	// 		lexer.String{Value: "bar"},
	// 		lexer.Semicolon{},
	// 	}, lx)
	// })
}

func Test_FunctionDefinition(t *testing.T) {
	t.Run("Function definition", func(t *testing.T) {
		lx := lexer.MustTokenize("fn foo() {}")
		assert.Equal(t, []lexer.Token{
			lexer.FnDef{},
			lexer.Identifier{Value: "foo"},
			lexer.OpenParen{},
			lexer.CloseParen{},
			lexer.OpenBrace{},
			lexer.CloseBrace{},
		}, lx)
	})

	// t.Run("Function definition, single parameter", func(t *testing.T) {
	// err := lexer.MustTokenize("fn foo(a) {}"
	// })

	// t.Run("Function definition, return type", func(t *testing.T) {
	// err := lexer.MustTokenize("fn foo(a) int {}"
	// 	}, lx)
	// })

	// t.Run("Function definition, multiple parameters", func(t *testing.T) {
	// err := lexer.MustTokenize("fn foo(a, b) {}"
	// })

	// t.Run("Function definition, body", func(t *testing.T) {
	// err := lexer.MustTokenize("fn foo(a) { return 1 + 2; }"
	// })
}

func Test_ControlFlow(t *testing.T) {
	t.Run("If block, empty body", func(t *testing.T) {
		lx := lexer.MustTokenize("if a==b {}")
		assert.Equal(t, []lexer.Token{
			lexer.If{},
			lexer.Identifier{Value: "a"},
			lexer.Operator{Op: "=="},
			lexer.Identifier{Value: "b"},
			lexer.OpenBrace{},
			lexer.CloseBrace{},
		}, lx)
	})

	t.Run("If block, with body", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if foo >= bar {
			foo = 10;
		}
		`)
		assert.Equal(t, []lexer.Token{
			lexer.If{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: ">="},
			lexer.Identifier{Value: "bar"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 10},
			lexer.Semicolon{},
			lexer.CloseBrace{},
		}, lx)
	})

	t.Run("If-else block", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if foo >= bar {
			foo = 10;
		} else {
			bar = 20;
		}
		`)
		assert.Equal(t, []lexer.Token{
			lexer.If{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: ">="},
			lexer.Identifier{Value: "bar"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 10},
			lexer.Semicolon{},
			lexer.CloseBrace{},
			lexer.Else{},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "bar"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 20},
			lexer.Semicolon{},
			lexer.CloseBrace{},
		}, lx)
	})

	t.Run("If-else if-else block", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		if foo >= bar {
			foo = 10;
		} else if bar < baz {
			bar = 20;
		} else {
			baz = 30;
		}
		`)
		assert.Equal(t, []lexer.Token{
			lexer.If{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: ">="},
			lexer.Identifier{Value: "bar"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 10},
			lexer.Semicolon{},
			lexer.CloseBrace{},
			lexer.Else{},
			lexer.If{},
			lexer.Identifier{Value: "bar"},
			lexer.Operator{Op: "<"},
			lexer.Identifier{Value: "baz"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "bar"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 20},
			lexer.Semicolon{},
			lexer.CloseBrace{},
			lexer.Else{},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "baz"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 30},
			lexer.Semicolon{},
			lexer.CloseBrace{},
		}, lx)
	})

	t.Run("While loop", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		while foo < bar {
			foo = foo + 1;
		}
		`)
		assert.Equal(t, []lexer.Token{
			lexer.While{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "<"},
			lexer.Identifier{Value: "bar"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "="},
			lexer.Identifier{Value: "foo"},
			lexer.Operator{Op: "+"},
			lexer.Integer{Value: 1},
			lexer.Semicolon{},
			lexer.CloseBrace{},
		}, lx)
	})
}

func Test_Comments(t *testing.T) {
	t.Run("Single-line comment", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		// This is a comment
		var x = 10;
		`)
		assert.Equal(t, []lexer.Token{
			lexer.VarDecl{},
			lexer.Identifier{Value: "x"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 10},
			lexer.Semicolon{},
		}, lx)
	})

	t.Run("Multi-line comment", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		/* 
		This is a
		multi-line comment
		*/
		var y = 20;
		`)
		assert.Equal(t, []lexer.Token{
			lexer.VarDecl{},
			lexer.Identifier{Value: "y"},
			lexer.Operator{Op: "="},
			lexer.Integer{Value: 20},
			lexer.Semicolon{},
		}, lx)
	})
}

func Test_StringLiterals(t *testing.T) {
	t.Run("Basic string literal", func(t *testing.T) {
		lx := lexer.MustTokenize(`"hello world"`)
		assert.Equal(t, []lexer.Token{
			lexer.StringLiteral{Value: "hello world"},
		}, lx)
	})

	t.Run("String literal with escaped quotes", func(t *testing.T) {
		lx := lexer.MustTokenize(`"hello \"world\" "`)
		assert.Equal(t, []lexer.Token{
			lexer.StringLiteral{Value: "hello \"world\" "},
		}, lx)
	})

	t.Run("Unclosed string literal", func(t *testing.T) {
		lx, err := lexer.Tokenize(`"hello world`)
		assert.ErrorContains(t, err, "unclosed string literal")
		assert.Nil(t, lx)
	})

	t.Run("String with escape sequences", func(t *testing.T) {
		lx := lexer.MustTokenize(`"\\hello\nworld\t\\"`)
		assert.Equal(t, []lexer.Token{
			lexer.StringLiteral{Value: "\\hello\nworld\t\\"},
		}, lx)
	})
}

func Test_Structs(t *testing.T) {
	t.Run("Struct definition", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		struct Point {
			x: int,
			y: int
		}
		`)
		assert.Equal(t, []lexer.Token{
			lexer.StructDef{},
			lexer.Identifier{Value: "Point"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "x"},
			lexer.Colon{},
			lexer.Identifier{Value: "int"},
			lexer.Comma{},
			lexer.Identifier{Value: "y"},
			lexer.Colon{},
			lexer.Identifier{Value: "int"},
			lexer.CloseBrace{},
		}, lx)
	})

	t.Run("Field access", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		var p = Point { x: 10, y: 20 };
		var x = p.x;
		var y = p.y;
		`)
		assert.Equal(t, []lexer.Token{
			lexer.VarDecl{},
			lexer.Identifier{Value: "p"},
			lexer.Operator{Op: "="},
			lexer.Identifier{Value: "Point"},
			lexer.OpenBrace{},
			lexer.Identifier{Value: "x"},
			lexer.Colon{},
			lexer.Integer{Value: 10},
			lexer.Comma{},
			lexer.Identifier{Value: "y"},
			lexer.Colon{},
			lexer.Integer{Value: 20},
			lexer.CloseBrace{},
			lexer.Semicolon{},
			lexer.VarDecl{},
			lexer.Identifier{Value: "x"},
			lexer.Operator{Op: "="},
			lexer.Identifier{Value: "p.x"},
			lexer.Semicolon{},
			lexer.VarDecl{},
			lexer.Identifier{Value: "y"},
			lexer.Operator{Op: "="},
			lexer.Identifier{Value: "p.y"},
			lexer.Semicolon{},
		}, lx)
	})
}
