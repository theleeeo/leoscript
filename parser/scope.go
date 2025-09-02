package parser

import (
	"fmt"
	"leoscript/types"
	"slices"
	"strings"
)

type Scope struct {
	parent *Scope

	fnDefs   map[string]FnDef
	varDecls map[string]VarDecl
	types    map[string]types.Type
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent: parent,
		fnDefs:   make(map[string]FnDef),
		varDecls: make(map[string]VarDecl),
		types:    make(map[string]types.Type),
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

func (s *Scope) ResolveVarType(name string) (types.Type, bool) {
	varName, fieldSelection, _ := strings.Cut(name, ".")

	varDecl, ok := s.varDecls[varName]
	if !ok && s.parent != nil {
		return s.parent.ResolveVarType(name)
	}

	if fieldSelection != "" {
		varStruct, ok := varDecl.Type.(*types.Struct)
		if !ok {
			panic(fmt.Sprint("variable is not a struct: ", varDecl.Name))
		}

		fieldIndex := slices.IndexFunc(varStruct.Fields, func(f types.Field) bool {
			return f.Name == fieldSelection
		})

		if fieldIndex == -1 {
			panic(fmt.Sprint("unknown field:", fieldSelection))
		}

		return varStruct.Fields[fieldIndex].Type, ok
	}

	return varDecl.Type, ok
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

func (s *Scope) ResolveType(name string) (types.Type, bool) {
	typeDecl, ok := s.types[name]
	if !ok && s.parent != nil {
		return s.parent.ResolveType(name)
	}

	return typeDecl, ok
}

func (s *Scope) RegisterType(name string, typeDecl types.Type) error {
	if _, ok := s.types[name]; ok {
		return fmt.Errorf("type with name \"%s\" already declared", name)
	}

	s.types[name] = typeDecl
	return nil
}
