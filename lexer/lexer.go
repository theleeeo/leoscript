package lexer

type TokenType string

const (
	// An integer
	Integer TokenType = "INT"
	// A binary operator
	Binary TokenType = "BINOP"
)

type Token struct {
	Type    TokenType
	Literal string
}

func Parse(input string) ([]Token, error) {
	var tokens []Token

	for i := 0; i < len(input); i++ {
		if isWhitespace(input[i]) {
			continue
		}

		if isNumeric(input[i]) {
			tokens = append(tokens, Token{Type: Integer, Literal: string(input[i])})
		} else if input[i] == '+' || input[i] == '-' || input[i] == '*' || input[i] == '/' {
			tokens = append(tokens, Token{Type: Binary, Literal: string(input[i])})
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

func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}
