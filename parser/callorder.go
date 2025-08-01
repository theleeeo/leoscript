package parser

import "fmt"

func buildCallOrder(vars []VarDecl) ([]VarDecl, error) {
	type coVar struct {
		Var       VarDecl
		DependsOn []string
	}

	unresolvedVars := make(map[string]coVar)

	for _, v := range vars {
		dependantVars := getDependantVars(v.Value)
		unresolvedVars[v.Name] = coVar{
			Var:       v,
			DependsOn: dependantVars,
		}
	}

	callOrder := make([]VarDecl, 0, len(vars))

	visited := make(map[string]struct{})
	var visit func(name string) error
	visit = func(name string) error {
		if _, ok := visited[name]; ok {
			// If the variable has been visited but not resolved, it means it's a circular dependency.
			if _, ok := unresolvedVars[name]; ok {
				return fmt.Errorf("circular dependency detected: %s", name)
			}
			return nil // Already visited
		}
		visited[name] = struct{}{}
		for _, dep := range unresolvedVars[name].DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		callOrder = append(callOrder, unresolvedVars[name].Var)
		delete(unresolvedVars, name)
		return nil
	}

	for _, v := range vars {
		if _, ok := visited[v.Name]; !ok {
			if err := visit(v.Name); err != nil {
				return nil, err
			}
		}
	}

	if len(callOrder) != len(vars) {
		panic("call order length does not match number of variables") // Sanity check, i do not know how this could happen.
	}

	return callOrder, nil
}

func getDependantVars(expr Expression) []string {
	dependantVars := make([]string, 0)

	switch e := expr.(type) {
	case VarIdentifier:
		dependantVars = append(dependantVars, e.Name)
	case BinaryExpression:
		dependantVars = append(dependantVars, getDependantVars(e.Left)...)
		dependantVars = append(dependantVars, getDependantVars(e.Right)...)
	case UnaryExpression:
		dependantVars = append(dependantVars, getDependantVars(e.Expression)...)
	}

	return dependantVars
}
