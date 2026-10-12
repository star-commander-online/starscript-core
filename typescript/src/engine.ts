/**
 * TypeScript wrapper for the Go WASM StarScript engine.
 */

import type {
  StarScriptEngine,
  StarCommand,
  ExecuteResult,
  ExecuteOptions,
  ExecuteError,
  StarValue,
  WasmExports,
  EngineOptions,
} from "./types.js";
import { CommandRegistry } from "./command-registry.js";
import { WasmBridge } from "./wasm-bridge.js";

/**
 * Implementation of StarScriptEngine using Go WASM backend.
 */
export class WasmStarScriptEngine implements StarScriptEngine {
  private registry: CommandRegistry;
  private bridge: WasmBridge;
  private wasm: WasmExports;
  private engineId: number;
  private disposed = false;

  constructor(wasm: WasmExports, engineId: number) {
    this.wasm = wasm;
    this.engineId = engineId;
    this.registry = new CommandRegistry();
    this.bridge = new WasmBridge(this.registry);
    this.bridge.setupCallback();
  }

  registerCommand(name: string, handler: StarCommand): void {
    this.checkDisposed();
    this.registry.register(name, handler);
  }

  unregisterCommand(name: string): void {
    this.checkDisposed();
    this.registry.unregister(name);
  }

  async execute(source: string, options?: ExecuteOptions): Promise<ExecuteResult> {
    this.checkDisposed();

    // Set abort signal for command invocations during execution
    this.bridge.setAbortSignal(options?.signal);

    try {
      const result = await this.wasm.execute(this.engineId, source);

      return {
        ok: true,
        value: result.value,
        variables: result.variables,
        steps: result.steps,
      };
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      const executeError: ExecuteError = {
        message,
        kind: this.classifyError(message),
      };

      return {
        ok: false,
        error: executeError,
      };
    }
  }

  dispose(): void {
    if (!this.disposed) {
      this.wasm.disposeEngine(this.engineId);
      this.registry.clear();
      this.disposed = true;
    }
  }

  private checkDisposed(): void {
    if (this.disposed) {
      throw new Error("Engine has been disposed");
    }
  }

  private classifyError(message: string): ExecuteError["kind"] {
    if (message.includes("syntax") || message.includes("parse")) {
      return "syntax_error";
    } else if (message.includes("unknown command")) {
      return "unknown_command";
    } else if (message.includes("budget") || message.includes("exceeds")) {
      return "budget_exceeded";
    }
    return "runtime_error";
  }
}
