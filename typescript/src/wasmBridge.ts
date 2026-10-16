/**
 * WASM bridge - handles Go ↔ JS communication for the async command protocol.
 */

import { CommandRegistry } from "./commandRegistry.js";
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
 */
export class WasmBridge {
  private wasm: WasmExports;
  private registry: CommandRegistry;
  private pendingRuns: Map<string, AbortController> = new Map();

  constructor(wasm: WasmExports, registry: CommandRegistry) {
    this.wasm = wasm;
    this.registry = registry;
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
      const onCommand = (
        id: string,
        runId: string,
        name: string,
        args: any[],
        rawLine: string
      ) => {
        // Handle command asynchronously
        this.handleCommand(id, runId, name, args, rawLine).catch((error) => {
          // If command handler errors, reject it to the Go side
          this.wasm.reject(id, "command_error", error.message);
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
   * Handle a single command invocation.
   * Looks up the handler in the registry, invokes it, and sends result back to Go.
   */
  private async handleCommand(
    id: string,
    runId: string,
    name: string,
    args: any[],
    rawLine: string
  ): Promise<void> {
    // Check if command is registered
    if (!this.registry.has(name)) {
      this.wasm.reject(id, "unknown_command", `Unknown command: ${name}`);
      return;
    }

    try {
      // Invoke the handler
      const result = await this.registry.invoke({
        name,
        args,
        rawLine,
        runId,
      });

      // Send result back to Go
      this.wasm.resolve(id, result);
    } catch (error) {
      // Send error back to Go
      const message = error instanceof Error ? error.message : String(error);
      this.wasm.reject(id, "command_error", message);
    }
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

  /**
   * Get the command registry.
   * Used for testing and inspection.
   */
  getRegistry(): CommandRegistry {
    return this.registry;
  }
}
