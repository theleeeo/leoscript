package types

import "strconv"

type Type interface {
	isType()
	Size() uint64
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BasicType

type BasicType int

func (b BasicType) isType() {}

func (b BasicType) Size() uint64 {
	switch b {
	case Void:
		return 0
	case Bool:
		return 1
	case Int:
		return 8 // Assuming 64-bit integers
	}

	panic("unhandled basic type: " + strconv.Itoa(int(b)))
}

const (
	Unspecified BasicType = iota

	// No type. Used for void functions
	Void

	Bool
	Int
)
