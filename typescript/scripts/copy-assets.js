#!/usr/bin/env node

/**
 * Build script to copy asset files (WASM and wasm_exec.js) to dist folder.
 * Run after TypeScript compilation.
 */

import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const scriptsDir = __dirname;
const typescriptDir = path.resolve(scriptsDir, "..");
const rootDir = path.resolve(typescriptDir, "..");
const distDir = path.join(typescriptDir, "dist");
const srcDir = path.join(typescriptDir, "src");

// Ensure dist directory exists
if (!fs.existsSync(distDir)) {
  fs.mkdirSync(distDir, { recursive: true });
}

// Copy wasm_exec.js from src to dist
const wasmExecSrc = path.join(srcDir, "wasm_exec.js");
const wasmExecDest = path.join(distDir, "wasm_exec.js");

if (fs.existsSync(wasmExecSrc)) {
  fs.copyFileSync(wasmExecSrc, wasmExecDest);
  console.log("✓ Copied wasm_exec.js to dist/");
} else {
  console.warn("⚠ wasm_exec.js not found in src/");
}

// Copy starscript.wasm from root dist to typescript dist
const wasmSrcPath = path.join(rootDir, "dist", "starscript.wasm");
const wasmDestPath = path.join(distDir, "starscript.wasm");

if (fs.existsSync(wasmSrcPath)) {
  fs.copyFileSync(wasmSrcPath, wasmDestPath);
  console.log("✓ Copied starscript.wasm to dist/");
} else {
  console.warn("⚠ starscript.wasm not found at " + wasmSrcPath);
}

console.log("✓ Asset copy complete");
