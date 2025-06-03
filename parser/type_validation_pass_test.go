package parser

import (
	"leoscript/lexer"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateTypes(t *testing.T) {
	t.Run("Variable type in return", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a = 1;

		fn b() int {
			return a;
		}
		`)
		p := Parser{tokens: lx}
		varDef, err := p.parseVarDecl()
		assert.NoError(t, err)

		p.next()
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		pg := Program{
			VarDecls: []VarDecl{varDef},
			FnDefs:   []FnDef{fnDef},
		}

		pg, err = TypeResolvingPass(pg)
		assert.NoError(t, err)

		_, err = TypeValidationPass(pg)
		assert.NoError(t, err)
	})
}

func Test_ValidateTypes_Invalid(t *testing.T) {
	t.Run("Bad constant assignment", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		bool a = 1;
		`)
		p := Parser{tokens: lx}
		varDef, err := p.parseVarDecl()
		assert.NoError(t, err)
		pg := Program{
			VarDecls: []VarDecl{varDef},
			FnDefs:   []FnDef{},
		}
		pg, err = TypeResolvingPass(pg)
		assert.NoError(t, err)
		_, err = TypeValidationPass(pg)
		assert.Error(t, err)
		assert.Equal(t, "type mismatch: expected Bool, got Int", err.Error())
	})

	t.Run("Bad variable assignment", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		bool a = true;
		int b = a;
		`)
		p := Parser{tokens: lx}
		varDef, err := p.parseVarDecl()
		assert.NoError(t, err)
		p.next()
		varDef2, err := p.parseVarDecl()
		assert.NoError(t, err)
		pg := Program{
			VarDecls: []VarDecl{varDef, varDef2},
			FnDefs:   []FnDef{},
		}
		pg, err = TypeResolvingPass(pg)
		assert.NoError(t, err)
		_, err = TypeValidationPass(pg)
		assert.Error(t, err)
		assert.Equal(t, "type mismatch: expected Int, got Bool", err.Error())
	})

	t.Run("Variable type in return", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		int a = 1;

		fn b() bool {
			return a;
		}
		`)
		p := Parser{tokens: lx}
		varDef, err := p.parseVarDecl()
		assert.NoError(t, err)

		p.next()
		fnDef, err := p.parseFnDef()
		assert.NoError(t, err)

		pg := Program{
			VarDecls: []VarDecl{varDef},
			FnDefs:   []FnDef{fnDef},
		}

		pg, err = TypeResolvingPass(pg)
		assert.NoError(t, err)

		_, err = TypeValidationPass(pg)
		assert.Equal(t, "type mismatch: expected Bool, got Int", err.Error())
	})
}
