package lexer

type TokenType string

type Token interface {
	token()
}

type BinaryToken struct {
	Operation string
}

func (BinaryToken) token() {}

type IntegerToken struct {
	Value int
}
