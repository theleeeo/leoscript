package token

type Token interface {
	Type() TokenType
}

type TokenType string

const (
	EOFType        TokenType = "EOF"
	IntegerType    TokenType = "INTEGER"
	BinaryType     TokenType = "BINARY"
	OpenParenType  TokenType = "OPEN_PAREN"
	CloseParenType TokenType = "CLOSE_PAREN"
	SemicolonType  TokenType = "SEMICOLON"
)

type EOF struct{}

func (t EOF) Type() TokenType { return EOFType }

type Integer struct {
	Value int
}

func (t Integer) Type() TokenType { return IntegerType }

type Binary struct {
	Operation string
}

func (t Binary) Type() TokenType { return BinaryType }

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
