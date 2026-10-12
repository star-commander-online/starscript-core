/**
 * WASM loading and initialization.
 * Handles fetching and instantiating the Go WASM binary and runtime.
 */

import type { WasmExports } from "./types.js";

/**
 * Load and initialize the WASM module.
 * @param wasmUrl URL to the starscript.wasm binary
 * @returns Promise resolving to the WASM exports
 * @throws Error if WASM loading or initialization fails
 */
export async function loadWasm(wasmUrl: string | URL): Promise<WasmExports> {
  // Ensure Go runtime is loaded
  await ensureGoRuntime();

  // Fetch and instantiate WASM
  const url = typeof wasmUrl === "string" ? wasmUrl : wasmUrl.href;

  const go = new (globalThis as any).Go();
  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`Failed to fetch WASM: ${response.statusText}`);
  }

  const buffer = await response.arrayBuffer();
  const wasmModule = await WebAssembly.instantiate(buffer, go.importObject);

  // Start the Go runtime
  go.run(wasmModule.instance);

  // Wait for StarScriptWasm to be available
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error("WASM initialization timeout")), 5000)
  );

  const ready = new Promise<WasmExports>((resolve) => {
    const checkWasm = () => {
      const wasm = (globalThis as any).StarScriptWasm;
      if (wasm && typeof wasm.createEngine === "function") {
        resolve(wasm);
      } else {
        setTimeout(checkWasm, 10);
      }
    };
    checkWasm();
  });

  return Promise.race([ready, timeout]);
}

/**
 * Ensure the Go WASM runtime (wasm_exec.js) is loaded.
 */
export async function ensureGoRuntime(): Promise<void> {
  if (typeof (globalThis as any).Go !== "undefined") {
    return; // Already loaded
  }

  // Load wasm_exec.js
  const execUrl = new URL("./wasm_exec.js", import.meta.url).href;
  const script = document.createElement("script");
  script.src = execUrl;
  script.type = "text/javascript";

  return new Promise((resolve, reject) => {
    script.onload = () => resolve();
    script.onerror = () =>
      reject(new Error(`Failed to load wasm_exec.js from ${execUrl}`));
    document.head.appendChild(script);
  });
}

/**
 * Get the default bundled WASM URL.
 */
export function getDefaultWasmUrl(): URL {
  return new URL("./starscript.wasm", import.meta.url);
}
