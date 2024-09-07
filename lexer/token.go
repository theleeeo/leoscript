package lexer

type Token interface{}

type BinaryToken struct {
	Operation string
}

type IntegerToken struct {
	Value int
}
