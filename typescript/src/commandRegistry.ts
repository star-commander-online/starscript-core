/**
 * Registry for managing host command handlers.
 * Handles both sync and async (Promise-returning) handlers transparently.
 */

import { CommandHandler, CommandContext, StarValue } from "./types";

export class CommandRegistry {
  private handlers: Map<string, CommandHandler> = new Map();

  /**
   * Register a command handler.
   * If a handler already exists for this name, it's replaced.
   */
  register(name: string, handler: CommandHandler): void {
    this.handlers.set(name, handler);
  }

  /**
   * Unregister a command handler.
   * No-op if the command is not registered.
   */
  unregister(name: string): void {
    this.handlers.delete(name);
  }

  /**
   * Check if a command handler is registered.
   */
  has(name: string): boolean {
    return this.handlers.has(name);
  }

  /**
   * Invoke a registered command handler.
   * If the handler returns a Promise, this will wait for it.
   * Throws if the command is not registered (caller should check has() first).
   */
  async invoke(context: CommandContext): Promise<StarValue> {
    const handler = this.handlers.get(context.name);
    if (!handler) {
      throw new Error(`Unknown command: ${context.name}`);
    }

    const result = handler(context);

    // Handle both sync and async handlers
    if (result instanceof Promise) {
      return await result;
    }

    return result;
  }

  /**
   * Clear all registered handlers.
   */
  clear(): void {
    this.handlers.clear();
  }

  /**
   * Get all registered command names.
   */
  getCommandNames(): string[] {
    return Array.from(this.handlers.keys());
  }
}
