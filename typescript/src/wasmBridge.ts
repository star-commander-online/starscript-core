/**
 * WASM bridge - handles Go ↔ JS communication for the async command protocol.
 * Passes raw command lines to the host invoker for full parsing and execution.
 */

import { WasmExports, ExecuteError, ExecuteResult, StarValue } from "./types.js";

/**
 * Internal result from Go execution.
 */
interface WasmExecutionResult {
  value: StarValue;
  variables: Record<string, StarValue>;
  steps: number;
}

/**
 * Wraps the WASM bridge API and coordinates command invocation.
 * Passes raw command lines to the host invoker.
 */
export class WasmBridge {
  private wasm: WasmExports;
  private invoker: (id: string, runId: string, rawLine: string) => Promise<StarValue>;
  private pendingRuns: Map<string, AbortController> = new Map();

  constructor(
    wasm: WasmExports,
    invoker: (id: string, runId: string, rawLine: string) => Promise<StarValue>
  ) {
    this.wasm = wasm;
    this.invoker = invoker;
  }

  /**
   * Execute a script asynchronously.
   * Returns a Promise that resolves with the execution result or rejects with an error.
   */
  async execute(
    source: string,
    maxSteps?: number
  ): Promise<ExecuteResult> {
    return new Promise((resolve, reject) => {
      // Set up callbacks for the Go bridge
      const onCommand = (id: string, runId: string, rawLine: string) => {
        
        // Handle command asynchronously - MUST call both .then() and .catch()
        this.invoker(id, runId, rawLine)
          .then((result) => {
            this.wasm.resolve(id, result);
          })
          .catch((error) => {
            // If command handler errors, reject it to the Go side
            const message = error instanceof Error ? error.message : String(error);
            this.wasm.reject(id, "command_error", message);
          });
      };

      const onDone = (result: any) => {
        // Clean up the abort controller
        this.pendingRuns.delete((result.runId as string) || "");

        // Check if result is an error
        if (result.code) {
          reject(result as ExecuteError);
        } else {
          resolve(result as ExecuteResult);
        }
      };

      // Build options object for Go bridge
      const opts = maxSteps ? { maxSteps } : undefined;

      // Start execution and get runId
      const runId = this.wasm.execute(source, onCommand, onDone, opts);

      // Store abort controller for this run
      const controller = new AbortController();
      this.pendingRuns.set(runId, controller);
    });
  }

  /**
   * Cancel an in-flight execution.
   */
  cancel(runId: string): void {
    this.wasm.cancel(runId);
    this.pendingRuns.delete(runId);
  }

  /**
   * Get the underlying WASM exports.
   * Used primarily for testing.
   */
  getWasm(): WasmExports {
    return this.wasm;
  }
}
