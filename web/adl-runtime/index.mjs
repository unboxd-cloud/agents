// ADL runtime loader.
//
// This module instantiates the Go-built WebAssembly runtime (adl.wasm) and
// exposes compile()/parse(). The Langium tooling uses this instead of its own
// generated parser, so there is a single ADL runtime shared with the Go
// backend.
//
// Works in both Node (>=18) and the browser. The Go WASM support shim
// (wasm_exec.js) is loaded from alongside this file and defines globalThis.Go.

const isNode =
  typeof process !== "undefined" &&
  process.versions != null &&
  process.versions.node != null;

async function ensureGo() {
  if (typeof globalThis.Go === "function") return;
  const shimUrl = new URL("./wasm_exec.js", import.meta.url);
  if (isNode) {
    const { readFile } = await import("node:fs/promises");
    const src = await readFile(shimUrl, "utf8");
    // wasm_exec.js assigns globalThis.Go; evaluate it in this scope.
    (0, eval)(src);
  } else {
    await import(/* @vite-ignore */ shimUrl.href);
  }
  if (typeof globalThis.Go !== "function") {
    throw new Error("adl: wasm_exec.js did not define globalThis.Go");
  }
}

async function loadWasmBytes(wasmUrl) {
  if (isNode) {
    const { readFile } = await import("node:fs/promises");
    return readFile(wasmUrl);
  }
  const resp = await fetch(wasmUrl);
  if (!resp.ok) {
    throw new Error(`adl: failed to fetch ${wasmUrl}: ${resp.status}`);
  }
  return new Uint8Array(await resp.arrayBuffer());
}

export async function loadADL(wasmUrl) {
  const url = wasmUrl ?? new URL("./adl.wasm", import.meta.url);
  await ensureGo();

  const go = new globalThis.Go();
  const bytes = await loadWasmBytes(url);
  const { instance } = await WebAssembly.instantiate(bytes, go.importObject);

  // Run the Go program; it registers the globals and then blocks on select{}.
  // We intentionally do not await go.run (it resolves only on exit).
  go.run(instance);

  const call = (fn, source) => {
    if (typeof globalThis[fn] !== "function") {
      throw new Error(`adl: runtime did not register ${fn}`);
    }
    return JSON.parse(globalThis[fn](source));
  };

  return {
    compile: (source) => call("adlCompile", source),
    parse: (source) => call("adlParse", source),
  };
}

export default loadADL;
