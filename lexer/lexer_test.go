package lexer_test

import (
	"leoscript/lexer"
	"leoscript/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MathExpression(t *testing.T) {
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
			token.Operator{Op: "+"},
			token.Integer{Value: 2},
			token.Operator{Op: "-"},
			token.Integer{Value: 3},
			token.Operator{Op: "*"},
			token.Integer{Value: 4},
			token.Operator{Op: "/"},
			token.Integer{Value: 5},
		}, lx)
	})

	t.Run("Multiple digit numbers", func(t *testing.T) {
		lx, err := lexer.Tokenize("123+456789-987 7898 / 898989")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Integer{Value: 123},
			token.Operator{Op: "+"},
			token.Integer{Value: 456789},
			token.Operator{Op: "-"},
			token.Integer{Value: 987},
			token.Integer{Value: 7898},
			token.Operator{Op: "/"},
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
			token.Operator{Op: "+"},
			token.Integer{Value: 2},
			token.CloseParen{},
			token.Operator{Op: "*"},
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
			token.Operator{Op: "+"},
			token.Identifier{Value: "bar"},
			token.Operator{Op: "-"},
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

func Test_LogicalExpressions(t *testing.T) {
	t.Run("Logical operators", func(t *testing.T) {
		lx, err := lexer.Tokenize("true && false || true")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Boolean{Value: true},
			token.Operator{Op: "&&"},
			token.Boolean{Value: false},
			token.Operator{Op: "||"},
			token.Boolean{Value: true},
		}, lx)
	})

	t.Run("Parentheses", func(t *testing.T) {
		lx, err := lexer.Tokenize("(true && false) || true")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.OpenParen{},
			token.Boolean{Value: true},
			token.Operator{Op: "&&"},
			token.Boolean{Value: false},
			token.CloseParen{},
			token.Operator{Op: "||"},
			token.Boolean{Value: true},
		}, lx)
	})

	t.Run("With identifiers", func(t *testing.T) {
		lx, err := lexer.Tokenize("true && bar || baz")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Boolean{Value: true},
			token.Operator{Op: "&&"},
			token.Identifier{Value: "bar"},
			token.Operator{Op: "||"},
			token.Identifier{Value: "baz"},
		}, lx)
	})

	t.Run("Single character ops invalid, for now", func(t *testing.T) {
		_, err := lexer.Tokenize("true & false | true")
		assert.ErrorContains(t, err, "invalid character: &")
	})

	t.Run("Comparison operators", func(t *testing.T) {
		lx, err := lexer.Tokenize("1 < 2 > 3 <= 4 >= 5 == 6 != 7")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.Integer{Value: 1},
			token.Operator{Op: "<"},
			token.Integer{Value: 2},
			token.Operator{Op: ">"},
			token.Integer{Value: 3},
			token.Operator{Op: "<="},
			token.Integer{Value: 4},
			token.Operator{Op: ">="},
			token.Integer{Value: 5},
			token.Operator{Op: "=="},
			token.Integer{Value: 6},
			token.Operator{Op: "!="},
			token.Integer{Value: 7},
		}, lx)
	})
}

func Test_VariableDeclaration(t *testing.T) {
	t.Run("Variable declaration", func(t *testing.T) {
		lx, err := lexer.Tokenize("var foo = 123;")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.VarDecl{},
			token.Identifier{Value: "foo"},
			token.Operator{Op: "="},
			token.Integer{Value: 123},
			token.Semicolon{},
		}, lx)
	})

	t.Run("Variable declaration with expression", func(t *testing.T) {
		lx, err := lexer.Tokenize("var foo = 1 + 2 * 3;")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.VarDecl{},
			token.Identifier{Value: "foo"},
			token.Operator{Op: "="},
			token.Integer{Value: 1},
			token.Operator{Op: "+"},
			token.Integer{Value: 2},
			token.Operator{Op: "*"},
			token.Integer{Value: 3},
			token.Semicolon{},
		}, lx)
	})

	t.Run("Integer variable declaration", func(t *testing.T) {
		lx, err := lexer.Tokenize("int foo = 123;")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.IntDecl{},
			token.Identifier{Value: "foo"},
			token.Operator{Op: "="},
			token.Integer{Value: 123},
			token.Semicolon{},
		}, lx)
	})

	t.Run("Boolean variable declaration", func(t *testing.T) {
		lx, err := lexer.Tokenize("bool foo = true;")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.BoolDecl{},
			token.Identifier{Value: "foo"},
			token.Operator{Op: "="},
			token.Boolean{Value: true},
			token.Semicolon{},
		}, lx)
	})

	// t.Run("String variable declaration", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("string foo = \"bar\";")
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, []token.Token{
	// 		token.StringDecl{},
	// 		token.Identifier{Value: "foo"},
	// 		token.Operator{Op: "="},
	// 		token.String{Value: "bar"},
	// 		token.Semicolon{},
	// 	}, lx)
	// })
}

func Test_FunctionDefinition(t *testing.T) {
	t.Run("Function definition", func(t *testing.T) {
		lx, err := lexer.Tokenize("fn foo() {}")
		assert.NoError(t, err)
		assert.Equal(t, []token.Token{
			token.FnDef{},
			token.Identifier{Value: "foo"},
			token.OpenParen{},
			token.CloseParen{},
			token.OpenBrace{},
			token.CloseBrace{},
		}, lx)
	})

	// t.Run("Function definition, single argument", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo(a) {}")
	// 	assert.NoError(t, err)
	// })

	// t.Run("Function definition, return type", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo(a) int {}")
	// 	assert.NoError(t, err)
	// 	}, lx)
	// })

	// t.Run("Function definition, multiple arguments", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo(a, b) {}")
	// 	assert.NoError(t, err)
	// })

	// t.Run("Function definition, body", func(t *testing.T) {
	// 	lx, err := lexer.Tokenize("fn foo(a) { return 1 + 2; }")
	// 	assert.NoError(t, err)
	// })
}
