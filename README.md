# StarScript Core

A fast, type-safe scripting language with WASM support. Write scripts that run in the browser, Node.js, and Go.

- **Scalar-first design**: Simple values (null, boolean, number, string)
- **Async command protocol**: Native Promise support in command handlers
- **Zero dependencies**: Pure Go interpreter + TypeScript bridge
- **Type-safe API**: Full TypeScript support with .d.ts declarations

## Quick Start

```bash
npm install @star-commander-online/starscript-core
```

```typescript
import { createEngine } from "@star-commander-online/starscript-core";

// Create an invoker function to handle commands
// commandLine has variables interpolated ($x becomes actual value)
// lineNumber is the source line for error reporting if command fails
const invoker = async (id: string, runId: string, commandLine: string, lineNumber: number) => {
  const [command, ...args] = commandLine.split(" ");
  
  if (command === "greet") {
    return `Hello, ${args.join(" ")}!`;
  }
  throw new Error(`Unknown command: ${command}`);
};

const engine = await createEngine({ invoker });

const result = await engine.execute(`
  greeting = greet "World"
  return greeting
`);

console.log(result.value); // "Hello, World!"
engine.dispose();
```

## Features

- **Variables & Expressions**: Arithmetic, string ops, boolean logic
- **Control Flow**: if/else, while, for [start..end]
- **Commands**: Extensible async-first command system (via invoker)
- **Interpolation**: `$varname` syntax in strings/expressions
- **Error Recovery**: Structured errors with line numbers
- **Execution Limits**: Step budget to prevent runaway scripts

## API

```typescript
// Create an engine instance
async function createEngine(options: EngineOptions): Promise<StarScriptEngine>

interface EngineOptions {
  invoker: CommandInvoker;      // Required: handles all commands
  wasmUrl?: string;              // Optional: custom WASM URL
}

// Main engine interface
interface StarScriptEngine {
  execute(source: string, options?: ExecuteOptions): Promise<ExecuteResult>;
  dispose(): void;
}

// Command handler signature (variables already interpolated, includes line number)
type CommandInvoker = (id: string, runId: string, commandLine: string, lineNumber: number) => Promise<StarValue>;

// Value types
type StarValue = null | boolean | number | string;

// Execution results
interface ExecuteResult {
  value: StarValue;
  variables: Record<string, StarValue>;
  steps: number;
}

interface ExecuteError {
  code: string;
  message: string;
  line: number;
}
```

## Script Syntax

```starscript
# Comments start with #
x = 42
name = "Alice"

# Variables and expressions
message = "Hello, $name! x is $x"

# Control flow
if x > 10 do
  result = "big"
else do
  result = "small"
end

# Loops
while x < 100 do
  x = x + 1
end

for i in [1..5] do
  action i
end

# Commands (passed to your invoker)
response = getData "key"
return response
```

**Error handling:**
```typescript
try {
  await engine.execute(`
    x = 10
    result = divide $x 0
    return $result
  `);
} catch (error) {
  const err = error as ExecuteError;
  console.log(`${err.code} at line ${err.line}: ${err.message}`);
  // Example output: "command_error at line 3: Division by zero"
}
```

## Building from Source

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Prerequisites (Go 1.23+, Node.js 18+)
- Build commands
- Testing procedures
- Code style guidelines

## Architecture

**Layer 1 (Go)**: Parser, runtime, value types
- `ast/` – AST node definitions
- `parser/` – Script parser
- `runtime/` – Execution engine
- `value/` – Scalar value representation

**Layer 2 (WASM)**: Async command bridge
- `wasm/main.go` – Converts Go sync calls to JS async via channels

**Layer 3 (TypeScript)**: Public API
- `typescript/src/` – Factory, engine, bridge, types

## Documentation

- [CONTRIBUTING.md](CONTRIBUTING.md) – Development setup, building, testing
- [SYNTAX_EXAMPLES.md](SYNTAX_EXAMPLES.md) – Extended language examples

## License

MIT License – see [LICENSE](LICENSE)
