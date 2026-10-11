package parser

import (
	"fmt"
	"strings"

	"github.com/star-commander-online/starscript-core/ast"
)

// Error is a structured parser failure with line information.
type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error on line %d: %s", e.Line, e.Message)
}

func Parse(source string) (*ast.Program, error) {
	// Lex the source into tokens
	lexer := NewLexer(source)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	// Parse tokens into AST
	parser := &parser{
		tokens: tokens,
		pos:    0,
	}

	statements, err := parser.parseProgram()
	if err != nil {
		return nil, err
	}

	return &ast.Program{Statements: statements}, nil
}

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) parseProgram() ([]ast.Statement, error) {
	var statements []ast.Statement

	for !p.isAtEnd() {
		// Skip newlines at top level
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}

		// Check for unexpected end markers
		if p.peekType() == TokenEnd || p.peekType() == TokenElif || p.peekType() == TokenElse {
			return nil, p.errorf("Unexpected %q at statement level", p.peek().Value)
		}

		// Parse the next statement in the program
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

func (p *parser) parseStatement() (ast.Statement, error) {
	p.skipNewlines()

	// Check if we reached the end of the input
	if p.isAtEnd() {
		return nil, nil
	}

	switch p.peekType() {
	case TokenIf:
		return p.parseIf()
	case TokenWhile:
		return p.parseWhile()
	case TokenFor:
		return p.parseFor()
	case TokenReturn:
		return p.parseReturn()
	case TokenIdent:
		return p.parseAssignmentOrCommand()
	default:
		return nil, p.errorf("Unexpected token %q", p.peek().Value)
	}
}

func (p *parser) parseIf() (ast.Statement, error) {
	ifToken := p.advance()
	if ifToken.Type != TokenIf && ifToken.Type != TokenElif {
		return nil, p.errorf("Expected if or elif")
	}

	// Parse the condition (everything until 'do')
	condition, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	if p.peekType() != TokenDo {
		return nil, p.errorf("Expected 'do' after condition, got %q", p.peek().Value)
	}
	p.advance() // consume 'do'

	// Parse then block
	thenBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBlock []ast.Statement

	// Handle elif/else by checking what comes next
	if p.peekType() == TokenElif {
		// Recursively parse elif as a nested if (don't consume elif here, parseIf will do it)
		elifStmt, err := p.parseIf()
		if err != nil {
			return nil, err
		}
		elseBlock = []ast.Statement{elifStmt}
	} else if p.peekType() == TokenElse {
		p.advance() // consume 'else'

		if p.peekType() != TokenDo {
			return nil, p.errorf("Expected 'do' after else")
		}
		p.advance() // consume 'do'

		finalElse, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		elseBlock = finalElse
	}

	if p.peekType() != TokenEnd {
		return nil, p.errorf("Expected 'end' to close if/elif block")
	}
	p.advance() // consume 'end'

	return &ast.IfStmt{
		Condition: condition,
		Then:      thenBlock,
		Else:      elseBlock,
		AtLine:    ifToken.Line,
	}, nil
}

func (p *parser) parseWhile() (ast.Statement, error) {
	whileToken := p.advance()
	if whileToken.Type != TokenWhile {
		return nil, p.errorf("Expected while")
	}

	condition, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	if p.peekType() != TokenDo {
		return nil, p.errorf("Expected 'do' after while condition")
	}
	p.advance() // consume 'do'

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.peekType() != TokenEnd {
		return nil, p.errorf("Expected 'end' to close while block")
	}
	p.advance() // consume 'end'

	return &ast.WhileStmt{
		Condition: condition,
		Body:      body,
		AtLine:    whileToken.Line,
	}, nil
}

func (p *parser) parseFor() (ast.Statement, error) {
	forToken := p.advance()
	if forToken.Type != TokenFor {
		return nil, p.errorf("Expected for")
	}

	// Expect: identifier 'in' '[' expr '..' expr ']'
	if p.peekType() != TokenIdent {
		return nil, p.errorf("Expected loop variable in for statement")
	}
	varToken := p.advance()
	varName := varToken.Value

	if p.peekType() != TokenIn {
		return nil, p.errorf("Expected 'in' in for statement")
	}
	p.advance() // consume 'in'

	if p.peekType() != TokenLBracket {
		return nil, p.errorf("Expected '[' in for range")
	}
	p.advance() // consume '['

	startExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peekType() != TokenDotDot {
		return nil, p.errorf("Expected '..' in for range")
	}
	p.advance() // consume '..'

	endExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peekType() != TokenRBracket {
		return nil, p.errorf("Expected ']' to close for range")
	}
	p.advance() // consume ']'

	if p.peekType() != TokenDo {
		return nil, p.errorf("Expected 'do' after for range")
	}
	p.advance() // consume 'do'

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.peekType() != TokenEnd {
		return nil, p.errorf("Expected 'end' to close for block")
	}
	p.advance() // consume 'end'

	return &ast.ForStmt{
		VarName:   varName,
		StartExpr: startExpr,
		EndExpr:   endExpr,
		Body:      body,
		AtLine:    forToken.Line,
	}, nil
}

func (p *parser) parseReturn() (ast.Statement, error) {
	returnToken := p.advance()
	if returnToken.Type != TokenReturn {
		return nil, p.errorf("Expected return")
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.ReturnStmt{
		Expression: expr,
		AtLine:     returnToken.Line,
	}, nil
}

func (p *parser) parseAssignmentOrCommand() (ast.Statement, error) {
	nameToken := p.advance()
	name := nameToken.Value

	if p.peekType() == TokenAssign {
		// Assignment
		p.advance() // consume '='

		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Return the assignment statement
		return &ast.AssignmentStmt{
			Name:       name,
			Expression: expr,
			AtLine:     nameToken.Line,
		}, nil
	}

	// Command: name followed by arguments
	var args []string

	// Get all arguments
	for !p.isAtEnd() && p.peekType() != TokenNewline && p.peekType() != TokenEnd &&
		p.peekType() != TokenElif && p.peekType() != TokenElse {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	// Return the command statement
	return &ast.CommandStmt{
		Name:   name,
		Args:   args,
		AtLine: nameToken.Line,
	}, nil
}

func (p *parser) parseBlock() ([]ast.Statement, error) {
	var statements []ast.Statement

	for !p.isAtEnd() {
		p.skipNewlines()

		if p.peekType() == TokenEnd || p.peekType() == TokenElif || p.peekType() == TokenElse {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

// reconstructTokens intelligently joins tokens back into a string,
// avoiding spaces between $ and identifiers
func (p *parser) reconstructTokens(startPos int, endPos int) string {
	var result strings.Builder

	for i := startPos; i < endPos; i++ {
		token := p.tokens[i]

		// Add space before this token unless it's preceded by $
		if result.Len() > 0 {
			// Check if previous token was $
			prevToken := p.tokens[i-1]
			if prevToken.Type != TokenDollar {
				result.WriteRune(' ')
			}
		}

		result.WriteString(token.Value)
	}

	return result.String()
}

func (p *parser) parseCondition() (string, error) {
	startPos := p.pos
	depth := 0

	for !p.isAtEnd() && p.peekType() != TokenDo {
		if p.peekType() == TokenLBracket || p.peekType() == TokenLParen {
			depth++
		} else if p.peekType() == TokenRBracket || p.peekType() == TokenRParen {
			depth--
		}
		if depth < 0 {
			break
		}
		p.advance()
	}

	if startPos == p.pos {
		return "", p.errorf("empty condition")
	}

	return p.reconstructTokens(startPos, p.pos), nil
}

func (p *parser) parseExpression() (string, error) {
	startPos := p.pos
	depth := 0

	for !p.isAtEnd() {
		tType := p.peekType()

		// Stop at statement delimiters
		if tType == TokenNewline || tType == TokenEnd || tType == TokenElif ||
			tType == TokenElse || tType == TokenDo {
			break
		}

		// Handle nested brackets/parentheses
		if tType == TokenLBracket || tType == TokenLParen {
			depth++
		} else if tType == TokenRBracket || tType == TokenRParen {
			depth--
			if depth < 0 {
				break
			}
		}

		// Stop at comma if not nested
		if tType == TokenComma && depth == 0 {
			break
		}

		p.advance()
	}

	if startPos == p.pos {
		return "", p.errorf("Empty expression")
	}

	return p.reconstructTokens(startPos, p.pos), nil
}

func (p *parser) skipNewlines() {
	for !p.isAtEnd() && p.peekType() == TokenNewline {
		p.advance()
	}
}

func (p *parser) peek() Token {
	if p.isAtEnd() {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos]
}

func (p *parser) peekType() TokenType {
	return p.peek().Type
}

func (p *parser) advance() Token {
	// Get the current token and advance the position if not at the end.
	token := p.peek()
	if !p.isAtEnd() {
		p.pos++
	}
	return token
}

func (p *parser) isAtEnd() bool {
	// If we've reached the end of the token list or encountered an EOF token, we're at the end.
	return p.pos >= len(p.tokens) || p.peekType() == TokenEOF
}

// errorf creates a parse error with line information.
func (p *parser) errorf(format string, args ...interface{}) error {
	token := p.peek()
	return &ParseError{
		Line:    token.Line,
		Message: fmt.Sprintf(format, args...),
	}
}
