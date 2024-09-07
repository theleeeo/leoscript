package parser

type Expression interface {
	isExpression()
}

type expression struct{}

func (e expression) isExpression() {}

type IntegerLiteral struct {
	expression
	Value int
}

// type BinaryExpression struct {
// 	expression
// 	Left  Expression
// 	Right Expression
// 	Op    string
// }
