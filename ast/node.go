package ast

type Program struct {
	Statements []Statement
}

type Statement interface {
	statementNode() // marker method to distinguish statement types
	Line() int      // returns the line statement appears in source
}

// name = expression
type AssignmentStmt struct {
	Name       string
	Expression string
	AtLine     int
}

func (s *AssignmentStmt) statementNode() {}
func (s *AssignmentStmt) Line() int      { return s.AtLine }

// Plain command lines.
type CommandStmt struct {
	Name   string
	Args   []string
	AtLine int
}

func (s *CommandStmt) statementNode() {}
func (s *CommandStmt) Line() int      { return s.AtLine }

// return <expression>
type ReturnStmt struct {
	Expression string
	AtLine     int
}

func (s *ReturnStmt) statementNode() {}
func (s *ReturnStmt) Line() int      { return s.AtLine }

// if <condition> ... else ... fi
type IfStmt struct {
	Condition string
	Then      []Statement
	Else      []Statement
	AtLine    int
}

func (s *IfStmt) statementNode() {}
func (s *IfStmt) Line() int      { return s.AtLine }

// while <condition> ... done
type WhileStmt struct {
	Condition string
	Body      []Statement
	AtLine    int
}

func (s *WhileStmt) statementNode() {}
func (s *WhileStmt) Line() int      { return s.AtLine }

// for <var> in [start..end]; do ... done
type ForStmt struct {
	VarName   string
	StartExpr string
	EndExpr   string
	Body      []Statement
	AtLine    int
}

func (s *ForStmt) statementNode() {}
func (s *ForStmt) Line() int      { return s.AtLine }
