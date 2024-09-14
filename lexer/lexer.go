package lexer

import (
	"fmt"
	"leoscript/token"
)

type lexer struct {
	input string
	pos   int

	tokens []token.Token
}

func (lx *lexer) next() byte {
	lx.pos++

	if lx.pos >= len(lx.input) {
		return 0
	}
	char := lx.input[lx.pos]
	return char
}

func (lx *lexer) putBack() {
	lx.pos--
}

func (lx *lexer) peek() byte {
	return lx.input[lx.pos]
}

func (lx *lexer) pushToken(tk token.Token) {
	lx.tokens = append(lx.tokens, tk)
}

func Tokenize(input string) ([]token.Token, error) {
	lx := lexer{input: input}

	for tk := lx.peek(); tk != 0; tk = lx.next() {
		switch tk {
		case ' ', '\n', '\t':
			// Skip whitespace
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			value := lx.parseInteger()
			lx.pushToken(token.Integer{Value: value})
		case '+', '-', '*', '/':
			lx.pushToken(token.Binary{Operation: string(tk)})
		case '(':
			lx.pushToken(token.OpenParen{})
		case ')':
			lx.pushToken(token.CloseParen{})
		case ';':
			lx.pushToken(token.Semicolon{})
		default:
			return nil, fmt.Errorf("invalid character: %c", tk)
		}
	}

	return lx.tokens, nil
}

func isNumeric(char byte) bool {
	return char >= '0' && char <= '9'
}

func (lx *lexer) parseInteger() int {
	value := int(lx.peek() - '0')

	for isNumeric(lx.next()) {
		value = value*10 + int(lx.peek()-'0')

	}

	// Put back the last character so that we do not return with the position past the bounds of what this funciton handled.
	lx.putBack()

	return value
}

// func isAlpha(char byte) bool {
// 	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
// }
