/**
 * WASM loading and initialization.
 * Loads the bundled starscript.wasm binary.
 * wasm_exec.js is imported as a side effect in index.ts.
 */

import { WasmExports } from "./types.js";

/**
 * Ensure the Go runtime is available globally.
 * The Go object is created by wasm_exec.js.
 */
function ensureGoRuntime(): void {
  if (typeof (globalThis as any).Go === "undefined") {
    throw new Error(
      "Go runtime not initialized. This should not happen if wasm_exec.js was imported."
    );
  }
}

/**
 * Get the default WASM URL.
 * Resolves relative to this module using import.meta.url.
 */
function getDefaultWasmUrl(): string {
  if (typeof import.meta !== "undefined" && import.meta.url) {
    return new URL("./starscript.wasm", import.meta.url).href;
  }
  throw new Error(
    "import.meta.url not available; provide wasmUrl in EngineOptions"
  );
}

/**
 * Fetch a WASM binary.
 */
async function fetchWasm(url: string): Promise<ArrayBuffer> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to fetch WASM: ${response.statusText}`);
  }
  return await response.arrayBuffer();
}

/**
 * Load the StarScript WASM module and return the bridge API.
 * wasm_exec.js must be imported before calling this (done in index.ts).
 */
export async function loadWasm(wasmUrl?: string): Promise<WasmExports> {
  ensureGoRuntime();

  const url = wasmUrl || getDefaultWasmUrl();

  // Fetch the WASM binary
  const wasmBuffer = await fetchWasm(url);

  // Instantiate the WASM module
  const go = new (globalThis as any).Go();
  const wasmModule = await WebAssembly.instantiate(wasmBuffer, go.importObject);

  // Run the Go program (initializes __starscript on globalThis)
  go.run(wasmModule.instance);

  // Wait for __starscript to be available
  let attempts = 0;
  const maxAttempts = 100; // 1 second timeout (100 * 10ms)

  while (
    typeof (globalThis as any).__starscript === "undefined" &&
    attempts < maxAttempts
  ) {
    await new Promise((resolve) => setTimeout(resolve, 10));
    attempts++;
  }

  if (typeof (globalThis as any).__starscript === "undefined") {
    throw new Error("__starscript API failed to initialize");
  }

  return (globalThis as any).__starscript as WasmExports;
}
