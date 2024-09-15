package token

type Token interface {
	Type() TokenType
}

type TokenType int

const (
	EOFType TokenType = iota
	IntegerType
	OperatorType
	OpenParenType
	CloseParenType
	SemicolonType
	IdentifierType
	BooleanType
	VarDefType
	IntDefType
	BoolDefType
)

type EOF struct{}

func (EOF) Type() TokenType { return EOFType }

type Integer struct {
	Value int
}

func (Integer) Type() TokenType { return IntegerType }

type Operator struct {
	Op string
}

func (Operator) Type() TokenType { return OperatorType }

func (t Operator) Priority() Priority {
	switch t.Op {
	case "=":
		return PRIO_ASSIGN
	case "==", "!=":
		return PRIO_EQUALS
	case "<", ">", "<=", ">=":
		return PRIO_COMPARISON
	case "&&":
		return PRIO_AND
	case "||":
		return PRIO_OR
	case "+", "-":
		return PRIO_SUM
	case "*", "/":
		return PRIO_PRODUCT
	}

	panic("invalid operator in binary expression")
}

type OpenParen struct{}

func (OpenParen) String() string {
	return "{(}"
}

func (OpenParen) Type() TokenType { return OpenParenType }

type CloseParen struct{}

func (CloseParen) String() string {
	return "{)}"
}

func (CloseParen) Type() TokenType { return CloseParenType }

type Semicolon struct{}

func (Semicolon) String() string {
	return "{;}"
}

func (Semicolon) Type() TokenType { return SemicolonType }

type Identifier struct {
	Value string
}

func (Identifier) Type() TokenType { return IdentifierType }

type Boolean struct {
	Value bool
}

func (Boolean) Type() TokenType { return BooleanType }

type VarDef struct{}

func (VarDef) Type() TokenType { return VarDefType }

type IntDef struct{}

func (IntDef) Type() TokenType { return IntDefType }

type BoolDef struct{}

func (BoolDef) Type() TokenType { return BoolDefType }
