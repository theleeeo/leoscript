package lexer

import (
	"fmt"
	"leoscript/token"
)

func Tokenize(input string) ([]token.Token, error) {
	var tokens []token.Token

	for i := 0; i < len(input); i++ {
		switch {
		case isWhitespace(input[i]):
			continue
		case isNumeric(input[i]):
			value, j := parseInteger(input[i:])
			tokens = append(tokens, token.Integer{Value: value})
			// Skip the number of characters we just parsed
			// -1 because the loop will increment i
			i += j - 1
		case input[i] == '+' || input[i] == '-' || input[i] == '*' || input[i] == '/':
			tokens = append(tokens, token.Binary{Operation: string(input[i])})
		case input[i] == '(':
			tokens = append(tokens, token.OpenParen{})
		case input[i] == ')':
			tokens = append(tokens, token.CloseParen{})
		case input[i] == ';':
			tokens = append(tokens, token.Semicolon{})
		default:
			return nil, fmt.Errorf("invalid character: %c", input[i])
		}
	}

	return tokens, nil
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\n' || char == '\t'
}

func isNumeric(char byte) bool {
	return char >= '0' && char <= '9'
}

func parseInteger(input string) (int, int) {
	var value int
	var i int

	for i = 0; i < len(input); i++ {
		if !isNumeric(input[i]) {
			break
		}

		value = value*10 + int(input[i]-'0')
	}

	return value, i
}

// func isAlpha(char byte) bool {
// 	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
// }
