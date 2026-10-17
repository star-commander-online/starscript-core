/**
 * Engine wrapper - implements the public StarScriptEngine interface.
 * Coordinates the WASM bridge and delegates command execution to the host invoker.
 */

import {
  StarScriptEngine,
  ExecuteResult,
  ExecuteError,
  ExecuteOptions,
} from "./types.js";
import { WasmBridge } from "./wasmBridge.js";

/**
 * Engine implementation wrapping the WASM bridge.
 */
export class WasmStarScriptEngine implements StarScriptEngine {
  private bridge: WasmBridge;
  private disposed = false;

  constructor(bridge: WasmBridge) {
    this.bridge = bridge;
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
