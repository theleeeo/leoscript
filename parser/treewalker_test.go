package parser

import (
	"fmt"
	"leoscript/lexer"
	"leoscript/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TreeWalking(t *testing.T) {
	program := Program{
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

	callbacks := TreeWalkerCallbacks{
		VarDecl: func(wctx WalkingContext, varDecl *VarDecl) *VarDecl {
			fmt.Println("Visiting variable declaration:", varDecl.Name)
			return varDecl
		},
		FnDefs: func(wctx WalkingContext, fnDef *FnDef) *FnDef {
			fmt.Println("Visiting function defintion:", fnDef.Name)
			return fnDef
		},
		Call: func(wctx WalkingContext, call *Call) *Call {
			fmt.Println("Visiting function call:", call.Name)
			return call
		},
	}

	treeWalker := NewTreeWalker(callbacks)
	pg := treeWalker.WalkProgram(program)
	assert.Equal(t, len(pg.VarDecls), 2, "Expected 2 variable declarations")
}

func Test_TreeWalking_RemoveVarDecl(t *testing.T) {
	program := Program{
		VarDecls: []VarDecl{
			{Name: "x", Value: BooleanLiteral{Value: true}},
			{Name: "y", Value: IntegerLiteral{Value: 42}},
		},
		FnDefs: []FnDef{
			{Name: "foo", Args: []Argument{}, Body: []Statement{}, ReturnType: types.Void},
			{Name: "bar", Args: []Argument{}, Body: []Statement{}, ReturnType: types.Int},
		},
	}

	callbacks := TreeWalkerCallbacks{
		VarDecl: func(wctx WalkingContext, varDecl *VarDecl) *VarDecl {
			if varDecl.Name == "x" {
				fmt.Println("Removing variable declaration:", varDecl.Name)
				return nil // Return nil to remove it
			}
			return varDecl
		},
		FnDefs: func(wctx WalkingContext, fnDef *FnDef) *FnDef {
			if fnDef.Name == "foo" {
				fmt.Println("Removing function definition:", fnDef.Name)
				return nil // Return nil to remove it
			}
			return fnDef
		},
	}

	treeWalker := NewTreeWalker(callbacks)
	pg := treeWalker.WalkProgram(program)
	assert.Equal(t, 1, len(pg.VarDecls), "Expected 1 variable declaration after removal")
	assert.Equal(t, 1, len(pg.FnDefs), "Expected 1 function definition after removal")
	assert.Equal(t, pg.VarDecls[0].Name, "y", "Expected remaining variable declaration to be 'y'")
	assert.Equal(t, pg.FnDefs[0].Name, "bar", "Expected remaining function definition to be 'bar'")
}

func Test_TreeWalking_Remove_2(t *testing.T) {
	lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				var b = 11;
				return a + b;
			}
		`)

	p := NewParser(lx)
	prog, err := p.ParseFile()
	assert.NoError(t, err)

	callbacks := TreeWalkerCallbacks{
		VarDecl: func(wctx WalkingContext, varDecl *VarDecl) *VarDecl {
			fmt.Println("Visiting variable declaration:", varDecl.Name)
			return varDecl
		},
		FnDefs: func(wctx WalkingContext, fnDef *FnDef) *FnDef {
			fmt.Println("Visiting function definition:", fnDef.Name)
			return fnDef
		},
		VarIdentifier: func(wctx WalkingContext, varIdentifier *VarIdentifier) *VarIdentifier {
			fmt.Println("Visiting variable identifier:", varIdentifier.Name)
			// if varIdentifier.Name == "a" {
			// 	fmt.Println("Removing variable identifier:", varIdentifier.Name)
			// 	return nil // Return nil to remove it
			// }
			return varIdentifier
		},
	}

	treeWalker := NewTreeWalker(callbacks)
	_ = treeWalker.WalkProgram(prog)

}

func Test_TreeWalking_AllNodesVisited(t *testing.T) {
	t.Run("global variable declarations", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			VarDecl: func(wctx WalkingContext, varDecl *VarDecl) *VarDecl {
				visitedNodes = append(visitedNodes, varDecl.Name)
				return varDecl
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

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

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			FnDefs: func(wctx WalkingContext, fnDef *FnDef) *FnDef {
				visitedNodes = append(visitedNodes, fnDef.Name)
				return fnDef
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 2, len(pg.FnDefs))
		assert.Equal(t, "foo", pg.FnDefs[0].Name)
		assert.Equal(t, "bar", pg.FnDefs[1].Name)

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
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			FnDefs: func(wctx WalkingContext, fnDef *FnDef) *FnDef {
				for _, arg := range fnDef.Args {
					visitedNodes = append(visitedNodes, arg.Name)
				}
				return fnDef
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

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

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			Call: func(wctx WalkingContext, call *Call) *Call {
				visitedNodes = append(visitedNodes, call.Name)
				return call
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 3, len(pg.FnDefs))
		assert.Equal(t, "main", pg.FnDefs[0].Name)
		assert.Equal(t, "add", pg.FnDefs[1].Name)
		assert.Equal(t, "subtract", pg.FnDefs[2].Name)

		assert.ElementsMatch(t, []string{"add", "subtract"}, visitedNodes)
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
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			VarIdentifier: func(wctx WalkingContext, varIdentifier *VarIdentifier) *VarIdentifier {
				visitedNodes = append(visitedNodes, varIdentifier.Name)
				return varIdentifier
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"a", "b", "a", "b"}, visitedNodes)
	})

	t.Run("if statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;

			fn main() {
				if (a < b) {
					return a;
				} else {
					return b;
				}
			}
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			If: func(wctx WalkingContext, ifStmt *If) *If {
				visitedNodes = append(visitedNodes, "if")
				return ifStmt
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"if"}, visitedNodes)
	})

	t.Run("return statements", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				return a;
			}
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			Return: func(wctx WalkingContext, returnStmt *Return) *Return {
				visitedNodes = append(visitedNodes, "return")
				return returnStmt
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

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

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			Assignment: func(wctx WalkingContext, assignment *Assignment) *Assignment {
				visitedNodes = append(visitedNodes, assignment.Name)
				return assignment
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"a"}, visitedNodes)
	})

	t.Run("complex expressions", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;
			var b = 20;

			fn main() {
				var result = (a + b) * (a - b);
				return result;
			}
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			BinaryExpression: func(wctx WalkingContext, binaryExpr *BinaryExpression) *BinaryExpression {
				visitedNodes = append(visitedNodes, binaryExpr.String())
				return binaryExpr
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 2, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)
		assert.Equal(t, "b", pg.VarDecls[1].Name)

		assert.ElementsMatch(t, []string{"((a + b) * (a - b))", "(a + b)", "(a - b)"}, visitedNodes)
	})

	t.Run("unary expressions", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				var negA = -a;
				return negA;
			}
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			UnaryExpression: func(wctx WalkingContext, unaryExpr *UnaryExpression) *UnaryExpression {
				visitedNodes = append(visitedNodes, unaryExpr.String())
				return unaryExpr
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

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

			fn main() {
				var result = add(a, 5);
				return result;
			}
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			Call: func(wctx WalkingContext, call *Call) *Call {
				visitedNodes = append(visitedNodes, call.String())
				return call
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

		assert.Equal(t, 1, len(pg.VarDecls))
		assert.Equal(t, "a", pg.VarDecls[0].Name)

		assert.ElementsMatch(t, []string{"add(a, 5)"}, visitedNodes)
	})

	t.Run("While loops", func(t *testing.T) {
		lx := lexer.MustTokenize(`
			var a = 10;

			fn main() {
				while (a > 0) {
					a = a - 1;
				}
				return a;
			}
		`)

		p := NewParser(lx)
		prog, err := p.ParseFile()
		assert.NoError(t, err)

		var visitedNodes []string
		callbacks := TreeWalkerCallbacks{
			While: func(wctx WalkingContext, whileStmt *While) *While {
				visitedNodes = append(visitedNodes, "while")
				return whileStmt
			},
		}

		treeWalker := NewTreeWalker(callbacks)
		pg := treeWalker.WalkProgram(prog)

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
	// 	prog, err := p.ParseFile()
	// 	assert.NoError(t, err)

	// 	var visitedNodes []string
	// 	callbacks := TreeWalkerCallbacks{
	// 		StubDef: func(wctx WalkingContext, stubDef *StubDef) *StubDef {
	// 			visitedNodes = append(visitedNodes, stubDef.Name)
	// 			return stubDef
	// 		},
	// 	}

	// 	treeWalker := NewTreeWalker(callbacks)
	// 	pg := treeWalker.WalkProgram(prog)

	// 	assert.Equal(t, 2, len(pg.StubDefs))
	// 	assert.Equal(t, "foo", pg.StubDefs[0].Name)
	// 	assert.Equal(t, "bar", pg.StubDefs[1].Name)

	// 	assert.ElementsMatch(t, []string{"foo", "bar"}, visitedNodes)
	// })
}
