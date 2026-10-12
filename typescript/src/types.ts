/**
 * Types for StarScript execution and command handling.
 */

/**
 * Scalar value types supported by StarScript.
 */
export type StarValue = null | boolean | number | string;

/**
 * Command callback context with arguments and cancellation signal.
 */
export interface CommandContext {
  /** Command arguments, already interpolated and evaluated */
  args: StarValue[];
  /** Cancellation signal if execution is aborted */
  signal?: AbortSignal;
}

/**
 * Host command callback that handles command invocation from scripts.
 * Can be synchronous or return a Promise.
 * Return value (or Promise-resolved value) becomes the command result.
 * Returning undefined/void becomes null.
 */
export type StarCommand = (
  context: CommandContext
) => StarValue | void | Promise<StarValue | void>;

/**
 * Structured result from script execution.
 */
export interface ExecuteResult {
  /** True if execution succeeded, false if error occurred */
  ok: boolean;
  /** Final value returned by script (if ok === true) */
  value?: StarValue;
  /** Error details (if ok === false) */
  error?: ExecuteError;
  /** Variables remaining in engine after execution */
  variables?: Record<string, StarValue>;
  /** Number of execution steps taken */
  steps?: number;
}

/**
 * Error information from failed script execution.
 */
export interface ExecuteError {
  message: string;
  kind?: "syntax_error" | "runtime_error" | "unknown_command" | "budget_exceeded";
  line?: number;
  column?: number;
}

/**
 * Options for engine creation.
 */
export interface EngineOptions {
  /** Custom WASM binary URL. If omitted, uses bundled WASM */
  wasmUrl?: string | URL;
  /** Maximum execution steps (defaults to 10000) */
  maxSteps?: number;
  /** Maximum source size in bytes (defaults to 256KB) */
  maxSourceBytes?: number;
}

/**
 * Options for script execution.
 */
export interface ExecuteOptions {
  /** Abort signal for cancellation */
  signal?: AbortSignal;
}

/**
 * StarScript engine for executing scripts and managing commands.
 */
export interface StarScriptEngine {
  /**
   * Register a host command that scripts can invoke.
   * Replaces any existing command with the same name.
   */
  registerCommand(name: string, handler: StarCommand): void;

  /**
   * Unregister a command.
   */
  unregisterCommand(name: string): void;

  /**
   * Execute a StarScript source string.
   * Returns a Promise that resolves with the execution result.
   */
  execute(source: string, options?: ExecuteOptions): Promise<ExecuteResult>;

  /**
   * Dispose the engine and free resources.
   * No further execution is possible after disposal.
   */
  dispose(): void;
}

/**
 * Internal WASM bridge exports (not part of public API).
 */
export interface WasmExports {
  createEngine(options?: { maxSteps?: number; maxSourceBytes?: number }): number;
  disposeEngine(engineId: number): void;
  execute(engineId: number, source: string): Promise<{
    value: any;
    variables: Record<string, any>;
    steps: number;
  }>;
}
