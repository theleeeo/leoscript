package types

import (
	"strconv"
)

//go:generate stringer -type=Kind
type Kind int

const (
	KindInvalid Kind = iota
	KindBool
	KindInt
	KindString
	KindStruct
	KindArray
)

type Type interface {
	Size() uint64
	Kind() Kind
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BasicType
type BasicType int

const (
	Unspecified BasicType = iota + 1 // TODO: Maybe remove in favor of a more explicit "implicitVariable". Like how it is for unresolved type identifier but for vars/calls. It is however used for two things, implicit var and not-resolved-yet.

	// No type. Used for void functions
	Void

	Bool
	Int
	String
)

func (b BasicType) Size() uint64 {
	switch b {
	case Void:
		return 0
	case Bool:
		return 1
	case Int:
		return 1
	case String:
		return 2
	}

	panic("unhandled basic type: " + strconv.Itoa(int(b)))
}

func (b BasicType) Kind() Kind {
	switch b {
	case Unspecified:
		return KindInvalid
	case Void:
		return KindInvalid
	case Bool:
		return KindBool
	case Int:
		return KindInt
	case String:
		return KindString
	}

	panic("unhandled basic type: " + strconv.Itoa(int(b)))
}

type Struct struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type Type
}

func (s *Struct) Size() uint64 {
	var size uint64
	for _, field := range s.Fields {
		size += field.Type.Size()
	}
	return size
}

func (s *Struct) Kind() Kind {
	return KindStruct
}

type Array struct {
	ElementType Type
}

func (a Array) Size() uint64 {
	return 2
}

func (a Array) Kind() Kind {
	if a.ElementType.Kind() == KindInvalid {
		return KindInvalid
	}

	return KindArray
}
