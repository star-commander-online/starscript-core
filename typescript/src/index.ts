/**
 * Public API for the StarScript WASM bridge.
 * This is the main entry point for consumers.
 */

// Load the Go WASM runtime (side effect - must happen first)
import "./wasm_exec.js";

export {
  StarValue,
  CommandContext,
  CommandHandler,
  ExecuteResult,
  ExecuteError,
  EngineOptions,
  ExecuteOptions,
  StarScriptEngine,
} from "./types";

import { loadWasm } from "./wasmLoader.js";
import { CommandRegistry } from "./commandRegistry.js";
import { WasmBridge } from "./wasmBridge.js";
import { WasmStarScriptEngine } from "./engine.js";
import { StarScriptEngine, EngineOptions } from "./types.js";

/**
 * Create and initialize a new StarScript engine instance.
 * Loads the WASM module and prepares the engine for script execution.
 * Must be called once before using the engine.
 *
 * The wasm_exec.js runtime must be loaded before calling this function.
 *
 * @param options - Engine options (wasmUrl, etc.)
 * @returns Promise that resolves with a ready-to-use engine
 */
export async function createEngine(
  options?: EngineOptions
): Promise<StarScriptEngine> {
  // Load the WASM module
  const wasm = await loadWasm(options?.wasmUrl);

  // Create the command registry
  const registry = new CommandRegistry();

  // Create the bridge
  const bridge = new WasmBridge(wasm, registry);

  // Create and return the engine
  return new WasmStarScriptEngine(bridge);
}
