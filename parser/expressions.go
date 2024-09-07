package parser

import (
	"fmt"
	"leoscript/token"
)

// Start parsing an expression group.
// An expression group is a group of expressions whose length is not known statically and can contain any number of sub expressions.
// An expression group is terminated by a semicolon token or a close parenthesis token.
func (p *Parser) parseExpressionGroup(isSubgroup bool) (Expression, error) {
	var root Expression

	for tk := p.peek(); tk != nil; tk = p.peek() {
		// fmt.Printf("tk: %T %v\n", tk, tk)

		switch tk := tk.(type) {

		case token.Integer: // This will only hit if it is the first token in the expression, otherwise it will have been handled by another branch.

			if root != nil {
				return nil, fmt.Errorf("unexpected integer token preceeding another expression, v=%v", tk)
			}
			root = IntegerLiteral{Value: tk.Value}

		case token.Semicolon:
			if root == nil {
				return nil, fmt.Errorf("unexpected semicolon token with no expression")
			}

			if isSubgroup {
				panic("subgroup ended by semicolon")
			}

			// TODO: Really needed?
			p.current-- // put the semicolon back, it might be used to end the parent
			return root, nil

		case token.OpenParen:
			p.next() // consume the open parenthesis token
			expr, err := p.parseExpressionGroup(true)
			if err != nil {
				return nil, fmt.Errorf("failed to parse expression group: %w", err)
			}

			root = expr

		case token.CloseParen:
			if !isSubgroup {
				// return nil, fmt.Errorf("unexpected close parenthesis token")
				panic("top-level expression group ended by close parenthesis")
			}

			return root, nil

		case token.Binary:
			expr, err := p.parseBinaryExpression(root)
			if err != nil {
				return nil, err
			}

			// If this is a subgroup, we need to force the priority to be higher
			if isSubgroup {
				if binExpr, ok := expr.(BinaryExpression); ok {
					binExpr.Priority = 100
					expr = binExpr
				}
			}

			root = expr

		default:
			return nil, fmt.Errorf("unexpected token in expression: T=%T V=%v", tk, tk)
		}

		p.next() // consume the token
	}

	panic("reached EOF without completing the expression")
}

// An aotmic expression is an expression that can stand on its own and does not
// require any other expressions to be complete.
func (p *Parser) parseAtomicExpression() (Expression, error) {
	tk := p.peek()
	if tk == nil {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if intTk, ok := tk.(token.Integer); ok {
		return IntegerLiteral{Value: intTk.Value}, nil
	}

	if _, ok := tk.(token.Binary); ok {
		return p.parseNumberOperation()
	}

	//function call

	return nil, fmt.Errorf("unexpected token: %v", tk)
}

func (p *Parser) parseNumberOperation() (Expression, error) {
	var binTk token.Binary
	tk := p.peek()
	if bt, ok := tk.(token.Binary); ok {
		binTk = bt
	} else {
		panic("unexpected token type")
	}

	switch binTk.Operation {
	case "-":
		// Negation of an integer
		intTk := p.next()
		if _, ok := intTk.(token.Integer); !ok {
			return nil, fmt.Errorf("expected integer token, got %v", intTk)
		}

		return IntegerLiteral{Value: -intTk.(token.Integer).Value}, nil
	case "+":
		intTk := p.next()
		if _, ok := intTk.(token.Integer); !ok {
			return nil, fmt.Errorf("expected integer token, got %v", intTk)
		}

		return IntegerLiteral{Value: intTk.(token.Integer).Value}, nil
	default:
		return nil, fmt.Errorf("unexpected operator: %v", binTk)
	}
}

func (p *Parser) parseBinaryExpression(root Expression) (Expression, error) {
	var binTk token.Binary
	tk := p.peek()
	if bt, ok := tk.(token.Binary); ok {
		binTk = bt
	} else {
		panic("unexpected token type")
	}

	// No preceeding expression, this is the first one
	if root == nil {
		return p.parseNumberOperation()
	}

	var left Expression = root
	var right Expression

	var priority int
	switch binTk.Operation {
	case "+", "-":
		priority = 0
	case "*", "/":
		priority = 1
	default:
		panic("invalid operator in binary expression")
	}

	nextToken := p.next() // consume the operator token
	var err error
	if _, ok := nextToken.(token.OpenParen); ok {
		p.next() // consume the open parenthesis token
		right, err = p.parseExpressionGroup(true)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expression group: %w", err)
		}

		// force the priority to be higher than any other operator
		// so that the expression in the parenthesis is evaluated first
		priority = 100
	} else {
		right, err = p.parseAtomicExpression()
		if err != nil {
			return nil, fmt.Errorf("failed to parse right expression: %w", err)
		}
	}

	// fmt.Printf("binary consumes tk: %T %v\n", right, right)

	// Do a right swap if the priority of the current operator is higher
	if leftBinExpr, ok := left.(BinaryExpression); ok && priority > leftBinExpr.Priority {
		right = BinaryExpression{
			Left:     leftBinExpr.Right,
			Right:    right,
			Op:       binTk.Operation,
			Priority: priority,
		}

		leftBinExpr.Right = right
		root = leftBinExpr
	} else {
		root = BinaryExpression{
			Left:     left,
			Right:    right,
			Op:       binTk.Operation,
			Priority: priority,
		}
	}

	return root, nil
}
