package types

type Type interface {
	isType()
}

type BasicType int

func (b BasicType) isType() {}

const (
	_ BasicType = iota

	Bool
	Int
)
