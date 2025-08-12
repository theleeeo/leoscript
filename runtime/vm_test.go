package runtime_test

import (
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
		exe := compiler.CompileStatement(expr)
		vm := runtime.NewVM(exe)
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 5, ret)
	})

	t.Run("negative number expression", func(t *testing.T) {
		lx := lexer.MustTokenize("5 - 6;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		vm := runtime.NewVM(exe)
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, -1, ret)
	})

	t.Run("multiplication and division binary expression", func(t *testing.T) {
		lx := lexer.MustTokenize("3 * 4 - 4 / 2;")
		expr, _ := parser.NewParser(lx).ParseExpr()
		exe := compiler.CompileStatement(expr)
		vm := runtime.NewVM(exe)
		ret, err := vm.Run()
		assert.NoError(t, err)
		assert.Equal(t, 10, ret)
	})

}
