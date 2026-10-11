# StarScript Syntax Examples - New do/end Format

## Variable Assignment and Expressions

```
# Basic assignment
x = 10
name = "Alice"
flag = true

# With expressions
sum = 5 + 3
result = $x * 2
```

## Conditionals (if/elif/else)

```
# Simple if/else
age = 25
if $age < 18 do
  status = "minor"
else do
  status = "adult"
end

# With elif chain
score = 85
if $score >= 90 do
  grade = "A"
elif $score >= 80 do
  grade = "B"
elif $score >= 70 do
  grade = "C"
else do
  grade = "F"
end
```

## Loops

### While Loop

```
count = 0
sum = 0

while $count < 10 do
  sum = $sum + $count
  count = $count + 1
end

return $sum
```

### For Loop

```
# Iterate from 1 to 5
for i in [1..5] do
  echo "Iteration" $i
end

# Nested loop
for x in [1..3] do
  for y in [1..3] do
    echo "x=" $x "y=" $y
  end
end
```

## Comments

```
# This is a single-line comment
x = 10  # Comments can appear inline

# Multiple comment lines
# explaining complex logic
# work fine too

while $x > 0 do
  x = $x - 1  # Decrement
end
```

## Commands

```
# Simple command
echo "Hello, World!"

# Command with multiple arguments
echo "The answer is" 42

# Command with variable interpolation
name = "Alice"
echo "Hello," $name

# Command with expressions
x = 5
result = 10
echo "x =" $x "result =" $result
```

## Return Statements

```
# Return from script
if $x == 0 do
  return "zero"
end

return "not zero"
```

## Complete Example: FizzBuzz

```
# FizzBuzz implementation
for n in [1..15] do
  output = ""
  
  if $n % 3 == 0 do
    output = $output "Fizz"
  end
  
  if $n % 5 == 0 do
    output = $output "Buzz"
  end
  
  if $output == "" do
    output = $n
  end
  
  echo $output
end

return "done"
```

## Key Syntax Rules

1. **No semicolons** - statements end with newlines
2. **No indentation required** - indentation is optional and only for readability
3. **`do` and `end`** - all blocks use these keywords:
   - `if condition do ... end`
   - `while condition do ... end`
   - `for var in [start..end] do ... end`
4. **`elif` for chained conditionals** - supports multiple elif branches
5. **`#` for comments** - from # to end of line
6. **Variables** - referenced with `$varname` in expressions
7. **Expressions** - follow standard operator precedence
   - Logical: `||`, `&&`, `!`
   - Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
   - Arithmetic: `+`, `-`, `*`, `/`
