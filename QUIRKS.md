# StarScript Behavioral Quirks and Surprises

This list documents observable behaviors in the Go interpreter that are surprising, unintended-looking, or worth verifying during porting.

## Critical Quirks (High Priority)

### 1. Expression Evaluation Fallback
**Description:** When an assignment's RHS fails to parse/evaluate as an expression, it's stored as a **string** instead of erroring.

**Test case:**
```go
engine.Execute(ctx, `
  x = 1 + 2
  y = 1 + "hello"  # Type error, but stored as string
  return $y
`)
```
**Expected:** Result value is `"1 + \"hello\""`
**Why it's odd:** Silent coercion to string masks typos and errors

### 2. Undefined Variable in Assignment
**Description:** Referencing an undefined variable via `$var` in an assignment produces a runtime error, not a fallback.

**Test case:**
```go
engine.Execute(ctx, `
  x = $undefined
  return "done"
`)
```
**Expected:** Error code `runtime_error`, message `Unknown variable "undefined"`, line number of assignment

### 3. Variable Interpolation Regex
**Description:** Only `$[A-Za-z_][A-Za-z0-9_]*` patterns are interpolated. Malformed patterns are left as-is.

**Test case:**
```go
engine.Execute(ctx, `
  price = 5
  msg = "Price: $price; Qty: $123abc"
  return $msg
`)
```
**Expected:** `"Price: 5; Qty: $123abc"` (second `$` is left as-is)

### 4. Truthy Strings vs Numbers
**Description:** String `"0"` is truthy (non-empty), but number `0` is falsy.

**Test case:**
```go
engine.Execute(ctx, `
  if "0" do
    a = "string 0 is truthy"
  else do
    a = "string 0 is falsy"
  end
  
  if 0 do
    b = "number 0 is truthy"
  else do
    b = "number 0 is falsy"
  end
  
  return $a + " and " + $b
`)
```
**Expected:** `"string 0 is truthy and number 0 is falsy"`

### 5. For Loop Variable Shadowing
**Description:** A for loop's iteration variable overwrites any prior variable with the same name.

**Test case:**
```go
engine.Execute(ctx, `
  i = "before"
  for i in [1..2] do
    x = $i
  end
  return $i
`)
```
**Expected:** Result is `2` (number, not string)

### 6. Division by Zero
**Description:** `a / 0` produces a runtime error, not `Inf` or `NaN`.

**Test case:**
```go
engine.Execute(ctx, `
  x = 5 / 0
  return "done"
`)
```
**Expected:** Error code `runtime_error`, message contains "division by zero"

## Medium-Priority Quirks

### 7. While Loop Early Return
**Description:** A `return` inside a while loop exits the loop and all parent blocks, not just the loop.

**Test case:**
```go
engine.Execute(ctx, `
  i = 0
  while $i < 10 do
    i = $i + 1
    if $i == 2 do
      return $i
    end
  end
  return "done"
`)
```
**Expected:** Result is `2` (not `"done"`)

### 8. For Loop Non-Integer Bounds
**Description:** If a for loop bound is a non-integer, a runtime error is produced.

**Test case:**
```go
engine.Execute(ctx, `
  for i in [1..2.5] do
    x = $i
  end
  return "done"
`)
```
**Expected:** Error code `runtime_error`, message about integer bounds

### 9. For Loop Range Inclusivity
**Description:** The range `[start..end]` is **inclusive** on both ends.

**Test case:**
```go
engine.Execute(ctx, `
  sum = 0
  for i in [1..3] do
    sum = $sum + $i
  end
  return $sum
`)
```
**Expected:** Result is `6` (1 + 2 + 3)

### 10. NaN and Inf Truthiness
**Description:** `NaN` is falsy (not truthy despite being a number).

**Test case:**
```go
# Note: Go's math operations avoid producing NaN directly in most cases
# This test assumes a command returns NaN
engine.Execute(ctx, `
  nanVal = 0.0
  if $nanVal do
    result = "0.0 is truthy"
  else do
    result = "0.0 is falsy"
  end
  return $result
`)
```
**Expected:** `"0.0 is falsy"` (NaN is falsy like `0`)

## Low-Priority Quirks (Cosmetic/Edge Cases)

### 11. === and !== Operators
**Description:** Aliases for `==` and `!=`. No difference in behavior (not strict identity like JavaScript).

**Test case:**
```go
engine.Execute(ctx, `
  x = (1 === 1)
  y = (1 == 1)
  z = (1 === "1")
  return $x + " " + $y + " " + $z
`)
```
**Expected:** `"true true false"` (both === and == behave the same)

### 12. Operator Precedence
**Description:** Standard C-like precedence. Confirm in TS implementation.

**Test case:**
```go
engine.Execute(ctx, `
  a = 1 + 2 * 3
  b = (1 + 2) * 3
  c = true && false || true
  return $a + " " + $b + " " + $c
`)
```
**Expected:** `"7 9 true"` (* before +; && before ||)

### 13. Short-Circuit Evaluation
**Description:** `&&` and `||` do not evaluate the right side if the result is determined by the left.

**Test case (assumes `error` command exists):**
```go
engine.Execute(ctx, `
  # false && error() → no error, right side not evaluated
  a = false && error
  
  # true || error() → no error, right side not evaluated  
  b = true || error
  
  return "ok"
`)
```
**Expected:** Result is `"ok"` (no errors thrown)

### 14. Return Value from Command Invocation
**Description:** A command's return value becomes the statement's value, visible in next assignment.

**Test case:**
```go
engine.Execute(ctx, `
  x = mycommand 5
  y = $x + 1
  return $y
`)
```
**Expected (with mock `mycommand` returning 10):** Result is `11`

### 15. Line Number Accuracy
**Description:** Error line numbers refer to the statement line, not token position.

**Test case:**
```go
engine.Execute(ctx, `
  x = 1
  if $undefined_var do     # Line 2
    y = 2
  end
  return "done"
`)
```
**Expected:** Error line is `2`

## Verification Checklist for TS Port

- [ ] Expression fallback to string on error
- [ ] Undefined variable reference errors immediately (no fallback)
- [ ] Variable interpolation regex `$[A-Za-z_][A-Za-z0-9_]*`
- [ ] String `"0"` is truthy, number `0` is falsy
- [ ] For loop variable shadowing updates outer scope
- [ ] Division by zero errors (not NaN)
- [ ] While/for loop `return` exits all parent blocks
- [ ] For loop bounds must be non-negative integers
- [ ] For loop range is inclusive `[start..end]`
- [ ] NaN is falsy
- [ ] `===` and `!==` behave like `==` and `!=`
- [ ] Operator precedence is correct
- [ ] Short-circuit evaluation for `&&` and `||`
- [ ] Command return values propagate to next statement
- [ ] Error line numbers are accurate
