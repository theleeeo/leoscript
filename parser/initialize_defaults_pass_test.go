package parser

import (
	"leoscript/lexer"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_InitializeBasics(t *testing.T) {
	lx := lexer.MustTokenize(`
	int a;
	bool b;
	string c;
	`)
	program, err := Parse(lx)
	assert.NoError(t, err)

	assert.Equal(t, IntegerLiteral{Value: 0}, program.VarDecls[0].Value)
	assert.Equal(t, BooleanLiteral{Value: false}, program.VarDecls[1].Value)
	assert.Equal(t, StringLiteral{Value: ""}, program.VarDecls[2].Value)
}

func Test_InitializeStruct(t *testing.T) {
	lx := lexer.MustTokenize(`
	struct Foo {
		int x;
		bool y;
		string z;
	}
	Foo p;
	`)
	program, err := Parse(lx)
	assert.NoError(t, err)

	assert.Equal(t, StructLiteral{
		Type: &types.Struct{
			Name: "Foo",
			Fields: []types.Field{
				{Name: "x", Type: types.Int},
				{Name: "y", Type: types.Bool},
				{Name: "z", Type: types.String},
			},
		},
		Fields: []FieldLiteral{
			{Name: "x", Value: IntegerLiteral{Value: 0}},
			{Name: "y", Value: BooleanLiteral{Value: false}},
			{Name: "z", Value: StringLiteral{Value: ""}},
		},
	}, program.VarDecls[0].Value)
}
