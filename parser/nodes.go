package parser

type Expression interface {
	isExpression()
}

type IntegerLiteral struct {
	Value int
}

func (i IntegerLiteral) isExpression() {}

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Op       string
	Priority int
}

func (b BinaryExpression) isExpression() {}

type UnaryExpression struct {
	Expression Expression
	Op         string
}

func (u UnaryExpression) isExpression() {}
