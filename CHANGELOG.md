# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [0.1.0] - 2026-10-11
### Added
- Token-based recursive descent parser with proper operator precedence
- New `do/end` block syntax for all control structures (if/while/for)
- Full `elif` chain support with proper nesting
- Hash-style (`#`) comments throughout scripts
- Variable interpolation with `$varname` syntax
- Expression evaluation with support for arithmetic, comparison, and logical operators
- Control flow statements:
  - `if condition do ... elif condition do ... else do ... end`
  - `while condition do ... end`
  - `for var in [start..end] do ... end`
  - `return expression`
- Command invocation with argument evaluation
- Variable assignment with automatic type coercion
- Zero-dependency policy: Go standard library only
- WASM bridge with async/await Promise-based API for TypeScript/JavaScript
- Complete test suite covering control flow, math, variable interpolation, and error handling
- TypeScript wrapper package for browser integration
- Error handling with structured error codes (syntax_error, runtime_error, unknown_command, budget_exceeded)
- Execution limits: maximum steps and source size constraints
- Support for scalar values: null, bool, number, string
