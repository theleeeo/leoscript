package parser

import (
	"leoscript/types"
)

func defaultValue(t types.Type) Expression {
	switch t.Kind() {
	case types.KindInt:
		return IntegerLiteral{Value: 0}
	case types.KindBool:
		return BooleanLiteral{Value: false}
	case types.KindString:
		return StringLiteral{Value: ""}
	case types.KindStruct:
		st := t.(*types.Struct)
		sl := StructLiteral{
			Type:   st,
			Fields: make([]FieldLiteral, len(st.Fields)),
		}
		for i, field := range st.Fields {
			sl.Fields[i] = FieldLiteral{
				Name:  field.Name,
				Value: defaultValue(field.Type),
			}
		}
		return sl
	default:
		panic("unsupported type: " + t.Kind().String())
	}
}

// Checks for uninitialized variables and adds a literal representing the default value.
func initializeDefaultPass(program *Program) (err error) {
	tw := NewTreeWalker(TreeWalkerConfig{
		CallbackFn: func(wctx WalkingContext, node Statement) (Statement, error) {
			switch n := node.(type) {
			case VarDecl:
				if n.Value != nil {
					// Variable is initialized, no action needed
					return node, nil
				}

				// Variable is uninitialized, set to default value
				n.Value = defaultValue(n.Type)
				return n, nil
			}

			return node, nil
		},
	})

	return tw.WalkProgram(program)
}
