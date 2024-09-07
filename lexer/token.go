package lexer

type Token interface{}

type BinaryToken struct {
	Operation string
}

type IntegerToken struct {
	Value int
}

type OpenParenToken struct{}

type CloseParenToken struct{}

type SemicolonToken struct{}
