package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	// Special tokens
	TokenEOF TokenType = iota
	TokenNewline
	TokenError

	// Literals
	TokenNumber
	TokenString
	TokenIdent

	// Keywords
	TokenIf
	TokenElif
	TokenElse
	TokenEnd
	TokenDo
	TokenWhile
	TokenFor
	TokenIn
	TokenReturn
	TokenTrue
	TokenFalse
	TokenNull

	// Operators
	TokenAssign   // =
	TokenEq       // ==
	TokenNeq      // !=
	TokenLt       // <
	TokenGt       // >
	TokenLte      // <=
	TokenGte      // >=
	TokenAnd      // &&
	TokenOr       // ||
	TokenNot      // !
	TokenPlus     // +
	TokenMinus    // -
	TokenMul      // *
	TokenDiv      // /
	TokenLBracket // [
	TokenRBracket // ]
	TokenDotDot   // ..
	TokenDot      // .
	TokenLParen   // (
	TokenRParen   // )
	TokenComma    // ,
	TokenDollar   // $

	// Special
	TokenComment
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type:%d, Value:%q, Line:%d, Col:%d}", t.Type, t.Value, t.Line, t.Column)
}

// Lexer tokenizes script source into a stream of tokens.
type Lexer struct {
	source []rune
	pos    int
	line   int
	column int
	tokens []Token
}

func NewLexer(source string) *Lexer {
	// Initialize a new lexer with the given source code.
	return &Lexer{
		source: []rune(source),
		pos:    0,
		line:   1,
		column: 1,
		tokens: []Token{},
	}
}

func (l *Lexer) Lex() ([]Token, error) {
	// Loop through the source and lex each token
	for l.pos < len(l.source) {
		// Lex the next token in the source
		if err := l.nextToken(); err != nil {
			return nil, err
		}
	}

	// Append the EOF token to signify the end of the token stream
	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line, Column: l.column})
	return l.tokens, nil
}

func (l *Lexer) nextToken() error {
	l.skipWhitespaceExceptNewline()

	if l.pos >= len(l.source) {
		return nil
	}

	ch := l.source[l.pos]

	// Handle newlines
	if ch == '\n' {
		l.emit(TokenNewline, "\n")
		l.advance()
		l.line++
		l.column = 1
		return nil
	}

	// Handle comments (# to end of line)
	if ch == '#' {
		l.skipComment()
		return nil
	}

	// Handle strings
	if ch == '"' {
		return l.lexString()
	}

	// Handle numbers
	if unicode.IsDigit(rune(ch)) {
		return l.lexNumber()
	}

	// Handle identifiers and keywords
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		return l.lexIdentOrKeyword()
	}

	// Handle two-character operators
	if l.pos+1 < len(l.source) {
		twoChar := string([]rune{rune(l.source[l.pos]), rune(l.source[l.pos+1])})
		switch twoChar {
		case "==":
			l.emit(TokenEq, "==")
			l.advance()
			l.advance()
			return nil
		case "!=":
			l.emit(TokenNeq, "!=")
			l.advance()
			l.advance()
			return nil
		case "<=":
			l.emit(TokenLte, "<=")
			l.advance()
			l.advance()
			return nil
		case ">=":
			l.emit(TokenGte, ">=")
			l.advance()
			l.advance()
			return nil
		case "&&":
			l.emit(TokenAnd, "&&")
			l.advance()
			l.advance()
			return nil
		case "||":
			l.emit(TokenOr, "||")
			l.advance()
			l.advance()
			return nil
		case "..":
			l.emit(TokenDotDot, "..")
			l.advance()
			l.advance()
			return nil
		}
	}

	// Handle single-character tokens
	switch ch {
	case '=':
		l.emit(TokenAssign, "=")
	case '<':
		l.emit(TokenLt, "<")
	case '>':
		l.emit(TokenGt, ">")
	case '!':
		l.emit(TokenNot, "!")
	case '+':
		l.emit(TokenPlus, "+")
	case '-':
		l.emit(TokenMinus, "-")
	case '*':
		l.emit(TokenMul, "*")
	case '/':
		l.emit(TokenDiv, "/")
	case '[':
		l.emit(TokenLBracket, "[")
	case ']':
		l.emit(TokenRBracket, "]")
	case '(':
		l.emit(TokenLParen, "(")
	case ')':
		l.emit(TokenRParen, ")")
	case ',':
		l.emit(TokenComma, ",")
	case '$':
		l.emit(TokenDollar, "$")
	case '.':
		l.emit(TokenDot, ".")
	default:
		return fmt.Errorf("unexpected character %q at line %d, column %d", ch, l.line, l.column)
	}
	l.advance()
	return nil
}

func (l *Lexer) lexString() error {
	l.advance() // consume opening quote
	var sb strings.Builder
	startCol := l.column - 1

	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == '"' {
			l.advance() // consume closing quote
			l.emitAt(TokenString, sb.String(), l.line, startCol)
			return nil
		}
		if ch == '\\' && l.pos+1 < len(l.source) {
			l.advance()
			nextCh := l.source[l.pos]
			switch nextCh {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '"':
				sb.WriteRune('"')
			case '\\':
				sb.WriteRune('\\')
			default:
				sb.WriteRune(nextCh)
			}
			l.advance()
		} else if ch == '\n' {
			return fmt.Errorf("Unterminated string at line %d", l.line)
		} else {
			sb.WriteRune(rune(ch))
			l.advance()
		}
	}
	return fmt.Errorf("Unterminated string at line %d", l.line)
}

func (l *Lexer) lexNumber() error {
	startPos := l.pos
	startCol := l.column

	// Read integer part
	for l.pos < len(l.source) && unicode.IsDigit(rune(l.source[l.pos])) {
		l.advance()
	}

	// Read optional decimal part
	if l.pos < len(l.source) && l.source[l.pos] == '.' &&
		l.pos+1 < len(l.source) && unicode.IsDigit(rune(l.source[l.pos+1])) {
		l.advance() // consume dot
		for l.pos < len(l.source) && unicode.IsDigit(rune(l.source[l.pos])) {
			l.advance()
		}
	}

	value := string(l.source[startPos:l.pos])
	l.emitAt(TokenNumber, value, l.line, startCol)
	return nil
}

func (l *Lexer) lexIdentOrKeyword() error {
	startPos := l.pos
	startCol := l.column

	for l.pos < len(l.source) && (unicode.IsLetter(rune(l.source[l.pos])) ||
		unicode.IsDigit(rune(l.source[l.pos])) || l.source[l.pos] == '_') {
		l.advance()
	}

	value := string(l.source[startPos:l.pos])
	tokenType := l.keywordType(value)
	l.emitAt(tokenType, value, l.line, startCol)
	return nil
}

func (l *Lexer) keywordType(word string) TokenType {
	switch word {
	case "if":
		return TokenIf
	case "elif":
		return TokenElif
	case "else":
		return TokenElse
	case "end":
		return TokenEnd
	case "do":
		return TokenDo
	case "while":
		return TokenWhile
	case "for":
		return TokenFor
	case "in":
		return TokenIn
	case "return":
		return TokenReturn
	case "true":
		return TokenTrue
	case "false":
		return TokenFalse
	case "null":
		return TokenNull
	default:
		return TokenIdent
	}
}

func (l *Lexer) skipWhitespaceExceptNewline() {
	for l.pos < len(l.source) && (l.source[l.pos] == ' ' || l.source[l.pos] == '\t') {
		l.advance()
	}
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.source) && l.source[l.pos] != '\n' {
		l.advance()
	}
}

func (l *Lexer) advance() {
	if l.pos < len(l.source) {
		l.pos++
		l.column++
	}
}

func (l *Lexer) emit(tType TokenType, value string) {
	l.emitAt(tType, value, l.line, l.column-len(value))
}

func (l *Lexer) emitAt(tType TokenType, value string, line, column int) {
	l.tokens = append(l.tokens, Token{
		Type:   tType,
		Value:  value,
		Line:   line,
		Column: column,
	})
}
