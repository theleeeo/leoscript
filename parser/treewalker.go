package parser

import (
	"fmt"
	"slices"
)

type WalkingContext struct {
	Scope *Scope
}

type TreeWalkerCallbacks struct {
	// Statement callbacks
	VarDecl    func(wctx WalkingContext, varDecl *VarDecl) *VarDecl
	FnDefs     func(wctx WalkingContext, fnDef *FnDef) *FnDef
	StubDef    func(wctx WalkingContext, stubDef *StubDef) *StubDef
	Return     func(wctx WalkingContext, ret *Return) *Return
	Assignment func(wctx WalkingContext, assignment *Assignment) *Assignment
	If         func(wctx WalkingContext, ifStmt *If) *If
	While      func(wctx WalkingContext, whileStmt *While) *While

	// Expression callbacks
	VarIdentifier    func(wctx WalkingContext, varIdentifier *VarIdentifier) *VarIdentifier
	Call             func(wctx WalkingContext, call *Call) *Call
	BinaryExpression func(wctx WalkingContext, binaryExpr *BinaryExpression) *BinaryExpression
	UnaryExpression  func(wctx WalkingContext, unaryExpr *UnaryExpression) *UnaryExpression
}

type TreeWalker struct {
	Callbacks TreeWalkerCallbacks
}

// TODO: Remove
type NoopCallback struct{}

// Statement callbacks
func (n NoopCallback) VarDecl(wctx WalkingContext, varDecl *VarDecl) *VarDecl {
	return varDecl
}
func (n NoopCallback) FnDefs(wctx WalkingContext, fnDef *FnDef) *FnDef {
	return fnDef
}
func (n NoopCallback) StubDef(wctx WalkingContext, stubDef *StubDef) *StubDef {
	return stubDef
}
func (n NoopCallback) Return(wctx WalkingContext, ret *Return) *Return {
	return ret
}
func (n NoopCallback) Assignment(wctx WalkingContext, assignment *Assignment) *Assignment {
	return assignment
}
func (n NoopCallback) If(wctx WalkingContext, ifStmt *If) *If {
	return ifStmt
}
func (n NoopCallback) While(wctx WalkingContext, whileStmt *While) *While {
	return whileStmt
}

// Expression callbacks
func (n NoopCallback) VarIdentifier(wctx WalkingContext, varIdentifier *VarIdentifier) *VarIdentifier {
	return varIdentifier
}
func (n NoopCallback) Call(wctx WalkingContext, call *Call) *Call {
	return call
}
func (n NoopCallback) BinaryExpression(wctx WalkingContext, binaryExpr *BinaryExpression) *BinaryExpression {
	return binaryExpr
}
func (n NoopCallback) UnaryExpression(wctx WalkingContext, unaryExpr *UnaryExpression) *UnaryExpression {
	return unaryExpr
}

// TODO: Reevaluate the node if it has changed after the callback.
// NewTreeWalker creates a new TreeWalker with the provided callbacks.
func NewTreeWalker(callbacks TreeWalkerCallbacks) *TreeWalker {
	if callbacks.VarDecl == nil {
		callbacks.VarDecl = NoopCallback{}.VarDecl
	}
	if callbacks.FnDefs == nil {
		callbacks.FnDefs = NoopCallback{}.FnDefs
	}
	if callbacks.StubDef == nil {
		callbacks.StubDef = NoopCallback{}.StubDef
	}
	if callbacks.Return == nil {
		callbacks.Return = NoopCallback{}.Return
	}
	if callbacks.Assignment == nil {
		callbacks.Assignment = NoopCallback{}.Assignment
	}
	if callbacks.If == nil {
		callbacks.If = NoopCallback{}.If
	}
	if callbacks.While == nil {
		callbacks.While = NoopCallback{}.While
	}
	if callbacks.VarIdentifier == nil {
		callbacks.VarIdentifier = NoopCallback{}.VarIdentifier
	}
	if callbacks.Call == nil {
		callbacks.Call = NoopCallback{}.Call
	}
	if callbacks.BinaryExpression == nil {
		callbacks.BinaryExpression = NoopCallback{}.BinaryExpression
	}
	if callbacks.UnaryExpression == nil {
		callbacks.UnaryExpression = NoopCallback{}.UnaryExpression
	}

	return &TreeWalker{
		Callbacks: callbacks,
	}
}

// WalkProgram walks through the program and applies the refCheck to each statement and expression.
func (tw *TreeWalker) WalkProgram(program Program) Program {
	globalScope := NewScope(nil)

	// Setup the global scope with the global variable declarations and function definitions.
	for _, varDecl := range program.VarDecls {
		globalScope.RegisterVar(varDecl)
	}

	for _, fn := range program.FnDefs {
		globalScope.RegisterFn(fn)
	}

	i := 0
	for i < len(program.VarDecls) {
		retVal := tw.Callbacks.VarDecl(WalkingContext{
			Scope: globalScope,
		}, &program.VarDecls[i])
		if retVal == nil {
			// If the callback returns nil, we remove the variable declaration.
			// Do not increment i, as the next element has shifted into the current index.
			program.VarDecls = append(program.VarDecls[:i], program.VarDecls[i+1:]...)
			continue
		}

		program.VarDecls[i] = *retVal

		tw.walkExpression(&program.VarDecls[i].Value, WalkingContext{
			Scope: globalScope,
		})
		i++

	}

	i = 0
	for i < len(program.FnDefs) {
		functionScope := NewScope(globalScope)

		// Register function parameters in the function scope
		for _, param := range program.FnDefs[i].Args {
			functionScope.RegisterVar(VarDecl{
				Name: param.Name,
				Type: param.Type,
			})
		}

		// Register the function in the global scope
		retVal := tw.Callbacks.FnDefs(WalkingContext{
			Scope: functionScope,
		}, &program.FnDefs[i])
		if retVal == nil {
			// If the callback returns nil, we remove the function definition.
			program.FnDefs = slices.Delete(program.FnDefs, i, i+1)
			// Do not increment i, as the next element has shifted into the current index.
			continue
		}

		program.FnDefs[i] = *retVal

		// Walk the function body statements
		for j := range program.FnDefs[i].Body {
			tw.walkStatement(&program.FnDefs[i].Body[j], WalkingContext{
				Scope: functionScope,
			})
		}
		i++
	}

	return program
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// TODO: Handle a callback removing or altering statements and expressions.
func (tw *TreeWalker) walkStatement(stmtRef *Statement, wctx WalkingContext) {
	var retVal Statement
	var nextExprs []*Expression

	switch stmt := (*stmtRef).(type) {
	case VarDecl:
		rv := tw.Callbacks.VarDecl(wctx, &stmt)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &rv.Value)
	case Return:
		rv := tw.Callbacks.Return(wctx, &stmt)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &stmt.Value)

	case Assignment:
		rv := tw.Callbacks.Assignment(wctx, &stmt)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &rv.Value)

	case Call:
		rv := tw.Callbacks.Call(wctx, &stmt)
		if rv == nil {
			break
		}
		retVal = *rv

		for i := range stmt.Args {
			nextExprs = append(nextExprs, &rv.Args[i])
		}

	case If:
		rv := tw.Callbacks.If(wctx, &stmt)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &rv.Cond)
		for i := range rv.Then {
			tw.walkStatement(&rv.Then[i], wctx)
		}
		for i := range rv.Else {
			tw.walkStatement(&rv.Else[i], wctx)
		}
	case While:
		rv := tw.Callbacks.While(wctx, &stmt)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &rv.Cond)
		for i := range rv.Body {
			tw.walkStatement(&rv.Body[i], wctx)
		}
	default:
		panic(fmt.Errorf("unhandled statement type: %T", stmt))
	}

	if retVal == nil {
		return
	}

	*stmtRef = retVal

	if varDecl, ok := (*stmtRef).(VarDecl); ok {
		// If the statement is a variable declaration, we register it in the scope.
		must(wctx.Scope.RegisterVar(varDecl))
	} else if fnDef, ok := (*stmtRef).(FnDef); ok {
		// If the statement is a function definition, we register it in the scope.
		must(wctx.Scope.RegisterFn(fnDef))
	} else if stubDef, ok := (*stmtRef).(StubDef); ok {
		_ = stubDef // TODO

		// If the statement is a stub definition, we register it in the scope.
		// must(wctx.Scope.RegisterStub(stubDef))
	}

	for _, exprRef := range nextExprs {
		tw.walkExpression(exprRef, wctx)
	}
}

func (tw *TreeWalker) walkExpression(exprRef *Expression, wctx WalkingContext) {
	var retVal Expression
	var nextExprs []*Expression

	switch expr := (*exprRef).(type) {
	case VarIdentifier:
		rv := tw.Callbacks.VarIdentifier(wctx, &expr)
		if rv == nil {
			break
		}
		retVal = *rv

	case Call:
		rv := tw.Callbacks.Call(wctx, &expr)
		if rv == nil {
			break
		}
		retVal = *rv

		for i := range expr.Args {
			nextExprs = append(nextExprs, &rv.Args[i])
		}

	case BinaryExpression:
		rv := tw.Callbacks.BinaryExpression(wctx, &expr)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &rv.Left, &rv.Right)
	case UnaryExpression:
		rv := tw.Callbacks.UnaryExpression(wctx, &expr)
		if rv == nil {
			break
		}
		retVal = *rv

		nextExprs = append(nextExprs, &rv.Expression)
	case IntegerLiteral, BooleanLiteral, VoidLiteral:
		// No need to walk literals. (Maybe exept for strings in the future)
	default:
		panic(fmt.Errorf("unhandled expression type: %T", expr))
	}

	if retVal == nil {
		*exprRef = nil
		return
	}

	*exprRef = retVal

	for _, nextExprRef := range nextExprs {
		if nextExprRef != nil {
			tw.walkExpression(nextExprRef, wctx)
		}
	}
}
