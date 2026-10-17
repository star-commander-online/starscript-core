# Contributing to StarScript Core

## Prerequisites

- Go 1.23.0+
- Node.js 18+
- npm 8+

## Setup

```bash
git clone https://github.com/star-commander/starscript-core.git
cd starscript-core
cd typescript && npm ci && cd ..
```

## Building

**Windows:**
```powershell
$env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
cd typescript && npm run build && cd ..
```

**Linux/macOS:**
```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/starscript.wasm ./wasm
cd typescript && npm run build && cd ..
```

## Testing

```bash
# TO BE ADDED
```

## Workflow

1. Create a branch: `git checkout -b feature/name`
2. Make changes
3. Build and test
4. Commit: `git commit -m "feat: description"`
5. Push and open PR

## Code Style

**Go:** Follow [Effective Go](https://golang.org/doc/effective_go)
```bash
go fmt ./...
go vet ./...
```

**TypeScript:** 2 spaces, strict types, use `const`
```bash
cd typescript && npm run build
```

## Commit Messages

```
feat: new feature
fix: bug fix
docs: documentation
refactor: code refactoring
test: add tests
```

## Publishing

```bash
# For me (Dan) only!
cd typescript
npm run prepack
npm publish --access public
```

## Questions?

- Open GitHub issues
- Check README.md for language guide
- Review existing code for patterns
