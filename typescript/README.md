# StarScript Core - TypeScript WASM Bridge

StarScript is a lightweight, type-safe scripting language with a Go interpreter and TypeScript WASM bridge. This package provides a JavaScript/TypeScript API to execute StarScript programs, delegating all command handling to your application's invoker.

## Installation

```bash
npm install @star-commander-online/starscript-core
```

## Quick Start

```typescript
import { createEngine } from '@star-commander-online/starscript-core';

// Define a command invoker that handles all command parsing and execution
const invoker = async (id, runId, rawLine) => {
  // Your application handles parsing and executing the raw command line
  console.log(`Executing: ${rawLine}`);
  
  // Parse and execute the command
  if (rawLine.startsWith('greet')) {
    // Resolve with the result
    return `Hello from ${rawLine}`;
  }
  
  // Throw if command not found
  throw new Error(`Unknown command: ${rawLine}`);
};

// Create an engine with your invoker
const engine = await createEngine({ invoker });

// Execute a script
const result = await engine.execute('greet "World"');
console.log(result.value); // "Hello from greet \"World\""

// Dispose when done
engine.dispose();
```

## Core Concepts

### StarScript Values

StarScript uses four scalar types: `null`, `boolean`, `number` (float64), and `string`.

```typescript
export type StarValue = null | boolean | number | string;
```

### Command Invoker

Instead of registering individual command handlers, you provide a single `CommandInvoker` function that receives raw command lines and executes them. This gives you complete control over command parsing, routing, and execution.

```typescript
export type CommandInvoker = (
  id: string,           // Unique command ID
  runId: string,        // Execution context ID
  rawLine: string       // The full command line to execute
) => Promise<StarValue>;
```

The invoker is called for every command encountered during script execution. It must:
- Parse the raw line however you see fit (regex, lexer, command patterns, etc.)
- Execute the command
- Return a StarValue
- Throw an error if the command fails

### Command Context

Each command invocation receives:
- `rawLine: string` - The original line from the script source
- `runId: string` - The ID of this script execution run

### Execution Results

Scripts return an `ExecuteResult` with the final value, variable state, and step count:

```typescript
const result = await engine.execute('x = 5; return $x * 2');
console.log(result.value);      // 10
console.log(result.variables);  // { x: 5 }
console.log(result.steps);      // execution steps count
```

## Language Features

StarScript supports:
- **Variables**: `name = value`
- **Operators**: Arithmetic, comparison, logical (OR, AND)
- **Control flow**: `if/else`, `while`, `for` loops
- **String interpolation**: `"Hello $name"` with variable substitution
- **Commands**: Any line that starts with an identifier (non-keyword)
- **Return statements**: Early exit with values

Example script:
```
x = 10
if $x > 5 do
  message = "Large"
else do
  message = "Small"
fi
cmd arg1 arg2
return $message
```

## Error Handling

Execution failures return an `ExecuteError`:

```typescript
try {
  await engine.execute('unknown_command arg1');
} catch (error) {
  console.log(error.code);    // "unknown_command" or other error code
  console.log(error.message); // Error description
  console.log(error.line);    // Source line number
}
```

Common error codes: `parse_error`, `runtime_error`, `budget_exceeded`

## Execution Options

Control script execution with options:

```typescript
const result = await engine.execute(source, {
  maxSteps: 50000  // Limit execution steps (default: 10,000)
});
```

## Advanced Invoker Example

```typescript
const invoker = async (id, runId, rawLine) => {
  // Custom command parser
  const [cmd, ...args] = rawLine.split(/\s+/);
  
  // Route to different handlers
  switch (cmd) {
    case 'add':
      return parseInt(args[0]) + parseInt(args[1]);
    
    case 'fetch':
      const response = await fetch(args[0]);
      return response.ok ? 'success' : 'failed';
    
    case 'greet':
      return `Hello, ${args.join(' ')}!`;
    
    default:
      throw new Error(`Unknown command: ${cmd}`);
  }
};

const engine = await createEngine({ invoker });
```

## API Reference

### `createEngine(options): Promise<StarScriptEngine>`
Creates a new engine instance.

**Options:**
- `invoker: CommandInvoker` - Required. Function that handles command execution
- `wasmUrl?: string` - Optional. URL to starscript.wasm (auto-detected if omitted)

### `StarScriptEngine`
- `execute(source: string, options?: ExecuteOptions): Promise<ExecuteResult>`
- `dispose(): void`

### Types
- `StarValue` - Scalar value type (null | boolean | number | string)
- `CommandContext` - Command info with `rawLine` and `runId`
- `CommandInvoker` - Function signature: `(id, runId, rawLine) => Promise<StarValue>`
- `ExecuteResult` - Execution output with `value`, `variables`, `steps`
- `ExecuteError` - Error object with `code`, `message`, `line`

## License

MIT - See LICENSE file in repository
