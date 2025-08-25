package types

import "strconv"

////go:generate stringer -type=Kind
// type Kind int

// const (
// 	KindInvalid Kind = iota
// 	KindInt
// 	KindBool
// 	KindString
// )

type Type interface {
	isType()
	Size() uint64
	// Kind() Kind
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BasicType
type BasicType int

const (
	Unspecified BasicType = iota

	// No type. Used for void functions
	Void

	Bool
	Int
	String
)

func (b BasicType) isType() {}

func (b BasicType) Size() uint64 {
	switch b {
	case Void:
		return 0
	case Bool:
		return 1
	case Int:
		return 8
	case String:
		return 16
	}

	panic("unhandled basic type: " + strconv.Itoa(int(b)))
}

// func (b BasicType) Kind() Kind {
// 	switch b {
// 	case Void:
// 		return KindInvalid
// 	case Bool:
// 		return KindBool
// 	case Int:
// 		return KindInt
// 	}

// 	panic("unhandled basic type: " + strconv.Itoa(int(b)))
// }
