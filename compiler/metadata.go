package compiler

import (
	"encoding/binary"
	"errors"
	"fmt"
	"leoscript/types"
)

type Metadata struct {
	// The exported functions of the program
	functions []exportedFunction
	// The exported variables of the program
	variables []ExportedVariable
}

func (md Metadata) Marshal() []byte {
	// Serialize the function metadata
	fnRaw := binary.BigEndian.AppendUint64(nil, uint64(len(md.functions)))
	for _, fn := range md.functions {
		fnRaw = binary.BigEndian.AppendUint64(fnRaw, uint64(len(fn.name)))
		fnRaw = append(fnRaw, []byte(fn.name)...)

		fnRaw = binary.BigEndian.AppendUint64(fnRaw, fn.startOffset)

		// Encode flags: bit 0 = exported, bit 1 = stub
		var flags byte
		if fn.exported {
			flags |= 1 << 0
		}
		if fn.stub {
			flags |= 1 << 1
		}
		fnRaw = append(fnRaw, flags)

		fnRaw = binary.BigEndian.AppendUint64(fnRaw, uint64(len(fn.args)))
		for _, arg := range fn.args {
			fnRaw = binary.BigEndian.AppendUint64(fnRaw, uint64(arg.argType.(types.BasicType)))
		}

		fnRaw = binary.BigEndian.AppendUint64(fnRaw, uint64(fn.returnType.(types.BasicType)))
	}

	varRaw := binary.BigEndian.AppendUint64(nil, uint64(len(md.variables)))
	for _, varDecl := range md.variables {
		varRaw = binary.BigEndian.AppendUint64(varRaw, uint64(len(varDecl.Name)))
		varRaw = append(varRaw, []byte(varDecl.Name)...)

		varRaw = binary.BigEndian.AppendUint64(varRaw, varDecl.Offset)
		varRaw = binary.BigEndian.AppendUint64(varRaw, uint64(varDecl.VarType.(types.BasicType)))
	}

	mdLen := binary.BigEndian.AppendUint64(nil, uint64(len(fnRaw)+len(varRaw)))
	return append(mdLen, append(fnRaw, varRaw...)...)
}

func (md *Metadata) Unmarshal(raw []byte) error {
	if len(raw) < 8 {
		return errors.New("metadata raw is too short")
	}

	mdSize := binary.BigEndian.Uint64(raw[:8])
	if len(raw) < int(mdSize)+8 {
		return errors.New("metadata raw is too short")
	}
	raw = raw[8 : mdSize+8]

	raw, err := md.unmarshalFunctions(raw)
	if err != nil {
		return fmt.Errorf("unmarshalling functions: %w", err)
	}

	raw, err = md.unmarshalVariables(raw)
	if err != nil {
		return fmt.Errorf("unmarshalling variables: %w", err)
	}

	if len(raw) != 0 {
		return fmt.Errorf("metadata raw has %d bytes left, expected 0", len(raw))
	}

	return nil
}

func (md *Metadata) unmarshalFunctions(raw []byte) (remaining []byte, err error) {
	if len(raw) < 8 {
		return nil, errors.New("metadata raw is too short for functions")
	}

	fnCount := binary.BigEndian.Uint64(raw[:8])
	raw = raw[8:]
	md.functions = make([]exportedFunction, fnCount)
	for i := uint64(0); i < fnCount; i++ {
		if len(raw) < 8 {
			return nil, errors.New("metadata raw is too short for function name length")
		}
		nameLen := binary.BigEndian.Uint64(raw[:8])
		raw = raw[8:]
		if len(raw) < int(nameLen) {
			return nil, errors.New("metadata raw is too short for function name")
		}
		name := string(raw[:nameLen])
		raw = raw[nameLen:]

		if len(raw) < 16 {
			return nil, errors.New("metadata raw is too short for function start offset")
		}
		startOffset := binary.BigEndian.Uint64(raw[:8])
		raw = raw[8:]

		if len(raw) < 1 {
			return nil, errors.New("metadata raw is too short for function flags")
		}
		flags := raw[0]
		raw = raw[1:]

		if len(raw) < 8 {
			return nil, errors.New("metadata raw is too short for argument count")
		}
		argCount := binary.BigEndian.Uint64(raw[:8])
		raw = raw[8:]
		args := make([]fnArg, argCount)
		for j := uint64(0); j < argCount; j++ {
			if len(raw) < 8 {
				return nil, errors.New("metadata raw is too short for argument type")
			}
			argType := types.BasicType(binary.BigEndian.Uint64(raw[:8]))
			raw = raw[8:]
			args[j] = fnArg{argType: argType}
		}

		if len(raw) < 8 {
			return nil, errors.New("metadata raw is too short for return type")
		}
		returnType := types.BasicType(binary.BigEndian.Uint64(raw[:8]))
		raw = raw[8:]

		md.functions[i] = exportedFunction{
			name:        name,
			startOffset: startOffset,
			args:        args,
			returnType:  returnType,
			exported:    flags&0x01 != 0, // bit 0
			stub:        flags&0x02 != 0, // bit 1
		}
	}

	return raw, nil
}

func (md *Metadata) unmarshalVariables(raw []byte) (remaining []byte, err error) {
	if len(raw) < 8 {
		return nil, errors.New("metadata raw is too short for variables")
	}

	varCount := binary.BigEndian.Uint64(raw[:8])
	raw = raw[8:]
	md.variables = make([]ExportedVariable, varCount)
	for i := uint64(0); i < varCount; i++ {
		if len(raw) < 8 {
			return nil, errors.New("metadata raw is too short for variable name length")
		}
		nameLen := binary.BigEndian.Uint64(raw[:8])
		raw = raw[8:]
		if len(raw) < int(nameLen) {
			return nil, errors.New("metadata raw is too short for variable name")
		}
		name := string(raw[:nameLen])
		raw = raw[nameLen:]

		if len(raw) < 16 {
			return nil, errors.New("metadata raw is too short for variable offset")
		}
		offset := binary.BigEndian.Uint64(raw[:8])
		raw = raw[8:]

		if len(raw) < 8 {
			return nil, errors.New("metadata raw is too short for variable type")
		}
		varType := types.BasicType(binary.BigEndian.Uint64(raw[:8]))
		raw = raw[8:]

		md.variables[i] = ExportedVariable{
			Name:    name,
			Offset:  offset,
			VarType: varType,
		}
	}

	return raw, nil
}

type fnArg struct {
	// The type of the function argument
	argType types.Type
}

type exportedFunction struct {
	name string
	// The offset in the code where the function starts
	startOffset uint64
	// The arguments of the function
	args []fnArg
	// The return type of the function
	returnType types.Type
	// Indicates if the function is exported
	exported bool
	// Indicates if the function is a stub
	stub bool
}

type ExportedVariable struct {
	Name string
	// The Offset in the variable stack where the variable will be stored
	Offset uint64
	// The type of the variable
	VarType types.Type
}

// func (md *Metadata) Functions() []exportedFunction {
// 	var exportedFunctions []exportedFunction
// 	for _, fn := range md.functions {
// 		if fn.exported {
// 			exportedFunctions = append(exportedFunctions, fn)
// 		}
// 	}
// 	return exportedFunctions
// }

func (md *Metadata) Variables() []ExportedVariable {
	return md.variables
}
