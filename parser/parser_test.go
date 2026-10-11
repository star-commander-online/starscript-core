package parser

import "testing"

func TestParse_ControlFlowScript(t *testing.T) {
	source := `
# Total counter
total = 0
i = 0

# While loop to sum
while $i < 3 do
  total = $total + $i
  i = $i + 1
end

# For loop with echo
for n in [1..2] do
  echo "loop" $n
end

# Conditional
if $total == 3 do
  return "ok"
else do
  return "bad"
end
`

	program, err := Parse(source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(program.Statements) != 5 {
		t.Fatalf("expected 5 top-level statements, got %d", len(program.Statements))
	}
}

func TestParse_IfElif(t *testing.T) {
	source := `
x = 5
if $x == 1 do
  y = "one"
elif $x == 5 do
  y = "five"
else do
  y = "other"
end
`

	program, err := Parse(source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestParse_Comments(t *testing.T) {
	source := `
# This is a comment
x = 1
# Another comment
y = 2
`

	program, err := Parse(source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestParse_BlankLines(t *testing.T) {
	source := `
x = 1

y = 2

z = 3
`

	program, err := Parse(source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(program.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(program.Statements))
	}
}
