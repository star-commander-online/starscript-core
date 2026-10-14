/**
 * Type declaration for wasm_exec.js Go runtime.
 * Provides type safety for the global Go class.
 */

declare global {
  class Go {
    constructor();
    importObject: WebAssembly.Imports;
    run(instance: WebAssembly.Instance): Promise<void>;
  }
}

export {};
