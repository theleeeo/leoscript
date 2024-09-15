package lexer

import (
	"fmt"
	"leoscript/token"
	"leoscript/types"
	"strings"
)

var keywords = map[string]token.Token{
	"true":   token.Boolean{Value: true},
	"false":  token.Boolean{Value: false},
	"var":    token.VarDecl{},
	"int":    token.Type{Kind: types.Int},
	"bool":   token.Type{Kind: types.Bool},
	"fn":     token.FnDef{},
	"return": token.Return{},
	// "if":       token.If{},
	// "else":     token.Else{},
	// "return":   token.Return{},
	// "while":    token.While{},
	// "for":      token.For{},
	//"in": token.In{},
	// "break":    token.Break{},
	// "continue": token.Continue{},
}

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
		if isNumeric(tk) {
			value := lx.parseInteger()
			lx.pushToken(token.Integer{Value: value})
			continue
		}

		if isAlpha(tk) {
			value := lx.parseAlpha()

			// Check if the value is a reserved keyword
			if keyword, ok := keywords[value]; ok {
				lx.pushToken(keyword)
				continue
			}

			lx.pushToken(token.Identifier{Value: value})

			continue
		}

		switch tk {
		case ' ', '\n', '\t':
			// Skip whitespace
		case '+', '-', '*', '/':
			lx.pushToken(token.Operator{Op: string(tk)})
		case '(':
			lx.pushToken(token.OpenParen{})
		case ')':
			lx.pushToken(token.CloseParen{})
		case '{':
			lx.pushToken(token.OpenBrace{})
		case '}':
			lx.pushToken(token.CloseBrace{})
		case ';':
			lx.pushToken(token.Semicolon{})
		case '&':
			if lx.next() == '&' {
				lx.pushToken(token.Operator{Op: "&&"})
			} else {
				return nil, fmt.Errorf("invalid character: %c", tk)
			}
		case '|':
			if lx.next() == '|' {
				lx.pushToken(token.Operator{Op: "||"})
			} else {
				return nil, fmt.Errorf("invalid character: %c", tk)
			}

		case '!':
			if lx.next() == '=' {
				lx.pushToken(token.Operator{Op: "!="})
			} else {
				lx.putBack()
				lx.pushToken(token.Operator{Op: "!"})
			}

		case '>':
			if lx.next() == '=' {
				lx.pushToken(token.Operator{Op: ">="})
			} else {
				lx.putBack()
				lx.pushToken(token.Operator{Op: ">"})
			}
		case '<':
			if lx.next() == '=' {
				lx.pushToken(token.Operator{Op: "<="})
			} else {
				lx.putBack()
				lx.pushToken(token.Operator{Op: "<"})
			}

		case '=':
			if lx.next() == '=' {
				lx.pushToken(token.Operator{Op: "=="})
			} else {
				lx.putBack()
				lx.pushToken(token.Operator{Op: "="})
			}

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

func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func (lx *lexer) parseAlpha() string {
	value := strings.Builder{}
	value.WriteByte(lx.peek())

	for isAlpha(lx.next()) {
		value.WriteByte(lx.peek())
	}

	// Put back the last character so that we do not return with the position past the bounds of what this funciton handled.
	lx.putBack()

	return value.String()
}
