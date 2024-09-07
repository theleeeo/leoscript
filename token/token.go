package token

type Token interface{}

type Binary struct {
	Operation string
}

type Integer struct {
	Value int
}

type OpenParen struct{}

type CloseParen struct{}

type Semicolon struct{}
