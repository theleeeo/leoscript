package runtime_test

import (
	"encoding/binary"
	"leoscript/compiler"
	"leoscript/lexer"
	"leoscript/parser"
	"leoscript/runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VM_ArithmeticExpr(t *testing.T) {
	t.Run("Single binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("2 + 3;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr, nil)
		vm := runtime.NewVM(exe)
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 5, ret)
	})

	t.Run("negative number expression", func(t *testing.T) {
		lx := lexer.MustTokenize("5 - 6;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr, nil)
		vm := runtime.NewVM(exe)
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, -1, ret)
	})

	t.Run("multiplication and division binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("3 * 4 - 4 / 2;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr, nil)
		vm := runtime.NewVM(exe)
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 10, ret)
	})
}

func Test_VM_Variables(t *testing.T) {
	t.Run("Variable assignment and retrieval", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a = 10;
		int b = a + 20;
		`)
		pg, _ := parser.NewParser(lx).Parse()
		exe := compiler.Compile(pg)
		vm := runtime.NewVM(exe.Raw())
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 0, ret)
		assert.Equal(t, uint64(10), binary.BigEndian.Uint64(vm.VariableStack()[0:8]))
		assert.Equal(t, uint64(30), binary.BigEndian.Uint64(vm.VariableStack()[8:16]))
	})

	// t.Run("Variable reassignment", func(t *testing.T) {
	// 	lx := lexer.MustTokenize("var x = 5; var x = 15;")
	// 	prog, _ := parser.NewParser(lx).Parse()
	// 	exe := compiler.Compile(prog)
	// 	vm := runtime.NewVM(exe.Raw())
	// 	ret, err := vm.Run()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, 0, ret)
	// 	assert.Equal(t, uint64(30), binary.BigEndian.Uint64(vm.VariableStack()[8:16]))
	// })
}
