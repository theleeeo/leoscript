package parser

import (
	"leoscript/lexer"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResolveTypes(t *testing.T) {
	t.Run("Resolve variable type", func(t *testing.T) {
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

		assert.Equal(t, types.Unspecified, fnDef.Body[0].(Return).Value.ReturnType())

		pg := Program{
			VarDecls: []VarDecl{varDef},
			FnDefs:   []FnDef{fnDef},
		}

		pg, err = TypeResolvingPass(pg, NewScope(nil))
		assert.NoError(t, err)

		assert.Equal(t, types.Int, pg.FnDefs[0].Body[0].(Return).Value.ReturnType())
	})
}
