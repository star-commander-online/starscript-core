import { WasmStarScriptEngine } from "./engine.js";
import { loadWasmExports, type WasmLoaderOptions } from "./wasm-loader.js";
import type { EngineOptions, StarScriptEngine } from "./types.js";

export type {
  CommandCall,
  CommandHandler,
  EngineOptions,
  ExecutionResult,
  StarScriptEngine,
  StarValue,
  WasmExports,
} from "./types.js";

export async function createEngine(
  loaderOptions: WasmLoaderOptions,
  engineOptions?: EngineOptions
): Promise<StarScriptEngine> {
  const wasm = await loadWasmExports(loaderOptions);
  return new WasmStarScriptEngine(wasm, engineOptions);
}
