package parser

import (
	"leoscript/lexer"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TreeWalking(t *testing.T) {
	pg := &Program{
		VarDecls: []VarDecl{
			{Name: "x", Value: BooleanLiteral{Value: true}},
			{Name: "y", Value: IntegerLiteral{Value: 42}},
		},
		FnDefs: []FnDef{
			{Name: "foo", Params: []Parameter{}, Body: []Statement{}, ReturnType: types.Void},
			{Name: "bar", Params: []Parameter{{Name: "a", Type: types.Int}}, Body: []Statement{
				VarDecl{Name: "result", Value: IntegerLiteral{Value: 0}},
				Assignment{
					Target: VariableTarget("result"),
					Value: Call{
						Name: "foo",
						Args: []Expression{IntegerLiteral{Value: 10}},
					},
				},
			}, ReturnType: types.Int},
		},
	}

	callback := func(wctx WalkingContext, stmt Statement) (Statement, error) {
		return stmt, nil
	}

	tw := NewTreeWalker(TreeWalkerConfig{
		CallbackFn: callback,
	})
	err := tw.WalkProgram(pg)
	assert.NoError(t, err)
	assert.Equal(t, len(pg.VarDecls), 2, "Expected 2 variable declarations")
}

func Test_TreeWalking_RemoveVarDecl(t *testing.T) {
	pg := &Program{
		VarDecls: []VarDecl{
			{Name: "x", Value: BooleanLiteral{Value: true}},
			{Name: "y", Value: IntegerLiteral{Value: 42}},
		},
		FnDefs: []FnDef{
			{Name: "foo", Params: []Parameter{}, Body: []Statement{}, ReturnType: types.Void},
			{Name: "bar", Params: []Parameter{}, Body: []Statement{}, ReturnType: types.Int},
		},
	}

	callback := func(wctx WalkingContext, stmt Statement) (Statement, error) {
		switch s := stmt.(type) {
		case VarDecl:
			if s.Name == "x" {
				return nil, nil // Remove it
			}
		case FnDef:
			if s.Name == "foo" {
				return nil, nil // Remove it
			}
		}
		return stmt, nil
	}

	tw := NewTreeWalker(TreeWalkerConfig{
		CallbackFn: callback,
	})
	err := tw.WalkProgram(pg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pg.VarDecls), "Expected 1 variable declaration after removal")
	assert.Equal(t, 1, len(pg.FnDefs), "Expected 1 function definition after removal")
	assert.Equal(t, "y", pg.VarDecls[0].Name, "Expected remaining variable declaration to be 'y'")
	assert.Equal(t, "bar", pg.FnDefs[0].Name, "Expected remaining function definition to be 'bar'")
}

// func Test_TreeWalking_Remove_2(t *testing.T) {
// 	lx := lexer.MustTokenize(`
// 			var a = 10;

// 			fn main() {
// 				var b = 11;
// 				return a + b;
// 			}
// 		`)

// 	p := NewParser(lx)
// 	pg, err := p.ParseFile()
// 	assert.NoError(t, err)

// 	callbacks := TreeWalkerCallbacks{
// 		VarDecl: func(wctx WalkingContext, varDecl *VarDecl) *VarDecl {
// 			return varDecl
// 		},
// 		FnDefs: func(wctx WalkingContext, fnDef *FnDef) *FnDef {
// 			return fnDef
// 		},
// 		VarIdentifier: func(wctx WalkingContext, varIdentifier *VarIdentifier) *VarIdentifier {
// 			// if varIdentifier.Name == "a" {
// 			// 	return nil // Return nil to remove it
// 			// }
// 			return varIdentifier
// 		},
// 	}

// 	treeWalker := NewTreeWalker2(callback)
// 	_ = treeWalker.WalkProgram(pg)

// }

func Test_TreeWalking_AllNodesVisited(t *testing.T) {
	t.Run("global variable declarations", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case VarDecl:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"a", "b"}, visitedNodes)
	})

	t.Run("function definitions", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn foo() {
				return;
			}

			fn bar(int x) int {
				return x + 1;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case FnDef:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(pg.FnDefs))
		assert.Equal(t, "foo", pg.FnDefs[0].Name)
		assert.Equal(t, "bar", pg.FnDefs[1].Name)

		assert.ElementsMatch(t, []string{"foo", "bar"}, visitedNodes)
	})

	t.Run("function parameters", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn add(int a, int b) int {
				return a + b;
			}

			fn subtract(int x, int y) int {
				return x - y;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case Parameter:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(pg.FnDefs))
		assert.Equal(t, "add", pg.FnDefs[0].Name)
		assert.Equal(t, "subtract", pg.FnDefs[1].Name)

		assert.ElementsMatch(t, []string{"a", "b", "x", "y"}, visitedNodes)
	})

	t.Run("function calls", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn main() {
				var result = add(5, 10);
				var diff = subtract(20, 5);
			}

			fn add(int a, int b) int {
				return a + b;
			}

			fn subtract(int x, int y) int {
				return x - y;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case Call:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(pg.FnDefs))
		assert.Equal(t, "main", pg.FnDefs[0].Name)
		assert.Equal(t, "add", pg.FnDefs[1].Name)
		assert.Equal(t, "subtract", pg.FnDefs[2].Name)

		assert.ElementsMatch(t, []string{"add(5, 10)", "subtract(20, 5)"}, visitedNodes)
	})

	t.Run("variable identifiers", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;

			fn main() {
				var sum = a + b;
				var product = a * b;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case VarIdentifier:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"a", "b", "a", "b"}, visitedNodes)
	})

	t.Run("if statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;

			fn main() int {
				if (a < b) {
					return a;
				} else {
					return b;
				}
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch node.(type) {
			case If:
				visitedNodes = append(visitedNodes, "if")
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"if"}, visitedNodes)
	})

	t.Run("return statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() int {
				return a;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch node.(type) {
			case Return:
				visitedNodes = append(visitedNodes, "return")
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"return"}, visitedNodes)
	})

	t.Run("assignment statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				a = a + 5;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case Assignment:
				visitedNodes = append(visitedNodes, string(n.Target.(VariableTarget)))
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"a"}, visitedNodes)
	})

	t.Run("complex expressions", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;

			fn main() int {
				var result = (a + b) * (a - b);
				return result;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case BinaryExpression:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"((a + b) * (a - b))", "(a + b)", "(a - b)"}, visitedNodes)
	})

	t.Run("unary expressions", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() int {
				var negA = -a;
				return negA;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case UnaryExpression:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"-a"}, visitedNodes)
	})

	t.Run("function calls with arguments", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn add(int x, int y) int {
				return x + y;
			}

			fn main() int {
				var result = add(a, 5);
				return result;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case Call:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"add(a, 5)"}, visitedNodes)
	})

	t.Run("While loops", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() int {
				while (a > 0) {
					a = a - 1;
				}
				return a;
			}
		`)

		pg := MustParse(lx)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) (Statement, error) {
			switch node.(type) {
			case While:
				visitedNodes = append(visitedNodes, "while")
			}
			return node, nil
		}

		tw := NewTreeWalker(TreeWalkerConfig{
			CallbackFn: callback,
		})
		err := tw.WalkProgram(pg)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"while"}, visitedNodes)
	})

	// t.Run("Stub definitions", func(t *testing.T) {
	// 	lx := lexer.MustTokenize(`
	// 		stub foo(int x, int y);
	// 		stub bar();
	// 	`)

	// 	p := NewParser(lx)
	// 	pg, err := p.ParseFile()
	// 	assert.NoError(t, err)

	// 	var visitedNodes []string
	// 	callbacks := TreeWalkerCallbacks{
	// 		StubDef: func(wctx WalkingContext, stubDef *StubDef) *StubDef {
	// 			visitedNodes = append(visitedNodes, stubDef.Name)
	// 			return stubDef
	// 		},
	// 	}

	// 	treeWalker := NewTreeWalker2(callback)
	// 	err := tw.WalkProgram(pg)
	// 	assert.NoError(t, err)

	// 	assert.Equal(t, 2, len(pg.StubDefs))
	// 	assert.Equal(t, "foo", pg.StubDefs[0].Name)
	// 	assert.Equal(t, "bar", pg.StubDefs[1].Name)

	// 	assert.ElementsMatch(t, []string{"foo", "bar"}, visitedNodes)
	// })
}
