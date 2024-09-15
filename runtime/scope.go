package runtime

import "fmt"

type scope struct {
	parent *scope

	variables map[string]runtimeVal
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, variables: make(map[string]runtimeVal)}
}

func (s *scope) GetVar(name string) (runtimeVal, bool) {
	val, ok := s.variables[name]
	if !ok && s.parent != nil {
		return s.parent.GetVar(name)
	}

	return val, ok
}

func (s *scope) DeclareVar(name string, val runtimeVal) error {
	if _, ok := s.variables[name]; ok {
		return fmt.Errorf("variable %s already declared", name)
	}

	s.variables[name] = val
	return nil
}

func (s *scope) SetVar(name string, val runtimeVal) error {
	if _, ok := s.variables[name]; !ok {
		if s.parent != nil {
			return s.parent.SetVar(name, val)
		}

		return fmt.Errorf("variable %s not declared", name)
	}

	s.variables[name] = val
	return nil
}
