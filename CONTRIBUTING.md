# Contributing to StarScript Core

Thank you for your interest in contributing to StarScript Core! This document provides guidelines and instructions for getting started with development, building the project, running tests, and submitting changes.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Development Setup](#development-setup)
3. [Project Architecture](#project-architecture)
4. [Development Workflow](#development-workflow)
5. [Building the Project](#building-the-project)
6. [Testing](#testing)
7. [Coding Standards](#coding-standards)
8. [Documentation](#documentation)
9. [Submitting Changes](#submitting-changes)
10. [Common Tasks](#common-tasks)

---

## Code of Conduct

We are committed to providing a welcoming and inspiring community for all. Please be respectful, inclusive, and professional in all interactions.

- Be kind and respectful to others
- Welcome differing opinions and viewpoints
- Focus on constructive feedback
- Report inappropriate behavior to maintainers

---

## Development Setup

### Prerequisites

- **Go 1.23.0+** - Download from [golang.org](https://golang.org)
- **Node.js 18+** - Download from [nodejs.org](https://nodejs.org)
- **npm 8+** - Included with Node.js
- **TypeScript 5.0+** - Will be installed via npm
- **Git** - For version control

### Initial Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/star-commander/starscript-core.git
   cd starscript-core
   ```

2. **Install Node.js dependencies:**

   ```bash
   cd typescript
   npm ci  # Use ci instead of install for reproducible builds
   cd ..
   ```

3. **Verify Go installation:**

   ```bash
   go version
   ```

4. **Run initial build to verify setup:**

   ```bash
   # Build WASM
   $env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm  # Windows
   # GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm  # Linux/macOS

   # Build TypeScript
   cd typescript && npm run build && cd ..

   # Verify
   node phase3-verification.js
   ```

---

## Project Architecture

### Three-Layer Design

```
Layer 1: Go Interpreter Core
  ├── ast/ - Abstract syntax tree nodes
  ├── parser/ - Script parser
  ├── runtime/ - Execution engine
  ├── command/ - Command invocation interface
  └── value/ - Scalar value representation

Layer 2: WASM Bridge (Go → JavaScript)
  └── wasm/main.go - Async command protocol bridge

Layer 3: TypeScript Public API (JavaScript → Consumers)
  ├── typescript/src/index.ts - Public factory
  ├── typescript/src/engine.ts - Engine wrapper
  ├── typescript/src/wasmBridge.ts - Command coordination
  ├── typescript/src/commandRegistry.ts - Handler storage
  ├── typescript/src/wasmLoader.ts - WASM initialization
  └── typescript/src/types.ts - Type definitions
```

### Key Concepts

- **Scalar Values Only**: null, boolean, number, string (no arrays/objects)
- **Step Budget**: Configurable limit to prevent runaway scripts
- **Command Handler**: Sync or async function that processes script commands
- **Execution Pipeline**: Parse → Execute → Command Callbacks → Return Result

---

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or for bug fixes:
git checkout -b fix/your-bug-name
```

### 2. Make Your Changes

- Keep commits small and focused
- Write clear commit messages
- Update documentation as needed

### 3. Build and Test

```bash
# Build both layers
./build.ps1  # Windows (if available)
# or manually:
$env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
cd typescript && npm run build && cd ..

# Run tests
node smoke-test-minimal.js
node phase3-verification.js
```

### 4. Verify Code Quality

```bash
# Go
go fmt ./...
go vet ./...

# TypeScript
cd typescript
npm run build  # Ensures no TypeScript errors
```

### 5. Commit and Push

```bash
git add .
git commit -m "feat: descriptive message"
git push origin feature/your-feature-name
```

### 6. Submit Pull Request

- Open a PR with a clear title and description
- Reference any related issues
- Ensure CI checks pass
- Address review feedback

---

## Building the Project

### Build Go WASM Binary

**Windows (PowerShell):**
```powershell
$env:GOOS='js'
$env:GOARCH='wasm'
go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
```

**Linux/macOS:**
```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
```

**Flags explained:**
- `-ldflags="-s -w"` - Strip debug symbols to minimize binary size
- `GOOS=js GOARCH=wasm` - Cross-compile to JavaScript/WASM target

### Build TypeScript Layer

```bash
cd typescript
npm ci              # Install dependencies
npm run build       # Compile TypeScript to JavaScript
cd ..
```

**Output files:**
- `typescript/dist/*.js` - Compiled ESM modules
- `typescript/dist/*.d.ts` - TypeScript type declarations

### Full Build

```powershell
# Windows
$env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
cd typescript; npm ci; npm run build; cd ..
```

```bash
# Linux/macOS
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
cd typescript && npm ci && npm run build && cd ..
```

### Verify Build Output

```bash
# Check binary exists and has reasonable size
ls -lh dist/starscript.wasm
ls -lh typescript/dist/

# Run verification
node phase3-verification.js
```

---

## Testing

### Test Suites

#### 1. Core Bridge Tests (smoke-test-minimal.js)

Tests the WASM bridge API without executing complex scripts:

```bash
node smoke-test-minimal.js
```

**Tests:**
- WASM binary loads correctly
- Go bridge initialization works
- Command protocol functions (resolve/reject/cancel)
- Error handling

#### 2. Full Verification Suite (phase3-verification.js)

Comprehensive verification of all layers:

```bash
node phase3-verification.js
```

**Checks:**
- All 6 TypeScript files compiled
- Package.json properly configured
- TypeScript strict mode enabled
- All 9 public API symbols exported
- Correct type definitions
- Module dependencies properly wired
- WASM binary and runtime available
- Build artifacts ready for npm

### Running Specific Tests

```bash
# Run only smoke tests
node smoke-test-minimal.js

# Run only verification
node phase3-verification.js

# Run with output
node phase3-verification.js 2>&1 | tee test-output.log
```

### Test Coverage

To check which code paths are tested:

```bash
# Go tests (if test files exist)
go test -v -cover ./...

# TypeScript is verified via compilation
cd typescript
npm run build  # No errors = types are correct
```

### Writing New Tests

When adding features, add corresponding tests:

1. **Go Code**: Add tests in `*_test.go` files
2. **TypeScript Code**: Add tests in `*.test.ts` files
3. **Integration**: Update verification scripts if needed

---

## Coding Standards

### Go Code

Follow [Effective Go](https://golang.org/doc/effective_go) conventions:

```go
// Use clear variable names
var commandName string
var maxSteps int = 10000

// Comments explain the why, not the what
// Execute the script in a separate goroutine to handle async commands
go func() {
  // implementation
}()

// Use errors.Is/errors.As for error handling
if errors.As(err, &parseErr) {
  // handle parse error
}

// Keep functions small and focused
func (b *WasmBridge) handleCommand(id, runId, name string, args []interface{}) error {
  // focused logic
}
```

**Formatting:**
```bash
go fmt ./...    # Auto-format all Go files
go vet ./...    # Check for issues
```

### TypeScript Code

Follow [TypeScript Best Practices](https://www.typescriptlang.org/docs/handbook/2/types-from-types.html):

```typescript
// Use explicit types
interface CommandContext {
  name: string;
  args: StarValue[];
  runId: string;
}

// Prefer interfaces over types for object contracts
interface StarScriptEngine {
  registerCommand(name: string, handler: CommandHandler): void;
  execute(source: string): Promise<ExecuteResult>;
}

// Use async/await instead of .then()
async function loadWasm(url: string): Promise<WasmExports> {
  const response = await fetch(url);
  const buffer = await response.arrayBuffer();
  return initWasm(buffer);
}

// Prefer const over let
const maxSteps = 10000;
```

**Code Style:**
- Indentation: 2 spaces
- Line length: 100 characters (soft limit)
- No unused variables or imports
- Use TSLint/ESLint rules

### Commit Messages

Follow conventional commits:

```
feat: add new feature description
fix: bug fix description
docs: documentation updates
refactor: code refactoring without behavior change
test: add or update tests
chore: maintenance tasks, dependencies
ci: CI/CD configuration changes

Examples:
feat: add async command handler support
fix: resolve WASM loading issue in Node.js
docs: update README with examples
refactor: simplify command registry logic
test: add verification for error handling
```

### Comments and Documentation

```go
// Comment public functions and types
// Execute runs the script with the provided context
func (e *Engine) Execute(ctx context.Context, source string) (*Result, error) {
  // ...
}

// Explain non-obvious logic
// We use a channel to block the goroutine until the JS handler completes,
// allowing sync Go code to wait for async JS operations
cmdResult := <-pendingCmds[cmdID]
```

```typescript
// Document parameters and return types
/**
 * Load and initialize WASM module
 * @param url - URL to WASM binary (auto-detected if omitted)
 * @returns Bridge API for command execution
 */
async function loadWasm(url?: string): Promise<WasmExports> {
  // ...
}
```

---

## Documentation

### README

Update [README.md](README.md) for:
- New features
- API changes
- Build instructions
- Usage examples

### Phase Documentation

Maintain phase completion files:
- [PHASE-2-COMPLETE.md](PHASE-2-COMPLETE.md) - WASM bridge details
- [PHASE-3-COMPLETE.md](PHASE-3-COMPLETE.md) - TypeScript layer details

### API Documentation

- Ensure all public functions have JSDoc/Go comments
- Keep type definitions clear and well-commented
- Document error codes and their meanings

### Inline Comments

Comment the "why" not the "what":

```typescript
// GOOD: Explains the reason
// We explicitly handle file:// URLs in Node.js because
// the fetch API doesn't support them by default
if (url.startsWith("file://")) {
  return fs.readFileSync(url.slice(7));
}

// BAD: Repeats what the code already says
// Check if URL starts with file://
if (url.startsWith("file://")) {
```

---

## Submitting Changes

### Before Submitting

1. **Check for existing issues/PRs** - Avoid duplicate work
2. **Create an issue first** - For major features, discuss before implementing
3. **Keep changes focused** - One feature/fix per PR
4. **Test locally** - Run all tests before pushing
5. **Update documentation** - Keep README and comments in sync

### Pull Request Checklist

- [ ] Branch name follows convention (`feature/*`, `fix/*`)
- [ ] Code follows style guidelines
- [ ] Commit messages are clear and descriptive
- [ ] Tests pass locally (`phase3-verification.js`)
- [ ] Documentation is updated
- [ ] No console errors or warnings
- [ ] Binary size hasn't increased significantly
- [ ] Breaking changes documented

### PR Description Template

```markdown
## Description
Brief summary of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Related Issues
Closes #123

## Testing
- Test cases run successfully
- Verified on [browser/Node.js]
- No regressions observed

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] Code follows guidelines
```

---

## Common Tasks

### Adding a New Command Handler

```typescript
// In your application
engine.registerCommand("myCommand", (ctx) => {
  console.log(`Command: ${ctx.name}`);
  console.log(`Args: ${ctx.args}`);
  console.log(`RunID: ${ctx.runId}`);
  return "result";
});

// In a script
result = myCommand arg1 arg2
return result
```

### Debugging Scripts

```typescript
// Add logging to handlers
engine.registerCommand("debug", (ctx) => {
  console.log("Debug:", ctx);
  return null;
});

// Check execution metrics
const result = await engine.execute(script, { maxSteps: 1000 });
console.log(`Steps used: ${result.steps}/1000`);
console.log(`Variables:`, result.variables);
```

### Profiling Execution

```typescript
// Measure execution time
const start = performance.now();
const result = await engine.execute(script);
const elapsed = performance.now() - start;
console.log(`Execution took ${elapsed}ms`);
console.log(`Steps: ${result.steps}`);
```

### Building for Production

```bash
# Go build
$env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm

# TypeScript build with tree-shaking
cd typescript
npm run build
npm publish --access public
```

### Updating Dependencies

```bash
# Go dependencies
go get -u ./...
go mod tidy

# TypeScript dependencies
cd typescript
npm update
cd ..
```

### Running Local Tests Before Commit

```bash
#!/bin/bash
# Quick pre-commit check

# Go
go fmt ./...
go vet ./...

# TypeScript
cd typescript
npm run build
cd ..

# Run tests
node smoke-test-minimal.js
node phase3-verification.js

echo "Pre-commit checks complete!"
```

---

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: File an issue with reproducible example
- **Features**: Discuss in an issue before implementing
- **Code Review**: Submit a PR and request review

---

## Release Process

### For Maintainers

1. **Update version:**
   ```bash
   # Update typescript/package.json
   "version": "0.2.0"
   ```

2. **Update changelog:**
   ```bash
   # Add to CHANGELOG.md
   ## [0.2.0] - 2026-10-13
   ### Added
   - New feature description
   ### Fixed
   - Bug fix description
   ```

3. **Build and test:**
   ```bash
   $env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
   cd typescript && npm run build && cd ..
   node phase3-verification.js
   ```

4. **Tag release:**
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

5. **Publish to npm:**
   ```bash
   cd typescript
   npm publish --access public
   ```

---

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [WebAssembly](https://webassembly.org/)
- [StarScript Language Guide](README.md#language-guide)
- [Phase 2 WASM Bridge Details](PHASE-2-COMPLETE.md)
- [Phase 3 TypeScript Layer Details](PHASE-3-COMPLETE.md)

---

## Questions?

Feel free to:
1. Check existing GitHub issues and discussions
2. Review phase completion documents
3. Look at example code in test files
4. Open a new issue with questions

Thank you for contributing! 🎉
