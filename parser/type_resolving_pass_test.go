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

		pg := &Program{
			VarDecls: []VarDecl{varDef},
			FnDefs:   []FnDef{fnDef},
		}

		err = typeResolvingPass(pg)
		assert.NoError(t, err)

		assert.Equal(t, types.Int, pg.FnDefs[0].Body[0].(Return).Value.ReturnType())
	})

	t.Run("Resolve function call type", func(t *testing.T) {
		lx := lexer.MustTokenize(`
		fn add(int a, int b) int {
			return a + b;
		}

		fn main() int {
			return add(1, 2);
		}
		`)
		p := Parser{tokens: lx}
		fnDef1, err := p.parseFnDef()
		assert.NoError(t, err)

		p.next()
		fnDef2, err := p.parseFnDef()
		assert.NoError(t, err)

		pg := &Program{
			FnDefs: []FnDef{fnDef1, fnDef2},
		}

		err = typeResolvingPass(pg)
		assert.NoError(t, err)

		assert.Equal(t, types.Int, pg.FnDefs[1].Body[0].(Return).Value.ReturnType())
	})
}
