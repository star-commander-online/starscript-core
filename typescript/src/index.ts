/**
 * Public API for StarScript WASM engine.
 */

import { WasmStarScriptEngine } from "./engine.js";
import { loadWasm, getDefaultWasmUrl } from "./wasm-loader.js";
import type {
  StarScriptEngine,
  EngineOptions,
  ExecuteResult,
  ExecuteOptions,
  StarValue,
  StarCommand,
  CommandContext,
  ExecuteError,
} from "./types.js";

// Export public types
export type {
  StarScriptEngine,
  EngineOptions,
  ExecuteResult,
  ExecuteOptions,
  ExecuteError,
  StarValue,
  StarCommand,
  CommandContext,
};

/**
 * Create a new StarScript engine instance.
 * Automatically loads and initializes the bundled WASM binary.
 *
 * @param options Engine configuration (optional)
 * @returns Promise resolving to an initialized engine
 * @throws Error if WASM loading or initialization fails
 */
export async function createEngine(options?: EngineOptions): Promise<StarScriptEngine> {
  // Determine WASM URL: use custom or default bundled
  const wasmUrl = options?.wasmUrl || getDefaultWasmUrl();

  // Load and initialize WASM
  const wasm = await loadWasm(wasmUrl);

  // Create engine instance in Go
  const engineId = wasm.createEngine({
    maxSteps: options?.maxSteps,
    maxSourceBytes: options?.maxSourceBytes,
  });

  // Return TypeScript wrapper
  return new WasmStarScriptEngine(wasm, engineId);
}
