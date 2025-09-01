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
		assert.Len(t, md.functions, 3)
		assert.Equal(t, ExportedFunction{
			Name:        "foo",
			ReturnType:  types.Void,
			StartOffset: 1,
			exported:    true,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[0])

		assert.Equal(t, ExportedFunction{
			Name:        "bar",
			ReturnType:  types.Void,
			StartOffset: 2,
			exported:    true,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[1])

		assert.Equal(t, ExportedFunction{
			Name:        "baz",
			ReturnType:  types.Void,
			StartOffset: 3,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[2])

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 3)
		assert.Equal(t, ExportedFunction{
			Name:        "foo",
			ReturnType:  types.Void,
			StartOffset: 1,
			exported:    true,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[0])

		assert.Equal(t, ExportedFunction{
			Name:        "bar",
			ReturnType:  types.Void,
			StartOffset: 2,
			exported:    true,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[1])

		assert.Equal(t, ExportedFunction{
			Name:        "baz",
			ReturnType:  types.Void,
			StartOffset: 3,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[2])
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
		assert.Equal(t, "a", md.variables[0].Name)
		assert.Equal(t, types.Int, md.variables[0].VarType)
		assert.Equal(t, uint64(0), md.variables[0].Offset)

		assert.Equal(t, "b", md.variables[1].Name)
		assert.Equal(t, types.Bool, md.variables[1].VarType)
		assert.Equal(t, uint64(8), md.variables[1].Offset)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md.variables, 2)
		assert.Equal(t, "a", md.variables[0].Name)
		assert.Equal(t, types.Int, md.variables[0].VarType)
		assert.Equal(t, uint64(0), md.variables[0].Offset)

		assert.Equal(t, "b", md.variables[1].Name)
		assert.Equal(t, types.Bool, md.variables[1].VarType)
		assert.Equal(t, uint64(8), md.variables[1].Offset)
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
		assert.Len(t, md.functions, 2)
		assert.Equal(t, ExportedFunction{
			Name:        "foo",
			ReturnType:  types.Void,
			StartOffset: 37,
			exported:    true,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[0])
		assert.Equal(t, ExportedFunction{
			Name:        "bar",
			ReturnType:  types.Void,
			StartOffset: 38,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[1])

		assert.Len(t, md.variables, 1)
		assert.Equal(t, "a", md.variables[0].Name)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 2)
		assert.Equal(t, ExportedFunction{
			Name:        "foo",
			ReturnType:  types.Void,
			StartOffset: 37,
			exported:    true,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[0])
		assert.Equal(t, ExportedFunction{
			Name:        "bar",
			ReturnType:  types.Void,
			StartOffset: 38,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[1])

		assert.Len(t, md2.variables, 1)
		assert.Equal(t, "a", md2.variables[0].Name)
	})

	t.Run("no exports", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn foo() {}
		fn bar() {}
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.functions, 2)
		assert.Equal(t, ExportedFunction{
			Name:        "foo",
			ReturnType:  types.Void,
			StartOffset: 1,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[0])
		assert.Equal(t, ExportedFunction{
			Name:        "bar",
			ReturnType:  types.Void,
			StartOffset: 2,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md.functions[1])

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 2)
		assert.Equal(t, ExportedFunction{
			Name:        "foo",
			ReturnType:  types.Void,
			StartOffset: 1,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[0])
		assert.Equal(t, ExportedFunction{
			Name:        "bar",
			ReturnType:  types.Void,
			StartOffset: 2,
			exported:    false,
			stub:        false,
			Params:      []FnParam{},
		}, md2.functions[1])
	})

	t.Run("function with parameters and returntype", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		export fn add(int a, int b) int {
			return a + b;
		}
		`)
		pg := parser.MustParse(lx)
		exe := Compile(pg)

		md := exe.Metadata()
		assert.Len(t, md.functions, 1)
		assert.Equal(t, "add", md.functions[0].Name)
		assert.Equal(t, types.Int, md.functions[0].ReturnType)
		assert.Len(t, md.functions[0].Params, 2)
		assert.Equal(t, types.Int, md.functions[0].Params[0].Type)
		assert.Equal(t, types.Int, md.functions[0].Params[1].Type)

		md2 := new(Metadata)
		md2.Unmarshal(md.Marshal())
		assert.Len(t, md2.functions, 1)
		assert.Equal(t, "add", md2.functions[0].Name)
		assert.Equal(t, types.Int, md2.functions[0].ReturnType)
		assert.Len(t, md2.functions[0].Params, 2)
		assert.Equal(t, types.Int, md2.functions[0].Params[0].Type)
		assert.Equal(t, types.Int, md2.functions[0].Params[1].Type)
	})
}
