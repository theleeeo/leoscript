package runtime

import "leoscript/types"

func toRuntimeVal(val any) runtimeVal {
	switch v := val.(type) {
	case int:
		return numberVal{value: v}
	case bool:
		return booleanVal{value: v}
	default:
		panic("unsupported type for runtime value conversion")
	}
}

func fromRuntimeVal(val runtimeVal) any {
	switch v := val.(type) {
	case numberVal:
		return v.value
	case booleanVal:
		return v.value
	default:
		panic("unsupported type for runtime value conversion")
	}
}

type runtimeVal interface {
	Type() types.Type
}

type numberVal struct {
	value int
}

func (numberVal) Type() types.Type { return types.Int }

type booleanVal struct {
	value bool
}

func (booleanVal) Type() types.Type { return types.Bool }
