package types

type Type interface {
	isType()
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BasicType

type BasicType int

func (b BasicType) isType() {}

const (
	_ BasicType = iota

	// No type. Used for void functions
	Void

	Bool
	Int
)
