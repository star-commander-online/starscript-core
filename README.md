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
const invoker = async (id: string, runId: string, rawLine: string) => {
  const [command, ...args] = rawLine.split(" ");
  
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

// Command handler signature
type CommandInvoker = (id: string, runId: string, rawLine: string) => Promise<StarValue>;

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
  await engine.execute("invalid !!!!");
} catch (error) {
  const err = error as ExecuteError;
  console.log(`${err.code} at line ${err.line}: ${err.message}`);
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

## Performance

- Startup: ~10-50ms (WASM loading)
- Execution: ~1-10ms per script
- Binary: ~3.8 MB raw, ~1 MB gzipped
- Memory: ~5-10 MB per engine

## Limitations

- Scalar-only values (no arrays/objects)
- Single execution context per engine
- Parser: Avoid deeply nested expressions

## Documentation

- [CONTRIBUTING.md](CONTRIBUTING.md) – Development setup, building, testing
- [SYNTAX_EXAMPLES.md](SYNTAX_EXAMPLES.md) – Extended language examples
- [CHANGELOG.md](CHANGELOG.md) – Release history

## Publishing

```bash
cd typescript && npm run build && npm publish --access public
```

Update `typescript/package.json` version and `CHANGELOG.md` first.

## License

MIT License – see [LICENSE](LICENSE)
