# StarScript Language Specification

This spec defines the observable behavior of the StarScript Go interpreter. The TypeScript port must match this spec exactly for each test case.

## 1. GRAMMAR AND SYNTAX

### 1.1 Tokens and Keywords

**Keywords (case-sensitive):**
- `if`, `elif`, `else`, `end` – conditionals
- `while`, `do`, `end` – loops
- `for`, `in`, `end` – range loops
- `return` – early exit
- `true`, `false` – boolean literals
- `null` – null literal

**Operators:**
- Arithmetic: `+`, `-`, `*`, `/`
- Comparison: `<`, `<=`, `>`, `>=`, `==`, `!=`, `===`, `!==`
- Logical: `&&`, `||`, `!`
- Assignment: `=`
- Range: `..`
- Variable reference: `$`
- String/block delimiters: `[`, `]`, `(`, `)`
- Separator: `,`

**Special tokens:**
- Comments: `#` to end of line (lexer skips)
- Newline: `\n` (significant; terminates statements)
- EOF: end of input

### 1.2 Statements

A script is a sequence of statements separated by **newlines**. Statements must be on separate lines; semicolons are **not** used as separators. Blank lines (newlines with only whitespace) are skipped.

**Statement types:**

1. **Assignment:**
   ```
   name = expression
   ```
   - `name` must be a valid identifier (letter/underscore + alphanumerics)
   - `expression` is parsed as a string and evaluated
   - Result is stored in the variable map
   - Whitespace around `=` is allowed

2. **Command invocation:**
   ```
   command_name arg1 arg2 ...
   ```
   - First token is command name (identifier)
   - Remaining tokens are arguments (expressions)
   - Arguments are evaluated left-to-right before invocation
   - Result stored in variables (if assignment-like) or ignored

3. **Return:**
   ```
   return expression
   ```
   - Terminates execution with a value
   - Expression is evaluated and returned
   - If no value on that line, defaults to `null`

4. **Conditional:**
   ```
   if condition do
     statements...
   [elif condition do
     statements...]
   [else do
     statements...]
   end
   ```
   - `condition` is evaluated as a truthy/falsy check
   - Only matching branch executes
   - `elif` chains are nested `if/else if`
   - Blocks must be terminated with `end`

5. **While loop:**
   ```
   while condition do
     statements...
   end
   ```
   - Condition is re-evaluated each iteration
   - Loop exits when condition is falsy
   - Blocks must be terminated with `end`

6. **For range loop:**
   ```
   for variable in [start..end] do
     statements...
   end
   ```
   - `variable` is assigned each value from `start` to `end` (inclusive)
   - Both `start` and `end` must evaluate to non-negative integers
   - Loop body has access to updated variable
   - Blocks must be terminated with `end`

### 1.3 Blocks

Blocks are terminated by `end`, `elif`, or `else` keywords. The parser tracks nested bracket/parenthesis depth to avoid premature termination.

Newlines within blocks are **significant** but **optional**; statements are separated by newlines, but empty lines (multiple consecutive newlines) are collapsed.

## 2. EXPRESSIONS

Expressions are parsed with recursive descent and evaluated with standard operator precedence (lowest to highest):

1. **Logical OR** (`||`)
2. **Logical AND** (`&&`)
3. **Equality** (`==`, `!=`, `===`, `!==`)
4. **Comparison** (`<`, `<=`, `>`, `>=`)
5. **Addition/Subtraction** (`+`, `-`)
6. **Multiplication/Division** (`*`, `/`)
7. **Unary** (`!`, `-`, prefix only)
8. **Primary** (literals, variables, parenthesized expressions)

### 2.1 String Interpolation

**Variable reference syntax:** `$variable_name` within expressions.

The interpolation regex is: `\$([A-Za-z_][A-Za-z0-9_]*)`

When a `$var` is found:
1. Look up variable in runtime state
2. If found, convert value to expression literal (quoted strings, formatted numbers, etc.)
3. If **not found**, insert `__UNRESOLVED__variable_name` placeholder
4. After full interpolation, check for `__UNRESOLVED__` – if present, error: `"Unknown variable..."`

**Example:**
```
x = 5
y = "result: " + $x    # "result: " + 5 → "result: 5"
z = $undefined         # Error: Unknown variable "undefined"
```

### 2.2 Operators

**Precedence and associativity:**
- All binary operators are left-associative
- Unary operators (`!`, `-`) are right-associative

**Plus (`+`):**
- If either operand is a string, **concatenate** (both sides coerced to strings via `Stringify()`)
- Else, numeric addition (both operands must be numbers)

**Minus (`-`), Multiplication (`*`), Division (`/`):**
- Both operands must be numbers; error if not
- Division by zero returns error: `"division by zero"`

**Comparison (`<`, `<=`, `>`, `>=`):**
- Both operands must be numbers; error if not
- Returns boolean

**Equality (`==`, `!=`, `===`, `!==`):**
- `==` and `!=`: kind-aware equality (see section 3.2)
- `===` and `!==`: treated as aliases for `==` and `!=` (same behavior)

**Logical AND (`&&`):**
- Short-circuits: if left is falsy, returns boolean `false` without evaluating right
- Returns boolean (truthiness of both sides)

**Logical OR (`||`):**
- Short-circuits: if left is truthy, returns boolean `true` without evaluating right
- Returns boolean (truthiness of either side)

**Unary NOT (`!`):**
- Returns boolean negation of operand's truthiness

**Unary Minus (`-`):**
- Operand must be a number; error if not
- Returns negated number

### 2.3 Literals

- **Numbers:** Floating-point (Go's `float64`). Format: `123`, `123.45`, `.5`, `5.`, or in scientific notation.
- **Strings:** Double-quoted, with escape sequences (`\n`, `\t`, `\r`, `\"`, `\\`).
- **Booleans:** `true`, `false` (keywords).
- **Null:** `null` (keyword).
- **Parentheses:** `( expression )` groups sub-expressions.

### 2.4 Expression Fallback

If expression evaluation fails (e.g., invalid operator, undefined variable, type error), the parser catches the error and **falls back to treating the RHS as a string literal**, storing it as-is (for assignments).

## 3. VALUES AND COERCION

### 3.1 Value Types

StarScript supports four scalar types:
- **Null:** `null`
- **Boolean:** `true` or `false`
- **Number:** IEEE 754 double-precision floating-point (Go's `float64`), including `NaN` and `Inf`
- **String:** UTF-8 text

### 3.2 Equality and Type Checking

Two values are equal if and only if:
1. Their **kinds** (types) are identical, **AND**
2. Their values are equal within that type

Examples:
- `null == null` → `true`
- `true == true` → `true`
- `1 == 1.0` → `true` (both numbers)
- `"1" == 1` → `false` (different kinds)
- `true == 1` → `false` (different kinds)

### 3.3 Truthiness

A value is **truthy** if:
- **Null:** `false`
- **Boolean:** the boolean's value
- **Number:** nonzero and not `NaN`
- **String:** non-empty string

Examples:
- `if 0 do ... end` → condition is **false** (0 is falsy)
- `if "" do ... end` → condition is **false** (empty string is falsy)
- `if "0" do ... end` → condition is **true** (non-empty string is truthy)
- `if $undefined do ... end` → **error** (undefined variable)

### 3.4 Stringify Conversion

Each value converts to string as follows (used in string concatenation and display):
- **Null:** `"null"`
- **Boolean:** `"true"` or `"false"`
- **Number:** `strconv.FormatFloat(v, 'f', -1, 64)` in Go
  - Integers: no decimal point (e.g., `123`, `-1`, `0`)
  - Decimals: minimal representation (e.g., `1.5`, `0.5`)
  - Very small/large: exponential notation (e.g., `1e-10`, `1e+20`)
  - Special: `NaN` → `"NaN"`, `Inf` → `"+Inf"`, `-Inf` → `"-Inf"`
- **String:** as-is

### 3.5 Numeric Coercion

Arithmetic operators (`-`, `*`, `/`) and comparisons (`<`, `<=`, `>`, `>=`) require numeric operands. If an operand is not a number, the operation errors.

### 3.6 For Loop Bounds

For loop start/end expressions must evaluate to:
1. A numeric value
2. Non-negative
3. An integer (no fractional part)

Failure at any step produces a runtime error.

## 4. EXECUTION MODEL

### 4.1 Variable Scoping

Variables are stored in a flat, global map. All assignments and reads within a script access the same namespace. No function-level or block-level scoping exists.

### 4.2 Return Semantics

`return expression` immediately terminates the current block and all parent blocks, propagating the return value to `Execute()`.

If no `return` is executed, the last evaluated statement's value becomes the result.

### 4.3 Step Budgeting

Each **statement** costs 1 step. Before executing any statement, the engine increments the step counter and checks:
- If `steps > maxSteps`, execution halts with `budget_exceeded` error
- If context is cancelled, execution halts with context error

### 4.4 Execution Result

A successful `Execute()` returns:
- `value`: the final result (return value or last statement value)
- `variables`: a snapshot of all variables at completion
- `steps`: total steps executed

## 5. ERRORS

### 5.1 Error Codes

- **`syntax_error`:** Parse failure (invalid syntax, unmatched blocks, etc.)
- **`runtime_error`:** Execution failure (undefined variable, type mismatch, etc.)
- **`unknown_command`:** Command not found in invoker
- **`budget_exceeded`:** Execution step limit exceeded

### 5.2 Error Line Numbers

The error's `line` field is set to the statement's line number (from the AST node). For parse errors, it's the lexer/parser's reported line.

### 5.3 Unknown Variables

If a variable is referenced via `$name` and not found, the error is:
- Code: `runtime_error`
- Message: `Unknown variable "name"`
- Line: the statement line

## 6. COMMANDS

### 6.1 Invocation Syntax

```
command_name arg1 arg2 ...
```

- `command_name` must be an identifier
- Arguments are space-separated expressions
- Arguments are evaluated left-to-right

### 6.2 Argument Evaluation

Each argument expression is:
1. Interpolated (if it contains `$var`)
2. Evaluated (arithmetic, operators, etc.)
3. Converted to a `StarValue` (one of null, bool, number, string)

### 6.3 Command Invocation

The host's `Invoker.Invoke(ctx, Call{Name, Args})` is called. The function receives the command name and evaluated arguments, and returns a `StarValue` or an error.

### 6.4 Error Handling

If the invoker returns an `UnknownCommandError`, the engine maps it to `unknown_command`.

If the invoker returns any other error, it's mapped to `runtime_error`.

### 6.5 Return Value

The command's return value becomes the statement's result (visible to `return` or next statement).

## 7. LIMITS

- **MaxSteps:** Default 10,000. Configurable per `Execute()` call.
- **MaxSourceBytes:** Default 256 KB. Configurable per `Execute()` call.

Both limits are enforced at the start of `Execute()`. Exceeding either produces an error.

## 8. QUIRKS AND SURPRISING BEHAVIORS

1. **Fallback to string on expression error:**
   When assignment evaluation fails (e.g., invalid arithmetic), the RHS is stored as a string literal rather than erroring.
   - Example: `x = 5 + "hello"` → error (operand type mismatch)
   - Example: `x = unknown_var + 1` → error (undefined variable)
   - BUT: `x = "no operators here"` → succeeds, `x` is `"no operators here"`

2. **Variable interpolation is regex-based:**
   Only `$identifier` patterns are replaced. Malformed patterns (e.g., `$123var`) are left as-is.
   - Example: `y = "Price: $price"` → if `price=5`, result is `"Price: 5"`
   - Example: `y = "Var: $123"` → left as `"Var: $123"` (no variable named `123`)

3. **Truthiness of zero strings:**
   The string `"0"` is truthy (it's a non-empty string), but the number `0` is falsy.
   - Example: `if "0" do ... end` → condition is true
   - Example: `if 0 do ... end` → condition is false

4. **NaN and Inf are numbers:**
   Special float values (`NaN`, `Inf`, `-Inf`) are treated as numbers in arithmetic and comparisons.
   - Example: `x = 0 / 0` → `NaN` (division by zero errors, but JS might produce NaN)
   - Example: `if $nan do ... end` → falsy (NaN is falsy)

5. **While and for loops with returns:**
   A `return` inside a loop body immediately exits the loop and all enclosing blocks, not just the loop.

6. **For loop variable shadowing:**
   If a for loop reuses a variable name, the variable is overwritten for each iteration.
   - Example:
     ```
     i = "original"
     for i in [1..2] do
       echo $i
     end
     # After loop, i = 2 (last iteration value)
     ```

7. **Division by zero:**
   Attempting `a / 0` produces a runtime error: `"division by zero"`.

8. **For loop with non-integer bounds:**
   If a for loop bound evaluates to a non-integer (e.g., `1.5`), a runtime error is produced.
   - Example: `for i in [1..2.5] do ... end` → error

9. **Line number tracking:**
   Line numbers in errors refer to the source line of the statement, not the token within the line.

10. **=== and !== operators:**
    Treated as aliases for `==` and `!=`. No difference in behavior (not strict-identity like in JS).

11. **Operator precedence:**
    Follows standard precedence (logical OR < AND < equality < comparison < addition < multiplication < unary < primary).
    - Example: `1 + 2 * 3` → `7` (multiplication first)
    - Example: `true && false || true` → `true` (AND first, then OR)

12. **Short-circuit evaluation:**
    `&&` and `||` do not evaluate the right side if the result is determined by the left.
    - Example: `false && error_call()` → `false` (no error, right side not evaluated)
    - Example: `true || error_call()` → `true` (no error, right side not evaluated)

## 9. NUMBER FORMATTING DIFFERENCES

Go's `strconv.FormatFloat(v, 'f', -1, 64)` produces:
- Integers without decimal: `1`, `-5`, `0`
- Decimals with minimal digits: `1.5`, `0.1`, `3.14159`
- Very large/small in exponential: `1e+20`, `1e-10`

JavaScript's `Number#toString()` is similar but has subtle differences:
- JS may use exponential notation at different thresholds
- Go's `-1` precision strips trailing zeros; JS varies

**TS implementation rule:** Use `Number.prototype.toLocaleString()` with appropriate options, or replicate Go's exact logic.

## 10. TODO/OPEN QUESTIONS

- Are there escape sequences in string literals other than `\n`, `\t`, `\r`, `\"`, `\\`?
- Does the parser support Unicode identifiers?
- What is the exact threshold for exponential notation in number formatting?
