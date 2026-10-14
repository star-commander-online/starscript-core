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
 * Context passed to a command handler.
 */
export interface CommandContext {
  name: string;
  args: StarValue[];
  runId: string; // The execution ID that issued this command
}

/**
 * Handler for a command.
 * Can be synchronous or return a Promise for async operations.
 * Must return a StarValue (scalar only).
 */
export type CommandHandler = (
  context: CommandContext
) => StarValue | Promise<StarValue>;

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
  code: string; // e.g., "parse_error", "runtime_error", "unknown_command", "budget_exceeded"
  message: string;
  line: number; // 0 if not applicable
}

/**
 * Options for engine creation.
 */
export interface EngineOptions {
  wasmUrl?: string; // URL to starscript.wasm; auto-detected if omitted
}

/**
 * Options for script execution.
 */
export interface ExecuteOptions {
  maxSteps?: number; // Budget for execution steps; default 10,000
}

/**
 * Public interface for the StarScript engine.
 * Users interact with this to execute scripts and manage commands.
 */
export interface StarScriptEngine {
  /**
   * Register a command handler.
   * If a handler already exists for this name, it's replaced.
   */
  registerCommand(name: string, handler: CommandHandler): void;

  /**
   * Unregister a command handler.
   * No-op if the command is not registered.
   */
  unregisterCommand(name: string): void;

  /**
   * Execute a script asynchronously.
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
    onCommand: (id: string, runId: string, name: string, args: any[]) => void,
    onDone: (result: any) => void,
    opts?: { maxSteps?: number }
  ) => string; // Returns runId
  resolve: (id: string, value: any) => void;
  reject: (id: string, code: string, message: string) => void;
  cancel: (runId: string) => void;
}
