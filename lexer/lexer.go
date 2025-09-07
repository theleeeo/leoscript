package lexer

import (
	"fmt"
	"strings"
)

var keywords = map[string]Token{
	// Literals
	"true":  Boolean{Value: true},
	"false": Boolean{Value: false},

	"var":    VarDecl{},
	"fn":     FnDef{},
	"return": Return{},
	"if":     If{},
	"else":   Else{},
	"while":  While{},
	"stub":   StubDef{},
	"export": Exported{},
	"struct": StructDef{},
}

type lexer struct {
	input string
	pos   int

	tokens []Token
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
	if lx.pos >= len(lx.input) {
		return 0
	}

	return lx.input[lx.pos]
}

func (lx *lexer) pushToken(tk Token) {
	lx.tokens = append(lx.tokens, tk)
}

func MustTokenize(input string) []Token {
	tokens, err := Tokenize(input)
	if err != nil {
		panic(err)
	}

	return tokens
}

func Tokenize(input string) ([]Token, error) {
	lx := lexer{input: input}

	for tk := lx.peek(); tk != 0; tk = lx.next() {
		if isNumeric(tk) {
			value := lx.parseInteger()
			lx.pushToken(Integer{Value: value})
			continue
		}

		if isAlpha(tk) {
			var v strings.Builder

			for {
				value := lx.parseAlpha()

				// Check if the value is a reserved keyword
				if keyword, ok := keywords[value]; ok {
					if v.Len() > 0 {
						return nil, fmt.Errorf("invalid identifier: %s%s", v.String(), value) // TODO
					}
					lx.pushToken(keyword)
					break
				}

				v.WriteString(value)

				if lx.next() != '.' {
					lx.putBack()
					break
				}

				v.WriteByte('.')
				lx.next() // Consume the dot (.)
			}

			if v.Len() > 0 {
				lx.pushToken(Identifier{Value: v.String()})
			}

			continue
		}

		switch tk {
		case ' ', '\n', '\t', '\r':
			// Skip whitespace
		case '/':
			// Check if the next token matches a pattern for a comment
			switch tk := lx.next(); tk {
			case '/': // Single line comment
				for {
					tk = lx.next()
					if tk == '\n' || tk == 0 {
						break
					}
				}
			case '*': // Multi-line comment
				for {
					tk = lx.next()
					if tk == 0 {
						return nil, fmt.Errorf("unclosed multi-line comment")
					}
					if tk == '*' {
						if lx.next() == '/' {
							break
						} else {
							lx.putBack()
						}
					}
				}
			default:
				// The next character did not match a comment pattern.
				// Put the next token back and push the operator
				lx.putBack()
				lx.pushToken(Operator{Op: "/"})
			}
		case '+', '-', '*':
			lx.pushToken(Operator{Op: string(tk)})
		case '(':
			lx.pushToken(OpenParen{})
		case ')':
			lx.pushToken(CloseParen{})
		case '{':
			lx.pushToken(OpenBrace{})
		case '}':
			lx.pushToken(CloseBrace{})
		case '[':
			lx.pushToken(OpenBracket{})
		case ']':
			lx.pushToken(CloseBracket{})
		case ';':
			lx.pushToken(Semicolon{})
		case ',':
			lx.pushToken(Comma{})
		case '&':
			if lx.next() == '&' {
				lx.pushToken(Operator{Op: "&&"})
			} else {
				return nil, fmt.Errorf("invalid character: %c", tk)
			}
		case '|':
			if lx.next() == '|' {
				lx.pushToken(Operator{Op: "||"})
			} else {
				return nil, fmt.Errorf("invalid character: %c", tk)
			}

		case '!':
			if lx.next() == '=' {
				lx.pushToken(Operator{Op: "!="})
			} else {
				lx.putBack()
				lx.pushToken(Operator{Op: "!"})
			}

		case '>':
			if lx.next() == '=' {
				lx.pushToken(Operator{Op: ">="})
			} else {
				lx.putBack()
				lx.pushToken(Operator{Op: ">"})
			}
		case '<':
			if lx.next() == '=' {
				lx.pushToken(Operator{Op: "<="})
			} else {
				lx.putBack()
				lx.pushToken(Operator{Op: "<"})
			}

		case '=':
			if lx.next() == '=' {
				lx.pushToken(Operator{Op: "=="})
			} else {
				lx.putBack()
				lx.pushToken(Operator{Op: "="})
			}
		case '"':
			str, err := lx.parseString()
			if err != nil {
				return nil, err
			}

			lx.pushToken(StringLiteral{Value: str})
		case ':':
			lx.pushToken(Colon{})
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

func (lx *lexer) parseString() (string, error) {
	value := strings.Builder{}

	for tk := lx.next(); tk != 0 && tk != '"'; tk = lx.next() {
		if tk == '\\' {
			tk = lx.next()
			if tk == 0 {
				return "", fmt.Errorf("unclosed string literal")
			}

			switch tk {
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case '"':
				value.WriteByte('"')
			case '\\':
				value.WriteByte('\\')
			default:
				return "", fmt.Errorf("invalid escape sequence: \\%c", tk)
			}
			continue
		}
		value.WriteByte(tk)
	}

	if lx.peek() != '"' {
		return "", fmt.Errorf("unclosed string literal")
	}

	return value.String(), nil
}
