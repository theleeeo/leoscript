package parser

import (
	"errors"
	"fmt"
	"leoscript/types"
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

	universeScope = NewScope(nil) // The universe scope contains all built-in functions and types. It is the root of all scopes.
)

func init() {
	universeScope.RegisterType("int", types.Int)
	universeScope.RegisterType("bool", types.Bool)
	universeScope.RegisterType("string", types.String)
}

type WalkingContext struct {
	Scope *Scope

	// The function whose body we are currently walking.
	// Will be nil if we are not in a function body (only in the global scope).
	ParentFn *FnDef

	ParentNode Statement // TODO: There should maybe be a common "Node" type encasing both statement and expression
	// Will be set to the index the current node is in its parent.
	// Only relevant for node types that have lists of children, like a function call, array literal, etc.
	// Will not be set for statements in a block or parameters in a function definition.
	// IndexInParentNode int
}

type TreeWalker struct {
	CallbackFn func(wctx WalkingContext, node Statement) (Statement, error)
	BeforeWalk func(pg *Program, globalScope *Scope) error
}

type TreeWalkerConfig struct {
	BeforeWalk func(pg *Program, globalScope *Scope) error
	CallbackFn func(wctx WalkingContext, node Statement) (Statement, error)
}

// NewTreeWalker creates a new TreeWalker with the provided callbacks.
func NewTreeWalker(config TreeWalkerConfig) *TreeWalker {
	return &TreeWalker{
		CallbackFn: config.CallbackFn,
		BeforeWalk: config.BeforeWalk,
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

	globalScope := NewScope(universeScope)

	// Setup the global scope with the global variable declarations and function definitions.
	for _, structDef := range program.Structs {
		must(globalScope.RegisterType(structDef.Name, structDef))
	}

	for _, varDecl := range program.VarDecls {
		must(globalScope.RegisterVar(varDecl))
	}

	for _, fn := range program.FnDefs {
		must(globalScope.RegisterFn(fn))
	}

	for _, stub := range program.StubDefs {
		must(globalScope.RegisterFn(stub))
	}

	if tw.BeforeWalk != nil {
		if err := tw.BeforeWalk(program, globalScope); err != nil {
			return fmt.Errorf("before walk: %w", err)
		}
	}

	i := 0
	for i < len(program.VarDecls) {
		globalScope.deregisterVar(program.VarDecls[i].Name)

		resp := tw.walkStatement(program.VarDecls[i], WalkingContext{
			Scope:      globalScope,
			ParentFn:   nil, // No parent function in the global scope
			ParentNode: nil, // TODO: What do? There is no parent node but like... Is this fine?
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
			Scope:      globalScope,
			ParentFn:   &program.FnDefs[i],
			ParentNode: nil, // TODO: What do? There is no parent node but like... Is this fine?
		})
		if resp == nil {
			// If the callback returns nil, we remove the function definition.
			program.FnDefs = slices.Delete(program.FnDefs, i, i+1)
			continue
		}
		program.FnDefs[i] = resp.(FnDef)

		i++
	}

	i = 0
	for i < len(program.StubDefs) {
		globalScope.deregisterFn(program.StubDefs[i].Name)

		resp := tw.walkStatement(program.StubDefs[i], WalkingContext{
			Scope:      globalScope,
			ParentFn:   nil, // No parent function in the global scope
			ParentNode: nil, // TODO: What do? There is no parent node but like... Is this fine?
		})
		if resp == nil {
			// If the callback returns nil, we remove the function definition.
			program.StubDefs = slices.Delete(program.StubDefs, i, i+1)
			continue
		}
		program.StubDefs[i] = resp.(FnDef)

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
		// Uninitialized variables do not have a value.
		if rv.Value != nil {
			wctx.ParentNode = rv
			rv.Value = tw.walkExpression(rv.Value, wctx)
		}

		stmt = rv
	case Parameter:
		// NOOP, no children
	case Return:
		wctx.ParentNode = rv
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Assignment:
		wctx.ParentNode = rv
		rv.Value = tw.walkExpression(rv.Value, wctx)
		stmt = rv
	case Call:
		wctx.ParentNode = rv
		for i := range rv.Args {
			rv.Args[i] = tw.walkExpression(rv.Args[i], wctx)
		}
		stmt = rv
	case If:
		wctx.ParentNode = rv
		rv.Cond = tw.walkExpression(rv.Cond, wctx)
		for i := range rv.Then {
			rv.Then[i] = tw.walkStatement(rv.Then[i], wctx)
		}
		for i := range rv.Else {
			rv.Else[i] = tw.walkStatement(rv.Else[i], wctx)
		}
		stmt = rv
	case While:
		wctx.ParentNode = rv
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
		for i < len(rv.Params) {
			// walk the args
			retVal := tw.walkStatement(rv.Params[i], WalkingContext{
				Scope:      functionScope,
				ParentFn:   &rv, // Set the parent function to the current function
				ParentNode: nil, // The parent node is the function present in ParentFn
			})
			if retVal == nil {
				// If the callback returns nil, we remove the parameter.
				rv.Params = slices.Delete(rv.Params, i, i+1)
				continue
			}
			rv.Params[i] = retVal.(Parameter) // TODO: Check the type

			i++
		}

		for i := range rv.Body {
			rv.Body[i] = tw.walkStatement(rv.Body[i], WalkingContext{
				Scope:      functionScope,
				ParentFn:   &rv, // Set the parent function to the current function
				ParentNode: nil, // The parent node is the function present in ParentFn
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
	case Parameter:
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
	// case StructLiteral: // TODO
	// 	for i := range rv.Fields {
	// 		retExpr := tw.walkExpression(rv.Fields[i], wctx)
	// 		if _, ok := retExpr.(FieldLiteral); !ok {
	// 			panic(fmt.Errorf("non-field literal returned when walking struct field: %T", retExpr))
	// 		}
	// 		rv.Fields[i] = retExpr.(FieldLiteral)
	// 	}
	// 	expr = rv
	case IntegerLiteral,
		BooleanLiteral,
		VoidLiteral,
		VarIdentifier,
		StringLiteral,
		StructLiteral:
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
