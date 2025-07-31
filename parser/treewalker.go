package parser

import (
	"fmt"
	"slices"
)

// TODO: You are better than this
func must(err error) {
	if err != nil {
		panic(err)
	}
}

type WalkingContext struct {
	Scope *Scope

	// The function whose body we are currently walking.
	// Will be nil if we are not in a function body (only in the global scope).
	ParentFn *FnDef
}

type TreeWalker struct {
	CallbackFn func(wctx WalkingContext, node Statement) Statement
}

// NewTreeWalker creates a new TreeWalker with the provided callbacks.
func NewTreeWalker(callbackFn func(wctx WalkingContext, node Statement) Statement) *TreeWalker {
	return &TreeWalker{
		CallbackFn: callbackFn,
	}
}

// WalkProgram walks through the program and applies the refCheck to each statement and expression.
//
// TODO: Do a check and avoid walking nil statements or expressions.
// If a statement or expression is nil, it should be removed from the ast by some cleaning pass in the end.
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

	i := 0
	for i < len(program.VarDecls) {
		must(globalScope.DeregisterVar(program.VarDecls[i]))

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
		must(globalScope.DeregisterFn(program.FnDefs[i]))

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
	switch rv := stmt.(type) {
	case VarDecl:
		rv.Value = tw.walkExpression(rv.Value, wctx)

		// If the statement is a variable declaration, we register it in the scope.
		must(wctx.Scope.RegisterVar(rv))
		stmt = rv
	case Argument:
		// If the statement is an argument, we register it in the scope.
		must(wctx.Scope.RegisterVar(VarDecl{
			Name: rv.Name,
			Type: rv.Type,
		}))
	case Return:
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Assignment:
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Call:
		for i := range rv.Args {
			fmt.Printf("Walking argument %d: %T\n", i, rv.Args[i])
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
		// Walk the function body statements
		functionScope := NewScope(wctx.Scope)

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

		// If the statement is a function definition, we register it in the scope.
		must(wctx.Scope.RegisterFn(rv))
		stmt = rv
	}

	return tw.CallbackFn(wctx, stmt)
}

func (tw *TreeWalker) walkExpression(expr Expression, wctx WalkingContext) Expression {
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

	retVal := tw.CallbackFn(wctx, expr)
	if newExpr, ok := retVal.(Expression); ok {
		return newExpr
	}

	panic(fmt.Errorf("non-expression returned when walking expression: %T", retVal))
}
