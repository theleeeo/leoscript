package token

type Token interface{}

type Binary struct {
	Operation string
}

type Integer struct {
	Value int
}

type OpenParen struct{}

func (t OpenParen) String() string {
	return "{(}"
}

type CloseParen struct{}

func (t CloseParen) String() string {
	return "{)}"
}

type Semicolon struct{}

func (t Semicolon) String() string {
	return "{;}"
}
