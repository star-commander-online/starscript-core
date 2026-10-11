import type { WasmExports } from "./types.js";

export interface WasmLoaderOptions {
  wasmUrl: string;
  instantiate?: (
    url: string
  ) => Promise<void>;
}

declare global {
  interface Window {
    StarScriptWasm?: WasmExports;
  }

  // eslint-disable-next-line no-var
  var StarScriptWasm: WasmExports | undefined;
}

function readWasmExports(): WasmExports {
  const globalValue = globalThis as typeof globalThis & {
    StarScriptWasm?: WasmExports;
  };

  if (!globalValue.StarScriptWasm) {
    throw new Error(
      "StarScriptWasm is not available. Ensure wasm_exec.js and the Go WASM module are initialized."
    );
  }

  return globalValue.StarScriptWasm;
}

export async function loadWasmExports(
  options: WasmLoaderOptions
): Promise<WasmExports> {
  if (options.instantiate) {
    await options.instantiate(options.wasmUrl);
  }

  return readWasmExports();
}
