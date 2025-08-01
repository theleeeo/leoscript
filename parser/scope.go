package parser

import "fmt"

type Scope struct {
	parent *Scope

	fnDefs   map[string]FnDef
	varDecls map[string]VarDecl
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent: parent,
		fnDefs:   make(map[string]FnDef),
		varDecls: make(map[string]VarDecl),
	}
}

func (s *Scope) ResolveFn(name string) (FnDef, bool) {
	fnDef, ok := s.fnDefs[name]
	if !ok && s.parent != nil {
		return s.parent.ResolveFn(name)
	}

	return fnDef, ok
}

func (s *Scope) RegisterFn(fnDef FnDef) error {
	if _, ok := s.fnDefs[fnDef.Name]; ok {
		return fmt.Errorf("function %s already declared", fnDef.Name)
	}

	s.fnDefs[fnDef.Name] = fnDef
	return nil
}

func (s *Scope) deregisterFn(name string) {
	if _, ok := s.fnDefs[name]; !ok {
		panic(fmt.Sprintf("function %s not declared", name))
	}

	delete(s.fnDefs, name)
}

func (s *Scope) ResolveVar(name string) (VarDecl, bool) {
	varDecl, ok := s.varDecls[name]
	if !ok && s.parent != nil {
		return s.parent.ResolveVar(name)
	}

	return varDecl, ok
}

func (s *Scope) RegisterVar(varDecl VarDecl) error {
	if _, ok := s.varDecls[varDecl.Name]; ok {
		return fmt.Errorf("variable %s already declared", varDecl.Name)
	}

	s.varDecls[varDecl.Name] = varDecl
	return nil
}

func (s *Scope) deregisterVar(name string) {
	if _, ok := s.varDecls[name]; !ok {
		panic(fmt.Sprintf("variable %s is not declared", name)) // This function is for internal use only and this should never
	}

	delete(s.varDecls, name)
}
