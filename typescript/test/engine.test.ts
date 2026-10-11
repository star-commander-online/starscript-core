import { describe, expect, it } from "vitest";
import { WasmStarScriptEngine } from "../src/engine.js";
import type { StarValue, WasmExports } from "../src/types.js";

describe("WasmStarScriptEngine", () => {
  it("awaits async command handlers and resumes execution order", async () => {
    let nextId = 1;
    const states = new Map<number, Record<string, unknown>>();

    const wasm: WasmExports = {
      createEngine() {
        const id = nextId++;
        states.set(id, {});
        return id;
      },
      disposeEngine(engineId: number) {
        states.delete(engineId);
      },
      async execute(engineId: number, source: string) {
        if (!states.has(engineId)) {
          throw new Error("missing engine");
        }

        const lines = source.split("\n");
        const vars = states.get(engineId)!;

        for (const line of lines) {
          if (line.startsWith("target =")) {
            vars.target = "sector-7";
          }
          if (line.startsWith("factor =")) {
            vars.factor = 2;
          }
          if (line.startsWith("delayed_echo")) {
            const bridge = (globalThis as typeof globalThis & {
              __starscriptInvokeCommand: (
                id: number,
                name: string,
                args: unknown[]
              ) => Promise<unknown>;
            }).__starscriptInvokeCommand;
            vars.result = await bridge(engineId, "delayed_echo", [vars.target, vars.factor]);
          }
          if (line.startsWith("return")) {
            return {
              value: vars.result as StarValue,
              variables: vars as Record<string, StarValue>,
              steps: 4,
            };
          }
        }

        return {
          value: null,
          variables: vars as Record<string, StarValue>,
          steps: 0,
        };
      },
    };

    const engine = new WasmStarScriptEngine(wasm);
    engine.registerCommand("delayed_echo", async ({ args }) => {
      const [target, factor] = args;
      await new Promise((resolve) => setTimeout(resolve, 5));
      return `${target}:${factor}`;
    });

    const result = await engine.execute([
      'target = "sector-7"',
      "factor = 2",
      "delayed_echo $target $factor",
      "return $result",
    ].join("\n"));

    expect(result.value).toBe("sector-7:2");
    expect(result.variables.result).toBe("sector-7:2");

    engine.dispose();
  });
});
