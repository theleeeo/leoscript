// Code generated by "stringer -type=BasicType"; DO NOT EDIT.

package types

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Void-1]
	_ = x[Bool-2]
	_ = x[Int-3]
}

const _BasicType_name = "VoidBoolInt"

var _BasicType_index = [...]uint8{0, 4, 8, 11}

func (i BasicType) String() string {
	i -= 1
	if i < 0 || i >= BasicType(len(_BasicType_index)-1) {
		return "BasicType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _BasicType_name[_BasicType_index[i]:_BasicType_index[i+1]]
}
