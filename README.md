# starscript-core

StarScript Core is a standalone interpreter for StarScript.

Core design rule:
- One interpreter implementation exists, written in Go.
- Native Go hosts use it directly.
- Browser/TypeScript hosts use the same code via Go WASM.
- TypeScript package is integration and lifecycle code only.

## Zero-dependency policy

The Go interpreter uses only the Go standard library.
No third-party runtime dependencies are required.

## Language model

Statements are separated by newlines (not semicolons).
Indentation is optional and purely for readability.

Supported statements:
- Assignment: `name = expression`
- Command invocation: `command arg1 arg2 ...`
- Return: `return expression`
- If/elif/else block: `if condition do ... elif condition do ... else do ... end`
- While block: `while condition do ... end`
- For range block: `for var in [start..end] do ... end`

Comments:
- Any line whose first non-whitespace character is `#` is ignored.

Variable interpolation:
- Any $name token is replaced with current variable value before expression evaluation.

Assignment rule:
- Try to evaluate interpolated right-hand side as expression.
- If evaluation fails, store interpolated text as a string value.

## Repository map

- ast/
  - Statement structs for parsed scripts.
  - No execution logic.
- parser/
  - Line-oriented parser that produces ast.Program.
  - Owns block matching for if/else/fi and while/for done.
- runtime/
  - Main interpreter engine.
  - Statement execution, variable state, command calls, limits.
  - Internal expression tokenizer/parser/evaluator (no external libs).
- command/
  - Host command call interfaces and unknown-command typing.
- value/
  - Cross-platform scalar/JSON value representation.
- wasm/
  - JS/WASM bridge layer around runtime.Engine.
- typescript/
  - WASM loader, command bridge, engine lifecycle wrappers.

## How execution works internally

The runtime path is intentionally simple and explicit:

1. runtime.NewEngine creates engine state:
- vars map for persistent variables.
- host command invoker.
- execution limits.

2. runtime.Engine.Execute does:
- Source size check.
- parser.Parse to build ast.Program.
- Sequential statement execution.
- Early exit on return.
- Structured result assembly.

3. parser.Parse does:
- Split source into line records with original line numbers.
- Parse top-level statements.
- Recursively parse nested blocks.
- Emit parse errors with exact line.

4. runtime.execStatement dispatches by statement type:
- AssignmentStmt
- CommandStmt
- ReturnStmt
- IfStmt
- WhileStmt
- ForStmt

5. Expression evaluation pipeline:
- Interpolate $variables into expression-safe literals.
- runtime/evalExpression lexes expression text.
- Recursive descent parser builds local expression nodes.
- Evaluator computes final scalar value.

6. Command execution pipeline:
- Command args are interpolated/evaluated where possible.
- command.Invoker.Invoke is called with name + typed args.
- Unknown command errors map to ErrUnknownCommand.

## Runtime limits and safety

Current enforced limits:
- Maximum source size in bytes.
- Maximum executed steps.
- Context cancellation checks between steps.

Current constraints:
- Scalar-focused script semantics (null/bool/number/string).
- Array/object script literals are intentionally not part of simplified mode.

## Error model

Runtime returns typed ScriptError codes:
- syntax_error
- runtime_error
- unknown_command
- budget_exceeded

Errors include line numbers where available.

## WASM and TypeScript bridge

The WASM module exposes:
- createEngine
- execute
- disposeEngine

TypeScript installs command bridge callback:
- globalThis.__starscriptInvokeCommand(engineId, commandName, args)

When script executes a command:
- Go calls JS bridge.
- JS handler may return value or Promise.
- Go waits for Promise resolution and resumes in-order.

## Technical walkthrough by file

- parser/parser.go
  - Parse(source): entrypoint.
  - parseStatements: loop until EOF or stop marker.
  - parseIf/parseWhile/parseFor: block statement parsers.
  - splitCommandLine: command argument tokenizer with quoted-arg handling.

- runtime/engine.go
  - Execute: orchestration and lifecycle.
  - execStatement/execBlock: statement dispatch and recursion.
  - evalExpression/evalCondition: interpolation + eval semantics.
  - invokeCommand: host call and error mapping.
  - step: step budget + cancellation guard.

- runtime/expression.go
  - lexExpr: expression tokenizer.
  - exprParser: precedence-aware parser.
  - evalExprNode/evalBinaryNode: scalar evaluator.

## Build and test

Go:

```bash
go test ./...
go vet ./...
```

WASM build:

```bash
GOOS=js GOARCH=wasm go build -o dist/starscript.wasm ./wasm
```

TypeScript wrapper:

```bash
cd typescript
npm ci
npm run typecheck
npm test
npm run build
```

## Native usage example

```go
engine := runtime.NewEngine(hostInvoker{}, runtime.Options{MaxSteps: 5000})
result, err := engine.Execute(context.Background(), `
count = 0
while $count < 3
  echo "tick" $count
  count = $count + 1
done
return $count
`)
if err != nil {
	panic(err)
}
fmt.Println(result.Value.Stringify())
```

## TypeScript usage example

```ts
const engine = await createEngine({ wasmUrl: "/starscript.wasm", instantiate: initWasm });
engine.registerCommand("echo", ({ args }) => {
  console.log(args);
  return null;
});

const result = await engine.execute([
  "name = \"Commander\"",
  "echo $name",
  "return $name",
].join("\n"));
```

## Current gaps

- Script-to-script recursion depth enforcement not yet implemented in core.
- Full historical Star Commander syntax parity still in progress.
- Conformance corpus still small and will be expanded.

## Publishing and Release

### Versioning
- Uses semantic versioning (major.minor.patch)
- Update version in:
  - `go.mod` (implicit, tagged via git)
  - `typescript/package.json` (must match)
  - `CHANGELOG.md` (before release)

### Creating a Release
1. Update version in `typescript/package.json`
2. Update `CHANGELOG.md` with version and changes
3. Tag the release in git:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. Publish TypeScript package to npm:
   ```bash
   cd typescript
   npm run build
   npm publish --access public
   ```
5. GitHub Actions will automatically build and attach WASM binary to the release

### Using in Your React Client

#### 1. Install the package

```bash
npm install @star-commander-online/starscript-core
```

#### 2. Set up the engine

Create a wrapper hook (`useStarScript.ts`):

```typescript
import { createEngine } from "@star-commander-online/starscript-core";
import { useEffect, useRef } from "react";

export function useStarScript() {
  const engineRef = useRef<any>(null);
  const commandsRef = useRef<Map<string, Function>>(new Map());

  useEffect(() => {
    async function initEngine() {
      engineRef.current = await createEngine({
        wasmUrl: new URL("@star-commander-online/starscript-core/dist/starscript.wasm", import.meta.url).href,
      });
    }
    initEngine();
    return () => {
      engineRef.current?.dispose();
    };
  }, []);

  const registerCommand = (name: string, handler: (args: any[]) => Promise<any> | any) => {
    commandsRef.current.set(name, handler);
  };

  const execute = async (script: string) => {
    if (!engineRef.current) throw new Error("Engine not initialized");
    return engineRef.current.execute(script);
  };

  return { execute, registerCommand };
}
```

#### 3. Use in a React component

```typescript
import { useStarScript } from "./useStarScript";

export function CommandExecutor() {
  const { execute, registerCommand } = useStarScript();

  useEffect(() => {
    registerCommand("echo", ({ args }) => {
      console.log(...args);
      return null;
    });

    registerCommand("delay", async ({ args }) => {
      const ms = Number(args[0]);
      await new Promise(r => setTimeout(r, ms));
      return `delayed ${ms}ms`;
    });
  }, [registerCommand]);

  const handleRun = async () => {
    const script = `
# Example script
message = "Hello from StarScript"
echo $message
delay 1000
return $message
    `;

    try {
      const result = await execute(script);
      console.log("Result:", result.value);
      console.log("Variables:", result.variables);
    } catch (err) {
      console.error("Script error:", err);
    }
  };

  return <button onClick={handleRun}>Run Script</button>;
}
```

#### 4. Build configuration

Ensure your bundler (Vite, Webpack, etc.) can handle WASM:

**Vite (vite.config.ts):**
```typescript
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  optimizeDeps: {
    exclude: ["@star-commander-online/starscript-core"],
  },
});
```

**Next.js (next.config.js):**
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  webpack: (config) => {
    config.experiments = { ...config.experiments, asyncWebAssembly: true };
    config.module.rules.push({
      test: /\.wasm$/,
      type: "webassembly/async",
    });
    return config;
  },
};

module.exports = nextConfig;
```

#### 5. Environment setup

Make sure the WASM binary is accessible:
- Copy to `public/` folder for static serving
- Update `wasmUrl` in useStarScript to point to it
- Or use import URL resolution as shown above
