package compute

import (
	"errors"
	"fmt"
	"galculator/internel/lexer"
	"strconv"
)

// This interface happens to match https://golang.org/pkg/text/scanner/#Scanner.Next
type tokenEmitter interface {
	Next() lexer.Token
}

func parse(tokens tokenEmitter, vm variableMap) (operatorStack []operator, operantStack []expression, err error) {
	operators := map[lexer.Operator]operator{
		lexer.Add: add,
		lexer.Sub: sub,
		lexer.Mul: mul,
		lexer.Div: div,
	}

	exp, err2 := parseExpression(tokens, vm)
	if err2 != nil {
		return nil, nil, err2
	}
	if exp == nil {
		return
	}
	for next := tokens.Next(); next != nil; next = tokens.Next() {
		switch t := next.(type) {
		case lexer.Operator:
			if _, ok := operators[t]; ok {
				exp2, err2 := parseExpression(tokens, vm)
				if err2 != nil {
					return nil, nil, err2
				}
				exp = OperatorExpression{Left: exp, Op: t, Right: exp2}
			} else {
				panic("?")
			}
		case lexer.Number:
			// integer, err := strconv.ParseInt(t.Value, 10, 64)
			// if err != nil {
			// 	panic(err)
			// }
			// operantStack = append(operantStack, NumberExpression{number: integer})
			err = &ParsingError{"a number can not follow an expression"}
			return
		case lexer.LeftParentheses:
			// pe, err2 := ParseParenthesisExpression(tokens, vm)
			// if err2 != nil {
			// 	err = err2
			// 	return
			// }
			// operantStack = append(operantStack, pe)
			err = &ParsingError{"a ( can not follow an expression"}
			return
		case lexer.Identifier:

			// exp, err2 := parseIdentifierExpression(tokens, t, vm) // todo: initt variable map elsewhere
			// if err2 != nil {
			// 	err = err2
			// 	return
			// }
			// operantStack = append(operantStack, exp)
			err = &ParsingError{"an identifier can not follow an expression"}
			return
		default:
			fmt.Println("Warning:", t.Type(), t.Literal(), "is ignored")
		}
	}

	operantStack = append(operantStack, exp)

	return
}

// ParseParenthesisExpression parse a sequence of tokens and return the first full parenthesis expression.
// ( 1 + 1 ) + 1 ) will return expression"((1+1)+1)"
// This function always inserts a leading ( at the beginning of the tokens.
func ParseParenthesisExpression(tokens tokenEmitter, vm variableMap) (ParenthesesExpression, error) {
	count := 1
	var readTokens []lexer.Token
	for next := tokens.Next(); next != nil; next = tokens.Next() {
		readTokens = append(readTokens, next)
		switch next.(type) {
		case nil:
			return ParenthesesExpression{}, &ParsingError{fmt.Sprintf("Missing %d ) parentheses", count)}
		case lexer.LeftParentheses:
			count++
		case lexer.RightParentheses:
			count--
		}
		if count == 0 {
			operatorStack, operantStack, err := parse(&tokenSliceEmitter{tokens: readTokens[:len(readTokens)-1]}, vm)
			return ParenthesesExpression{
				OperatorStack: operatorStack,
				OperantStack:  operantStack,
			}, err
		}
	}
	return ParenthesesExpression{}, &ParsingError{fmt.Sprintf("Missing %d ) parentheses", count)}
}

// This is the context free grammar for identifier
// identifier = expression
// identifier operator expression
// identifier EOF
func parseIdentifierExpression(tokens tokenEmitter, identifier lexer.Identifier, vm map[string]int64) (expression, error) {
	next := tokens.Next()
	switch token := next.(type) {
	case lexer.EqualSign:
		// return an assignment expression
		exp, err := parseExpression(tokens, vm)
		if err != nil {
			return nil, err
		}
		return AssignmentExpression{
			Name:       identifier.Literal(),
			expression: exp,
			vm:         vm,
		}, nil
	case lexer.Operator:
		exp, err := parseExpression(tokens, vm)
		if err != nil {
			return nil, err
		}
		return OperatorExpression{
			Op:    token,
			Left:  IdentifierExpression{Name: identifier.Literal(), ValueMap: vm},
			Right: exp,
		}, nil
	case nil: // EOF
		// return the identifier itself
		return IdentifierExpression{Name: identifier.Literal(), ValueMap: vm}, nil
	default:
		return nil, &ParsingError{fmt.Sprintf("An identifier must be followed by 1 of the 4: = expression, operator expression, ), EOF. But got %T:%v", next, next)}
	}
	return nil, &ParsingError{"3 todo: should not reach"}
}

// This function returns the first expression parsed immediated or an error.
// expression |
// number
// parentheses-expression
// identifier-expression
// assignment-expression
func parseExpression(tokens tokenEmitter, vm variableMap) (expression, error) {
	next := tokens.Next()
	switch token := next.(type) {
	case nil: // EOF
		return nil, nil
	case lexer.Number:
		integer, err := strconv.ParseInt(token.Literal(), 10, 64)
		if err != nil {
			return nil, err
		}
		np := NumberExpression{number: integer}

		next = tokens.Next()
		if next == nil {
			return np, nil
		}

		if op, ok := next.(lexer.Operator); ok {
			exp, err := parseExpression(tokens, vm)
			if err != nil {
				return nil, err
			}
			return OperatorExpression{Op: op, Left: np, Right: exp}, nil
		}

		return nil, newParsingError("%s is not expected after %s.", next.Literal(), token.Literal())

	case lexer.LeftParentheses:
		lp, err := ParseParenthesisExpression(tokens, vm)
		return lp, err

	case lexer.Identifier:
		return parseIdentifierExpression(tokens, token, vm)
	case lexer.Operator:
		if token == lexer.Add || token == lexer.Sub {
			exp, err := parseExpression(tokens, vm)
			if err != nil {
				return nil, err
			}
			return OperatorExpression{Op: token, Left: nil, Right: exp}, nil
		}
		return nil, &ParsingError{fmt.Sprintf("invalid operator %s here", token.Literal())}
	}
	return nil, errors.New("1 todo: should not reach")
}
