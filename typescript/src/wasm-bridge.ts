/**
 * WASM bridge that handles Go↔JS communication.
 * Provides command invocation callback from Go to JavaScript.
 */

import type { StarValue } from "./types.js";
import { CommandRegistry } from "./command-registry.js";

/**
 * Command invocation payload from Go WASM.
 */
export interface CommandCall {
  name: string;
  args: StarValue[];
}

/**
 * Bridge that coordinates command invocation between Go and JavaScript.
 */
export class WasmBridge {
  private registry: CommandRegistry;
  private abortSignal?: AbortSignal;

  constructor(registry: CommandRegistry) {
    this.registry = registry;
  }

  /**
   * Set up the command invocation callback that Go WASM will call.
   * This must be set before Go tries to invoke any commands.
   */
  setupCallback(): void {
    const globalValue = globalThis as any;

    // Go WASM will call this when it needs to invoke a command
    globalValue.__starscriptInvokeCommand = async (call: CommandCall): Promise<any> => {
      try {
        const result = await this.registry.invoke(call.name, call.args, this.abortSignal);
        return { ok: true, value: result };
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return { ok: false, error: message };
      }
    };
  }

  /**
   * Set the abort signal for current execution.
   */
  setAbortSignal(signal?: AbortSignal): void {
    this.abortSignal = signal;
  }
}
