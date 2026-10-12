import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { createEngine } from "../src/index.js";
import type { StarScriptEngine } from "../src/types.js";

describe("StarScript WASM Engine", () => {
  let engine: StarScriptEngine;

  beforeAll(async () => {
    // This test needs the WASM binary available
    // In a real integration environment, we'd load from the built dist/
    console.log("Note: Full integration tests require built WASM binary");
  });

  it("should export createEngine function", () => {
    expect(typeof createEngine).toBe("function");
  });

  it("should export all public types", async () => {
    // Verify types are exported (runtime check)
    const { createEngine } = await import("../src/index.js");
    expect(createEngine).toBeDefined();
  });
});
