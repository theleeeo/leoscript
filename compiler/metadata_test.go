package compiler

import (
	"leoscript/lexer"
	"leoscript/parser"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Metadata_Marshaling(t *testing.T) {
	t.Run("function metadata", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn foo() {}
		export fn bar() {}
		fn baz() {}
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.functions, 2)
		assert.Equal(t, "foo", md.functions[0].name)
		assert.Equal(t, types.Void, md.functions[0].returnType)
		assert.Equal(t, uint64(1), md.functions[0].startOffset)

		assert.Equal(t, "bar", md.functions[1].name)
		assert.Equal(t, types.Void, md.functions[1].returnType)
		assert.Equal(t, uint64(2), md.functions[1].startOffset)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 2)
		assert.Equal(t, "foo", md2.functions[0].name)
		assert.Equal(t, "bar", md2.functions[1].name)
	})

	t.Run("variable metadata", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export int a = 5;
		export bool b = false;
		int c = 15;
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.variables, 2)
		assert.Equal(t, "a", md.variables[0].name)
		assert.Equal(t, types.Int, md.variables[0].varType)
		assert.Equal(t, uint64(0), md.variables[0].offset)

		assert.Equal(t, "b", md.variables[1].name)
		assert.Equal(t, types.Bool, md.variables[1].varType)
		assert.Equal(t, uint64(8), md.variables[1].offset)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md.variables, 2)
		assert.Equal(t, "a", md.variables[0].name)
		assert.Equal(t, types.Int, md.variables[0].varType)
		assert.Equal(t, uint64(0), md.variables[0].offset)

		assert.Equal(t, "b", md.variables[1].name)
		assert.Equal(t, types.Bool, md.variables[1].varType)
		assert.Equal(t, uint64(8), md.variables[1].offset)
	})

	t.Run("metadata with functions and variables", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn foo() {}
		export int a = 5;
		fn bar() {}
		int b = 10;
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.functions, 1)
		assert.Equal(t, "foo", md.functions[0].name)

		assert.Len(t, md.variables, 1)
		assert.Equal(t, "a", md.variables[0].name)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 1)
		assert.Equal(t, "foo", md2.functions[0].name)

		assert.Len(t, md2.variables, 1)
		assert.Equal(t, "a", md2.variables[0].name)
	})

	t.Run("no exports", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {}
		fn bar() {}
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.functions, 0)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 0)
	})

	t.Run("function with arguments and returntype", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.functions, 1)
		assert.Equal(t, "add", md.functions[0].name)
		assert.Equal(t, types.Int, md.functions[0].returnType)
		assert.Len(t, md.functions[0].args, 2)
		assert.Equal(t, types.Int, md.functions[0].args[0].argType)
		assert.Equal(t, types.Int, md.functions[0].args[1].argType)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 1)
		assert.Equal(t, "add", md2.functions[0].name)
		assert.Equal(t, types.Int, md2.functions[0].returnType)
		assert.Len(t, md2.functions[0].args, 2)
		assert.Equal(t, types.Int, md2.functions[0].args[0].argType)
		assert.Equal(t, types.Int, md2.functions[0].args[1].argType)
	})
}
