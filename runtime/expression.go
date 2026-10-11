package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/star-commander-online/starscript-core/value"
)

// evalExpression evaluates a scalar expression with no external dependencies.
// Supported features:
// - numbers, strings, booleans, null
// - unary: !, -
// - binary: +, -, *, /, <, <=, >, >=, ==, !=, &&, ||
// - parentheses
// - JS-style strict equals aliases: ===, !==
func evalInterpolatedExpression(raw string) (value.Value, error) {
	tokens, err := lexExpr(raw)
	if err != nil {
		return value.Value{}, err
	}
	p := &exprParser{tokens: tokens}
	expr, err := p.parseExpression()
	if err != nil {
		return value.Value{}, err
	}
	if !p.check(tokEOF) {
		return value.Value{}, fmt.Errorf("unexpected token %q", p.peek().lexeme)
	}
	return evalExprNode(expr)
}

// exprTokenKind enumerates tokenizer outputs for expression parsing.
type exprTokenKind int

const (
	tokEOF exprTokenKind = iota
	tokNumber
	tokString
	tokBool
	tokNull
	tokLParen
	tokRParen
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokBang
	tokAnd
	tokOr
	tokEqual
	tokNotEqual
	tokLess
	tokLessEqual
	tokGreater
	tokGreaterEqual
)

// exprToken is one lexical item in expression source.
type exprToken struct {
	kind   exprTokenKind
	lexeme string
	line   int
}

// lexExpr converts expression text into parser tokens.
func lexExpr(source string) ([]exprToken, error) {
	tokens := []exprToken{}
	runes := []rune(source)
	pos := 0
	line := 1

	for pos < len(runes) {
		ch := runes[pos]

		if ch == ' ' || ch == '\t' || ch == '\r' {
			pos++
			continue
		}
		if ch == '\n' {
			line++
			pos++
			continue
		}

		switch ch {
		case '(':
			tokens = append(tokens, exprToken{kind: tokLParen, lexeme: "(", line: line})
			pos++
			continue
		case ')':
			tokens = append(tokens, exprToken{kind: tokRParen, lexeme: ")", line: line})
			pos++
			continue
		case '+':
			tokens = append(tokens, exprToken{kind: tokPlus, lexeme: "+", line: line})
			pos++
			continue
		case '-':
			tokens = append(tokens, exprToken{kind: tokMinus, lexeme: "-", line: line})
			pos++
			continue
		case '*':
			tokens = append(tokens, exprToken{kind: tokStar, lexeme: "*", line: line})
			pos++
			continue
		case '/':
			tokens = append(tokens, exprToken{kind: tokSlash, lexeme: "/", line: line})
			pos++
			continue
		case '!':
			if pos+1 < len(runes) && runes[pos+1] == '=' {
				if pos+2 < len(runes) && runes[pos+2] == '=' {
					tokens = append(tokens, exprToken{kind: tokNotEqual, lexeme: "!==", line: line})
					pos += 3
					continue
				}
				tokens = append(tokens, exprToken{kind: tokNotEqual, lexeme: "!=", line: line})
				pos += 2
				continue
			}
			tokens = append(tokens, exprToken{kind: tokBang, lexeme: "!", line: line})
			pos++
			continue
		case '=':
			if pos+1 < len(runes) && runes[pos+1] == '=' {
				if pos+2 < len(runes) && runes[pos+2] == '=' {
					tokens = append(tokens, exprToken{kind: tokEqual, lexeme: "===", line: line})
					pos += 3
					continue
				}
				tokens = append(tokens, exprToken{kind: tokEqual, lexeme: "==", line: line})
				pos += 2
				continue
			}
			return nil, fmt.Errorf("unexpected '=' in expression")
		case '<':
			if pos+1 < len(runes) && runes[pos+1] == '=' {
				tokens = append(tokens, exprToken{kind: tokLessEqual, lexeme: "<=", line: line})
				pos += 2
				continue
			}
			tokens = append(tokens, exprToken{kind: tokLess, lexeme: "<", line: line})
			pos++
			continue
		case '>':
			if pos+1 < len(runes) && runes[pos+1] == '=' {
				tokens = append(tokens, exprToken{kind: tokGreaterEqual, lexeme: ">=", line: line})
				pos += 2
				continue
			}
			tokens = append(tokens, exprToken{kind: tokGreater, lexeme: ">", line: line})
			pos++
			continue
		case '&':
			if pos+1 < len(runes) && runes[pos+1] == '&' {
				tokens = append(tokens, exprToken{kind: tokAnd, lexeme: "&&", line: line})
				pos += 2
				continue
			}
			return nil, fmt.Errorf("single '&' is not valid; use &&")
		case '|':
			if pos+1 < len(runes) && runes[pos+1] == '|' {
				tokens = append(tokens, exprToken{kind: tokOr, lexeme: "||", line: line})
				pos += 2
				continue
			}
			return nil, fmt.Errorf("single '|' is not valid; use ||")
		case '"':
			str, nextPos, err := readStringToken(runes, pos)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, exprToken{kind: tokString, lexeme: str, line: line})
			pos = nextPos
			continue
		}

		if unicode.IsDigit(ch) {
			num, nextPos := readNumberToken(runes, pos)
			tokens = append(tokens, exprToken{kind: tokNumber, lexeme: num, line: line})
			pos = nextPos
			continue
		}

		if unicode.IsLetter(ch) {
			word, nextPos := readWordToken(runes, pos)
			pos = nextPos
			switch word {
			case "true", "false":
				tokens = append(tokens, exprToken{kind: tokBool, lexeme: word, line: line})
			case "null":
				tokens = append(tokens, exprToken{kind: tokNull, lexeme: word, line: line})
			default:
				return nil, fmt.Errorf("unknown identifier %q", word)
			}
			continue
		}

		return nil, fmt.Errorf("unexpected character %q", ch)
	}

	tokens = append(tokens, exprToken{kind: tokEOF, lexeme: "", line: line})
	return tokens, nil
}

// readStringToken reads a quoted string token and decodes escapes.
func readStringToken(runes []rune, start int) (string, int, error) {
	var b strings.Builder
	pos := start + 1 // skip opening quote
	escaped := false

	for pos < len(runes) {
		ch := runes[pos]
		if escaped {
			b.WriteRune('\\')
			b.WriteRune(ch)
			escaped = false
			pos++
			continue
		}
		if ch == '\\' {
			escaped = true
			pos++
			continue
		}
		if ch == '"' {
			decoded, err := strconv.Unquote("\"" + b.String() + "\"")
			if err != nil {
				return "", 0, fmt.Errorf("invalid string literal")
			}
			return decoded, pos + 1, nil
		}
		b.WriteRune(ch)
		pos++
	}

	return "", 0, fmt.Errorf("unterminated string literal")
}

// readNumberToken scans number literals with optional decimal point.
func readNumberToken(runes []rune, start int) (string, int) {
	pos := start
	hasDot := false
	for pos < len(runes) {
		ch := runes[pos]
		if unicode.IsDigit(ch) {
			pos++
			continue
		}
		if ch == '.' && !hasDot {
			hasDot = true
			pos++
			continue
		}
		break
	}
	return string(runes[start:pos]), pos
}

// readWordToken scans alpha keywords like true/false/null.
func readWordToken(runes []rune, start int) (string, int) {
	pos := start
	for pos < len(runes) && unicode.IsLetter(runes[pos]) {
		pos++
	}
	return string(runes[start:pos]), pos
}

// exprNode models one expression AST node for local evaluation.
type exprNode interface{}

// literalNode holds scalar literal values.
type literalNode struct {
	value value.Value
}

// unaryNode holds one unary operator and its right operand.
type unaryNode struct {
	op    exprTokenKind
	right exprNode
}

// binaryNode holds infix operator and both operands.
type binaryNode struct {
	left  exprNode
	op    exprTokenKind
	right exprNode
}

// exprParser performs recursive-descent parsing with precedence levels.
type exprParser struct {
	tokens []exprToken
	pos    int
}

// parseExpression is parser entrypoint.
func (p *exprParser) parseExpression() (exprNode, error) {
	return p.parseOr()
}

// parseOr parses logical OR chains.
func (p *exprParser) parseOr() (exprNode, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.match(tokOr) {
		right, parseErr := p.parseAnd()
		if parseErr != nil {
			return nil, parseErr
		}
		left = &binaryNode{left: left, op: tokOr, right: right}
	}
	return left, nil
}

// parseAnd parses logical AND chains.
func (p *exprParser) parseAnd() (exprNode, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(tokAnd) {
		right, parseErr := p.parseEquality()
		if parseErr != nil {
			return nil, parseErr
		}
		left = &binaryNode{left: left, op: tokAnd, right: right}
	}
	return left, nil
}

// parseEquality parses equality and inequality comparisons.
func (p *exprParser) parseEquality() (exprNode, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.match(tokEqual) || p.match(tokNotEqual) {
		op := p.previous().kind
		right, parseErr := p.parseComparison()
		if parseErr != nil {
			return nil, parseErr
		}
		left = &binaryNode{left: left, op: op, right: right}
	}
	return left, nil
}

// parseComparison parses numeric comparison operators.
func (p *exprParser) parseComparison() (exprNode, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.match(tokLess) || p.match(tokLessEqual) || p.match(tokGreater) || p.match(tokGreaterEqual) {
		op := p.previous().kind
		right, parseErr := p.parseTerm()
		if parseErr != nil {
			return nil, parseErr
		}
		left = &binaryNode{left: left, op: op, right: right}
	}
	return left, nil
}

// parseTerm parses additive operators (+ and -).
func (p *exprParser) parseTerm() (exprNode, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.match(tokPlus) || p.match(tokMinus) {
		op := p.previous().kind
		right, parseErr := p.parseFactor()
		if parseErr != nil {
			return nil, parseErr
		}
		left = &binaryNode{left: left, op: op, right: right}
	}
	return left, nil
}

// parseFactor parses multiplicative operators (* and /).
func (p *exprParser) parseFactor() (exprNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.match(tokStar) || p.match(tokSlash) {
		op := p.previous().kind
		right, parseErr := p.parseUnary()
		if parseErr != nil {
			return nil, parseErr
		}
		left = &binaryNode{left: left, op: op, right: right}
	}
	return left, nil
}

// parseUnary parses unary operators (! and -).
func (p *exprParser) parseUnary() (exprNode, error) {
	if p.match(tokBang) || p.match(tokMinus) {
		op := p.previous().kind
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{op: op, right: right}, nil
	}
	return p.parsePrimary()
}

// parsePrimary parses scalar literals and grouped sub-expressions.
func (p *exprParser) parsePrimary() (exprNode, error) {
	if p.match(tokNumber) {
		raw := p.previous().lexeme
		num, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", raw)
		}
		return &literalNode{value: value.Number(num)}, nil
	}
	if p.match(tokString) {
		return &literalNode{value: value.String(p.previous().lexeme)}, nil
	}
	if p.match(tokBool) {
		return &literalNode{value: value.Bool(p.previous().lexeme == "true")}, nil
	}
	if p.match(tokNull) {
		return &literalNode{value: value.Null()}, nil
	}
	if p.match(tokLParen) {
		inner, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(tokRParen) {
			return nil, fmt.Errorf("expected ')' after expression")
		}
		return inner, nil
	}
	return nil, fmt.Errorf("expected expression, got %q", p.peek().lexeme)
}

// match consumes a token if it matches the requested kind.
func (p *exprParser) match(kind exprTokenKind) bool {
	if p.check(kind) {
		p.pos++
		return true
	}
	return false
}

// check reports whether the current token has a given kind.
func (p *exprParser) check(kind exprTokenKind) bool {
	return p.peek().kind == kind
}

// previous returns the token immediately before current parser position.
func (p *exprParser) previous() exprToken {
	return p.tokens[p.pos-1]
}

// peek returns the current lookahead token.
func (p *exprParser) peek() exprToken {
	return p.tokens[p.pos]
}

// evalExprNode recursively evaluates expression nodes.
func evalExprNode(node exprNode) (value.Value, error) {
	switch n := node.(type) {
	case *literalNode:
		return n.value, nil
	case *unaryNode:
		right, err := evalExprNode(n.right)
		if err != nil {
			return value.Value{}, err
		}
		switch n.op {
		case tokBang:
			return value.Bool(!isTruthy(right)), nil
		case tokMinus:
			if right.Kind != value.KindNumber {
				return value.Value{}, fmt.Errorf("unary '-' requires numeric operand")
			}
			return value.Number(-right.Number), nil
		default:
			return value.Value{}, fmt.Errorf("unsupported unary operator")
		}
	case *binaryNode:
		left, err := evalExprNode(n.left)
		if err != nil {
			return value.Value{}, err
		}
		right, err := evalExprNode(n.right)
		if err != nil {
			return value.Value{}, err
		}
		return evalBinaryNode(n.op, left, right)
	default:
		return value.Value{}, fmt.Errorf("unsupported expression node")
	}
}

// evalBinaryNode applies one binary operator to two scalar runtime values.
// It intentionally keeps coercion rules conservative and explicit.
func evalBinaryNode(op exprTokenKind, left value.Value, right value.Value) (value.Value, error) {
	switch op {
	case tokPlus:
		if left.Kind == value.KindString || right.Kind == value.KindString {
			return value.String(left.Stringify() + right.Stringify()), nil
		}
		l, err := requireNumberValue(left)
		if err != nil {
			return value.Value{}, err
		}
		r, err := requireNumberValue(right)
		if err != nil {
			return value.Value{}, err
		}
		return value.Number(l + r), nil
	case tokMinus:
		l, err := requireNumberValue(left)
		if err != nil {
			return value.Value{}, err
		}
		r, err := requireNumberValue(right)
		if err != nil {
			return value.Value{}, err
		}
		return value.Number(l - r), nil
	case tokStar:
		l, err := requireNumberValue(left)
		if err != nil {
			return value.Value{}, err
		}
		r, err := requireNumberValue(right)
		if err != nil {
			return value.Value{}, err
		}
		return value.Number(l * r), nil
	case tokSlash:
		l, err := requireNumberValue(left)
		if err != nil {
			return value.Value{}, err
		}
		r, err := requireNumberValue(right)
		if err != nil {
			return value.Value{}, err
		}
		if r == 0 {
			return value.Value{}, fmt.Errorf("division by zero")
		}
		return value.Number(l / r), nil
	case tokEqual:
		return value.Bool(equals(left, right)), nil
	case tokNotEqual:
		return value.Bool(!equals(left, right)), nil
	case tokLess, tokLessEqual, tokGreater, tokGreaterEqual:
		l, err := requireNumberValue(left)
		if err != nil {
			return value.Value{}, err
		}
		r, err := requireNumberValue(right)
		if err != nil {
			return value.Value{}, err
		}
		switch op {
		case tokLess:
			return value.Bool(l < r), nil
		case tokLessEqual:
			return value.Bool(l <= r), nil
		case tokGreater:
			return value.Bool(l > r), nil
		default:
			return value.Bool(l >= r), nil
		}
	case tokAnd:
		return value.Bool(isTruthy(left) && isTruthy(right)), nil
	case tokOr:
		return value.Bool(isTruthy(left) || isTruthy(right)), nil
	default:
		return value.Value{}, fmt.Errorf("unsupported binary operator")
	}
}

// requireNumberValue is a strict numeric assertion helper.
func requireNumberValue(v value.Value) (float64, error) {
	if v.Kind != value.KindNumber {
		return 0, fmt.Errorf("expected number, got %s", v.Kind)
	}
	return v.Number, nil
}

// equals performs kind-aware equality checks for scalar StarScript values.
func equals(left value.Value, right value.Value) bool {
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case value.KindNull:
		return true
	case value.KindBool:
		return left.Bool == right.Bool
	case value.KindNumber:
		return left.Number == right.Number
	case value.KindString:
		return left.String == right.String
	default:
		return left.Stringify() == right.Stringify()
	}
}
