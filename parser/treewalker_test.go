package parser

import (
	"leoscript/lexer"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TreeWalking(t *testing.T) {
	prog := &Program{
		VarDecls: []VarDecl{
			{Name: "x", Value: BooleanLiteral{Value: true}},
			{Name: "y", Value: IntegerLiteral{Value: 42}},
		},
		FnDefs: []FnDef{
			{Name: "foo", Args: []Argument{}, Body: []Statement{}, ReturnType: types.Void},
			{Name: "bar", Args: []Argument{{Name: "a", Type: types.Int}}, Body: []Statement{
				VarDecl{Name: "result", Value: IntegerLiteral{Value: 0}},
				Assignment{
					Name: "result",
					Value: Call{
						Name: "foo",
						Args: []Expression{IntegerLiteral{Value: 10}},
					},
				},
			}, ReturnType: types.Int},
		},
	}

	callback := func(wctx WalkingContext, stmt Statement) Statement {
		return stmt
	}

	tw := NewTreeWalker(callback)
	err := tw.WalkProgram(prog)
	assert.NoError(t, err)
	assert.Equal(t, len(prog.VarDecls), 2, "Expected 2 variable declarations")
}

func Test_TreeWalking_RemoveVarDecl(t *testing.T) {
	prog := &Program{
		VarDecls: []VarDecl{
			{Name: "x", Value: BooleanLiteral{Value: true}},
			{Name: "y", Value: IntegerLiteral{Value: 42}},
		},
		FnDefs: []FnDef{
			{Name: "foo", Args: []Argument{}, Body: []Statement{}, ReturnType: types.Void},
			{Name: "bar", Args: []Argument{}, Body: []Statement{}, ReturnType: types.Int},
		},
	}

	callback := func(wctx WalkingContext, stmt Statement) Statement {
		switch s := stmt.(type) {
		case VarDecl:
			if s.Name == "x" {
				return nil // Remove it
			}
		case FnDef:
			if s.Name == "foo" {
				return nil // Remove it
			}
		}
		return stmt
	}

	tw := NewTreeWalker(callback)
	err := tw.WalkProgram(prog)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(prog.VarDecls), "Expected 1 variable declaration after removal")
	assert.Equal(t, 1, len(prog.FnDefs), "Expected 1 function definition after removal")
	assert.Equal(t, "y", prog.VarDecls[0].Name, "Expected remaining variable declaration to be 'y'")
	assert.Equal(t, "bar", prog.FnDefs[0].Name, "Expected remaining function definition to be 'bar'")
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
// 	prog, err := p.ParseFile()
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
// 	_ = treeWalker.WalkProgram(prog)

// }

func Test_TreeWalking_AllNodesVisited(t *testing.T) {
	t.Run("global variable declarations", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;
		`)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case VarDecl:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)
		assert.Equal(t, "b", prog.VarDecls[1].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case FnDef:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(prog.FnDefs))
		assert.Equal(t, "foo", prog.FnDefs[0].Name)
		assert.Equal(t, "bar", prog.FnDefs[1].Name)

		assert.ElementsMatch(t, []string{"foo", "bar"}, visitedNodes)
	})

	t.Run("function arguments", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			fn add(int a, int b) int {
				return a + b;
			}

			fn subtract(int x, int y) int {
				return x - y;
			}
		`)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case Argument:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(prog.FnDefs))
		assert.Equal(t, "add", prog.FnDefs[0].Name)
		assert.Equal(t, "subtract", prog.FnDefs[1].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case Call:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(prog.FnDefs))
		assert.Equal(t, "main", prog.FnDefs[0].Name)
		assert.Equal(t, "add", prog.FnDefs[1].Name)
		assert.Equal(t, "subtract", prog.FnDefs[2].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case VarIdentifier:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)
		assert.Equal(t, "b", prog.VarDecls[1].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case If:
				visitedNodes = append(visitedNodes, "if")
				return n
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)
		assert.Equal(t, "b", prog.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"if"}, visitedNodes)
	})

	t.Run("return statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() int {
				return a;
			}
		`)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch node.(type) {
			case Return:
				visitedNodes = append(visitedNodes, "return")
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"return"}, visitedNodes)
	})

	t.Run("assignment statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				a = a + 5;
			}
		`)

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case Assignment:
				visitedNodes = append(visitedNodes, n.Name)
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case BinaryExpression:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)
		assert.Equal(t, "b", prog.VarDecls[1].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case UnaryExpression:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch n := node.(type) {
			case Call:
				visitedNodes = append(visitedNodes, n.String())
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)

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

		p := NewParser(lx)
		prog, err := p.Parse()
		assert.NoError(t, err)

		var visitedNodes []string
		callback := func(wctx WalkingContext, node Statement) Statement {
			switch node.(type) {
			case While:
				visitedNodes = append(visitedNodes, "while")
			}
			return node
		}

		tw := NewTreeWalker(callback)
		err = tw.WalkProgram(prog)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(prog.VarDecls))
		assert.Equal(t, "a", prog.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"while"}, visitedNodes)
	})

	// t.Run("Stub definitions", func(t *testing.T) {
	// 	lx := lexer.MustTokenize(`
	// 		stub foo(int x, int y);
	// 		stub bar();
	// 	`)

	// 	p := NewParser(lx)
	// 	prog, err := p.ParseFile()
	// 	assert.NoError(t, err)

	// 	var visitedNodes []string
	// 	callbacks := TreeWalkerCallbacks{
	// 		StubDef: func(wctx WalkingContext, stubDef *StubDef) *StubDef {
	// 			visitedNodes = append(visitedNodes, stubDef.Name)
	// 			return stubDef
	// 		},
	// 	}

	// 	treeWalker := NewTreeWalker2(callback)
	// 	err := tw.WalkProgram(prog)
	// 	assert.NoError(t, err)

	// 	assert.Equal(t, 2, len(pg.StubDefs))
	// 	assert.Equal(t, "foo", pg.StubDefs[0].Name)
	// 	assert.Equal(t, "bar", pg.StubDefs[1].Name)

	// 	assert.ElementsMatch(t, []string{"foo", "bar"}, visitedNodes)
	// })
}
