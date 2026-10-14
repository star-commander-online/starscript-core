import { execSync } from "node:child_process";
import { copyFileSync, existsSync } from "node:fs";
import { join } from "node:path";

const goRoot = execSync("go env GOROOT", { encoding: "utf-8" }).trim();
const distDir = "dist";

// Try lib/wasm first (Go 1.20+), fall back to misc/wasm
const wasmExecPaths = [
  join(goRoot, "lib", "wasm", "wasm_exec.js"),
  join(goRoot, "misc", "wasm", "wasm_exec.js"),
];

const wasmExecSrc = wasmExecPaths.find((p) => existsSync(p));
if (!wasmExecSrc) {
  throw new Error(`wasm_exec.js not found in Go root: ${goRoot}`);
}

console.log(`[build] Building WASM...`);
execSync(
  `powershell -Command "$env:GOOS='js'; $env:GOARCH='wasm'; go build -ldflags='-s -w' -o ${distDir}/starscript.wasm ../wasm"`,
  { stdio: "inherit" }
);

console.log(`[build] Copying wasm_exec.js from ${wasmExecSrc}...`);
copyFileSync(wasmExecSrc, join(distDir, "wasm_exec.js"));

console.log(`[build] Compiling TypeScript...`);
execSync("tsc", { stdio: "inherit" });

console.log(`[build] Done!`);
