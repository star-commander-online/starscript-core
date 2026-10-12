/**
 * Command registry for managing host commands.
 */

import type { StarCommand, CommandContext, StarValue } from "./types.js";

/**
 * Registry that manages command callbacks and coordinates invocation from WASM.
 */
export class CommandRegistry {
  private commands = new Map<string, StarCommand>();

  /**
   * Register a command handler.
   */
  register(name: string, handler: StarCommand): void {
    this.commands.set(name, handler);
  }

  /**
   * Unregister a command handler.
   */
  unregister(name: string): void {
    this.commands.delete(name);
  }

  /**
   * Invoke a registered command.
   * Handles both sync and async handlers transparently.
   */
  async invoke(name: string, args: StarValue[], signal?: AbortSignal): Promise<StarValue> {
    const handler = this.commands.get(name);
    if (!handler) {
      throw new Error(`Unknown command: ${name}`);
    }

    const context: CommandContext = { args, signal };
    const result = handler(context);

    // Handle Promise returns
    if (result instanceof Promise) {
      return result.then((value) => value === undefined || value === void 0 ? null : value);
    }

    // Handle synchronous returns
    return result === undefined || result === void 0 ? null : result;
  }

  /**
   * Clear all registered commands.
   */
  clear(): void {
    this.commands.clear();
  }
}
