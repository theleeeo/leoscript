package parser

import (
	"errors"
	"fmt"
	"slices"
)

// TODO: You are better than this
func must(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	// The TreeWalker does a top-down traversal of the AST, first doing a callback for a node and only later walking its children.
	// In some cases however, you want the children to have been walked first.
	// By returning this error, the TreeWalker will walk the children before calling the callback again.
	ErrReturnLater = errors.New("walk children first") // Ugly hack or whatever but darn it, it makes the api simpler.
)

type WalkingContext struct {
	Scope *Scope

	// The function whose body we are currently walking.
	// Will be nil if we are not in a function body (only in the global scope).
	ParentFn *FnDef
}

type TreeWalker struct {
	CallbackFn func(wctx WalkingContext, node Statement) (Statement, error)
}

// NewTreeWalker creates a new TreeWalker with the provided callbacks.
func NewTreeWalker(callbackFn func(wctx WalkingContext, node Statement) (Statement, error)) *TreeWalker {
	return &TreeWalker{
		CallbackFn: callbackFn,
	}
}

// TODO: Do a check and avoid walking nil statements or expressions.
// If a statement or expression is nil, it should be removed from the ast by some cleaning pass in the end.
//
// TODO: Pass information about if ErrReturnLater was used
// WalkProgram walks through the AST and calls the callback for (almost) each node.
func (tw *TreeWalker) WalkProgram(program *Program) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	globalScope := NewScope(nil)

	// Setup the global scope with the global variable declarations and function definitions.
	for _, varDecl := range program.VarDecls {
		must(globalScope.RegisterVar(varDecl))
	}

	for _, fn := range program.FnDefs {
		must(globalScope.RegisterFn(fn))
	}

	for _, stub := range program.StubDefs {
		must(globalScope.RegisterFn(stub))
	}

	i := 0
	for i < len(program.VarDecls) {
		globalScope.deregisterVar(program.VarDecls[i].Name)

		resp := tw.walkStatement(program.VarDecls[i], WalkingContext{
			Scope:    globalScope,
			ParentFn: nil, // No parent function in the global scope
		})
		if resp == nil {
			// If the callback returns nil, we remove the variable declaration.
			program.VarDecls = slices.Delete(program.VarDecls, i, i+1)
			continue
		}
		program.VarDecls[i] = resp.(VarDecl)

		i++
	}

	i = 0
	for i < len(program.FnDefs) {
		globalScope.deregisterFn(program.FnDefs[i].Name)

		resp := tw.walkStatement(program.FnDefs[i], WalkingContext{
			Scope:    globalScope,
			ParentFn: &program.FnDefs[i],
		})
		if resp == nil {
			// If the callback returns nil, we remove the function definition.
			program.FnDefs = slices.Delete(program.FnDefs, i, i+1)
			continue
		}
		program.FnDefs[i] = resp.(FnDef)

		i++
	}

	return nil
}

func (tw *TreeWalker) walkStatement(stmt Statement, wctx WalkingContext) Statement {
	retStmt, callbackErr := tw.CallbackFn(wctx, stmt)
	if callbackErr != nil && callbackErr != ErrReturnLater {
		panic(callbackErr)
	}

	// If the ReturnLater sentinel is returned, retStmt is not yet populated and will be nil. Continue to use the original statement.
	if callbackErr != ErrReturnLater {
		stmt = retStmt
	}

	switch rv := stmt.(type) {
	case VarDecl:
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Argument:
		// NOOP, no children
	case Return:
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Assignment:
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Call:
		for i := range rv.Args {
			rv.Args[i] = tw.walkExpression(rv.Args[i], wctx)
		}
		stmt = rv
	case If:
		rv.Cond = tw.walkExpression(rv.Cond, wctx)
		for i := range rv.Then {
			rv.Then[i] = tw.walkStatement(rv.Then[i], wctx)
		}
		for i := range rv.Else {
			rv.Else[i] = tw.walkStatement(rv.Else[i], wctx)
		}
		stmt = rv
	case While:
		rv.Cond = tw.walkExpression(rv.Cond, wctx)
		for i := range rv.Body {
			rv.Body[i] = tw.walkStatement(rv.Body[i], wctx)
		}
		stmt = rv
	case FnDef:
		// Register the function in the scope before walking its body to allow recursive calls.
		must(wctx.Scope.RegisterFn(rv))
		functionScope := NewScope(wctx.Scope)

		// Walk the function body statements
		i := 0
		for i < len(rv.Args) {
			// walk the args
			retVal := tw.walkStatement(rv.Args[i], WalkingContext{
				Scope:    functionScope,
				ParentFn: &rv, // Set the parent function to the current function
			})
			if retVal == nil {
				// If the callback returns nil, we remove the argument.
				rv.Args = slices.Delete(rv.Args, i, i+1)
				continue
			}
			rv.Args[i] = retVal.(Argument) // TODO: Check the type

			i++
		}

		for i := range rv.Body {
			rv.Body[i] = tw.walkStatement(rv.Body[i], WalkingContext{
				Scope:    functionScope,
				ParentFn: &rv, // Set the parent function to the current function
			})
		}

		stmt = rv
	}

	if callbackErr == ErrReturnLater {
		// A lot of annoying shit happens if we were to do a complete walk of the node again.
		// Just force the api to work as such that the children is EITHER walked last by default OR walked before by opting in using ErrReturnLater, not both.

		retStmt, err := tw.CallbackFn(wctx, stmt)
		if err != nil {
			if err == ErrReturnLater {
				panic("ErrReturnLater returned twice for the same node")
			}
			panic(err)
		}
		stmt = retStmt
	}

	switch rv := stmt.(type) {
	case VarDecl:
		must(wctx.Scope.RegisterVar(stmt.(VarDecl)))
	case Argument:
		must(wctx.Scope.RegisterVar(VarDecl{
			Name: rv.Name,
			Type: rv.Type,
		}))
	case FnDef:
		wctx.Scope.deregisterFn(rv.Name)
		must(wctx.Scope.RegisterFn(rv))
	}

	return stmt
}

func (tw *TreeWalker) walkExpression(expr Expression, wctx WalkingContext) Expression {
	retVal, callbackErr := tw.CallbackFn(wctx, expr)
	if callbackErr != nil && callbackErr != ErrReturnLater {
		panic(callbackErr)
	}

	retExpr, ok := retVal.(Expression)
	if !ok {
		panic(fmt.Errorf("non-expression returned when walking expression: %T", retVal))
	}

	if callbackErr != ErrReturnLater {
		expr = retExpr
	}

	switch rv := (expr).(type) {
	case Call:
		for i := range rv.Args {
			rv.Args[i] = tw.walkExpression(rv.Args[i], wctx)
		}
		expr = rv
	case BinaryExpression:
		rv.Left = tw.walkExpression(rv.Left, wctx)
		rv.Right = tw.walkExpression(rv.Right, wctx)
		expr = rv
	case UnaryExpression:
		rv.Expression = tw.walkExpression(rv.Expression, wctx)
		expr = rv
	case IntegerLiteral,
		BooleanLiteral,
		VoidLiteral,
		VarIdentifier:
		// Nothing to walk
	default:
		panic(fmt.Errorf("unhandled expression type: %T", rv))
	}

	if callbackErr == ErrReturnLater {
		retVal, err := tw.CallbackFn(wctx, expr)
		if err != nil {
			if err == ErrReturnLater {
				panic("ErrReturnLater returned twice for the same node")
			}
			panic(err)
		}

		retExpr, ok := retVal.(Expression)
		if !ok {
			panic(fmt.Errorf("non-expression returned when walking expression: %T", retVal))
		}

		expr = retExpr
	}

	return expr
}
