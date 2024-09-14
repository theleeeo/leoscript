package token

type Token interface {
	Type() TokenType
}

type TokenType string

const (
	EOFType        TokenType = "EOF"
	IntegerType    TokenType = "INTEGER"
	MathOpType     TokenType = "MATH_OP"
	OpenParenType  TokenType = "OPEN_PAREN"
	CloseParenType TokenType = "CLOSE_PAREN"
	SemicolonType  TokenType = "SEMICOLON"
	IdentifierType TokenType = "IDENTIFIER"
	BooleanType    TokenType = "BOOLEAN"
)

type EOF struct{}

func (t EOF) Type() TokenType { return EOFType }

type Integer struct {
	Value int
}

func (t Integer) Type() TokenType { return IntegerType }

type MathOp struct {
	Operation string
}

func (t MathOp) Type() TokenType { return MathOpType }

type OpenParen struct{}

func (t OpenParen) String() string {
	return "{(}"
}

func (t OpenParen) Type() TokenType { return OpenParenType }

type CloseParen struct{}

func (t CloseParen) String() string {
	return "{)}"
}

func (t CloseParen) Type() TokenType { return CloseParenType }

type Semicolon struct{}

func (t Semicolon) String() string {
	return "{;}"
}

func (t Semicolon) Type() TokenType { return SemicolonType }

type Identifier struct {
	Value string
}

func (t Identifier) Type() TokenType { return IdentifierType }

type Boolean struct {
	Value bool
}

func (t Boolean) Type() TokenType { return BooleanType }
