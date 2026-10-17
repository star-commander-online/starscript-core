/**
 * Public type definitions for the StarScript WASM bridge.
 * Consumers should only import from this module.
 */

/**
 * A StarScript value.
 * Values are scalar: null, boolean, number (float64), or string.
 */
export type StarValue = null | boolean | number | string;

/**
 * Command information passed to the host invoker.
 * The host is responsible for parsing and executing the raw line.
 */
export interface CommandContext {
  rawLine: string; // The original command line from source
  runId: string; // The execution ID that issued this command
}

/**
 * Result of script execution.
 */
export interface ExecuteResult {
  value: StarValue;
  variables: Record<string, StarValue>;
  steps: number;
}

/**
 * Structured error from script execution.
 */
export interface ExecuteError {
  code: string; // e.g., "parse_error", "runtime_error", "budget_exceeded"
  message: string;
  line: number; // 0 if not applicable
}

/**
 * Command invoker function - responsible for parsing and executing raw command lines.
 * Receives the command line and must return the result or throw an error.
 */
export type CommandInvoker = (
  id: string,
  runId: string,
  rawLine: string
) => Promise<StarValue>;

/**
 * Options for engine creation.
 */
export interface EngineOptions {
  wasmUrl?: string; // URL to starscript.wasm; auto-detected if omitted
  invoker: CommandInvoker; // Required: handles command parsing and execution
}

/**
 * Options for script execution.
 */
export interface ExecuteOptions {
  maxSteps?: number; // Budget for execution steps; default 10,000
}

/**
 * Public interface for the StarScript engine.
 * Users interact with this to execute scripts.
 */
export interface StarScriptEngine {
  /**
   * Execute a script asynchronously.
   * The host invoker handles all command parsing and execution.
   * Returns a Promise that resolves with the result or rejects with an error.
   */
  execute(
    source: string,
    options?: ExecuteOptions
  ): Promise<ExecuteResult>;

  /**
   * Dispose the engine and release resources.
   * After disposal, the engine should not be used.
   */
  dispose(): void;
}

/**
 * Internal interface for WASM exports.
 * Not exported to consumers.
 */
export interface WasmExports {
  execute: (
    source: string,
    onCommand: (id: string, runId: string, rawLine: string) => void,
    onDone: (result: any) => void,
    opts?: { maxSteps?: number }
  ) => string; // Returns runId
  resolve: (id: string, value: any) => void;
  reject: (id: string, code: string, message: string) => void;
  cancel: (runId: string) => void;
}
