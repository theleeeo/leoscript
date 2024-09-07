package parser

type Expression interface {
	isExpression()
}

// type expression struct{}

// func (e expression) isExpression() {}

type IntegerLiteral struct {
	// expression
	Value int
}

func (i IntegerLiteral) isExpression() {}

type BinaryExpression struct {
	// expression
	Left     Expression
	Right    Expression
	Op       string
	Priority int
}

func (b BinaryExpression) isExpression() {}
