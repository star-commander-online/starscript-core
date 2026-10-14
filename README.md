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

const engine = await createEngine();

engine.registerCommand("greet", (ctx) => `Hello, ${ctx.args[0]}!`);

const result = await engine.execute(`
  greeting = greet "World"
  return greeting
`);

console.log(result.value); // "Hello, World!"
engine.dispose();
```

## Features

- **Variables & Expressions**: Arithmetic, string ops, boolean logic
- **Control Flow**: if/elif/else, while, for [start..end]
- **Commands**: Extensible async-first command system
- **Interpolation**: `$varname` syntax in strings/expressions
- **Error Recovery**: Structured errors with line numbers
- **Execution Limits**: Step budget to prevent runaway scripts

## API

```typescript
async function createEngine(options?: EngineOptions): Promise<StarScriptEngine>

interface StarScriptEngine {
  registerCommand(name: string, handler: CommandHandler): void;
  unregisterCommand(name: string): void;
  execute(source: string, options?: ExecuteOptions): Promise<ExecuteResult>;
  dispose(): void;
}

type StarValue = null | boolean | number | string;

interface CommandContext {
  name: string;
  args: StarValue[];
  runId: string;
}

type CommandHandler = (context: CommandContext) => StarValue | Promise<StarValue>;

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

# Variables, expressions, string interpolation
message = "Hello, $name! x is $x"

# Control flow
if x > 10 do
  result = "big"
else do
  result = "small"
end

# While and for loops
count = 0
while count < 5 do
  count = count + 1
end

for i in [1..10] do
  print i
end

# Command invocation
response = fetchData "https://example.com"
return response
```

## Examples

**Async handler:**
```typescript
engine.registerCommand("delay", async (ctx) => {
  const ms = Number(ctx.args[0]);
  await new Promise(r => setTimeout(r, ms));
  return `waited ${ms}ms`;
});

const result = await engine.execute(`
  status = delay 1000
  return status
`);
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

Prerequisites: Go 1.23+, Node.js 18+, TypeScript 5.0+

```bash
cd typescript
npm run build           # Build WASM + TypeScript
npm run publish        # Build and publish to npm
```

## Architecture

**Layer 1 (Go)**: Parser, runtime, value types in `ast/`, `parser/`, `runtime/`, `value/`

**Layer 2 (WASM)**: Async command bridge in `wasm/main.go` – converts Go sync calls to JS async via channels

**Layer 3 (TypeScript)**: Public API in `typescript/src/` – factory, engine, bridge, registry, types

## Performance

- Startup: ~10-50ms (WASM loading)
- Execution: ~1-10ms per script
- Binary: 3.8 MB raw, ~1 MB gzipped
- Memory: ~5-10 MB per engine

## Limitations

- Parser: Infinite recursion issue with deeply nested scripts (flatten as workaround)
- Values: Scalar-only (no arrays/objects)
- Concurrency: Single context per engine (use multiple engines for parallel execution)

## Documentation

- [CONTRIBUTING.md](CONTRIBUTING.md) – Development setup and workflow
- [SYNTAX_EXAMPLES.md](SYNTAX_EXAMPLES.md) – Extended language examples
- [CHANGELOG.md](CHANGELOG.md) – Release history

## Publishing

Update `typescript/package.json`, update `CHANGELOG.md`, then:

```bash
cd typescript && npm run build && npm publish --access public
```

## License

MIT License – see [LICENSE](LICENSE)
