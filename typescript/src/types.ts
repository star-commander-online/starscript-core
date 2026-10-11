export type StarValue =
  | null
  | boolean
  | number
  | string
  | StarValue[]
  | { [key: string]: StarValue };

export interface CommandCall {
  name: string;
  args: StarValue[];
}

export type CommandHandler = (
  call: CommandCall
) => StarValue | Promise<StarValue>;

export interface ExecutionResult {
  value: StarValue;
  variables: Record<string, StarValue>;
  steps: number;
}

export interface EngineOptions {
  maxSteps?: number;
  maxSourceBytes?: number;
}

export interface StarScriptEngine {
  registerCommand(name: string, handler: CommandHandler): void;
  execute(source: string): Promise<ExecutionResult>;
  dispose(): void;
}

export interface WasmExports {
  createEngine(options?: EngineOptions): number;
  disposeEngine(engineId: number): void;
  execute(engineId: number, source: string): Promise<ExecutionResult>;
}
