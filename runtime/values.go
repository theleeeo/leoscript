package runtime

import "leoscript/types"

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
