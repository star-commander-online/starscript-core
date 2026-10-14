/**
 * Engine wrapper - implements the public StarScriptEngine interface.
 * Coordinates the WASM bridge and command registry.
 */

import {
  StarScriptEngine,
  CommandHandler,
  ExecuteResult,
  ExecuteError,
  ExecuteOptions,
} from "./types.js";
import { WasmBridge } from "./wasmBridge.js";
import { CommandRegistry } from "./commandRegistry.js";

/**
 * Checks if a result is an ExecuteError.
 */
function isExecuteError(obj: any): obj is ExecuteError {
  return obj && typeof obj.code === "string" && typeof obj.message === "string";
}

/**
 * Engine implementation wrapping the WASM bridge.
 */
export class WasmStarScriptEngine implements StarScriptEngine {
  private bridge: WasmBridge;
  private registry: CommandRegistry;
  private disposed = false;

  constructor(bridge: WasmBridge) {
    this.bridge = bridge;
    this.registry = bridge.getRegistry();
  }

  registerCommand(name: string, handler: CommandHandler): void {
    this.ensureNotDisposed();
    this.registry.register(name, handler);
  }

  unregisterCommand(name: string): void {
    this.ensureNotDisposed();
    this.registry.unregister(name);
  }

  async execute(
    source: string,
    options?: ExecuteOptions
  ): Promise<ExecuteResult> {
    this.ensureNotDisposed();

    try {
      return await this.bridge.execute(source, options?.maxSteps);
    } catch (error) {
      // Re-throw as-is (should already be ExecuteError from bridge)
      throw error;
    }
  }

  dispose(): void {
    if (this.disposed) {
      return;
    }

    this.disposed = true;
    this.registry.clear();

    // Note: We don't clean up the WASM instance here since it's global.
    // Multiple engines could theoretically share it, though current design
    // is single-engine per WASM module.
  }

  private ensureNotDisposed(): void {
    if (this.disposed) {
      throw new Error("Engine has been disposed");
    }
  }
}
