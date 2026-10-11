import type {
  CommandHandler,
  EngineOptions,
  ExecutionResult,
  StarScriptEngine,
  WasmExports,
} from "./types.js";

const handlersByEngine = new Map<number, Map<string, CommandHandler>>();
let bridgeInstalled = false;

function installBridge() {
  if (bridgeInstalled) {
    return;
  }

  (globalThis as typeof globalThis & {
    __starscriptInvokeCommand?: (
      engineId: number,
      commandName: string,
      args: unknown[]
    ) => Promise<unknown>;
  }).__starscriptInvokeCommand = async (
    engineId: number,
    commandName: string,
    args: unknown[]
  ) => {
    const engineHandlers = handlersByEngine.get(engineId);
    if (!engineHandlers) {
      throw new Error(`No handler registry for engine ${engineId}`);
    }

    const handler = engineHandlers.get(commandName);
    if (!handler) {
      throw new Error(`Unknown command: ${commandName}`);
    }

    return handler({
      name: commandName,
      args: args as never,
    });
  };

  bridgeInstalled = true;
}

export class WasmStarScriptEngine implements StarScriptEngine {
  private readonly wasm: WasmExports;
  private readonly engineId: number;
  private disposed = false;

  constructor(wasm: WasmExports, options?: EngineOptions) {
    this.wasm = wasm;
    installBridge();
    this.engineId = wasm.createEngine(options);
    handlersByEngine.set(this.engineId, new Map<string, CommandHandler>());
  }

  registerCommand(name: string, handler: CommandHandler): void {
    this.assertNotDisposed();
    const handlers = handlersByEngine.get(this.engineId);
    if (!handlers) {
      throw new Error("Engine handler registry was not initialized");
    }
    handlers.set(name, handler);
  }

  async execute(source: string): Promise<ExecutionResult> {
    this.assertNotDisposed();
    return this.wasm.execute(this.engineId, source);
  }

  dispose(): void {
    if (this.disposed) {
      return;
    }
    this.wasm.disposeEngine(this.engineId);
    handlersByEngine.delete(this.engineId);
    this.disposed = true;
  }

  private assertNotDisposed(): void {
    if (this.disposed) {
      throw new Error("Engine already disposed");
    }
  }
}
