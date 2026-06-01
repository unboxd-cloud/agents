# ADL runtime (WASM)

**ADL** (Agent Definition Language) is the declarative DSL for describing agentic
domain models — entities and relations, brains and minds, beliefs, constitutions
and policies, HTTP APIs, functions, and SurrealDB/SurrealML bindings.

This directory packages the **canonical ADL runtime** for JavaScript/TypeScript
consumers (published as `@agennext/adl-runtime`). The runtime itself is the Go
implementation in [`pkg/adl`](../../pkg/adl); it is compiled to WebAssembly so the
editor tooling and the Go backend share **one** parser and **one** set of
semantics instead of maintaining two.

## Contents

| File | Source | Committed |
| --- | --- | --- |
| `index.mjs` | loader that instantiates the WASM and exposes `compile`/`parse` | yes |
| `index.d.ts` | TypeScript types for the AST and diagnostics | yes |
| `wasm_exec.js` | Go's WASM support shim (from the Go toolchain) | yes |
| `adl.wasm` | the compiled runtime | no — build it |

## Building

```sh
make adl-wasm
```

This runs `GOOS=js GOARCH=wasm go build` and refreshes `wasm_exec.js` from the
local Go toolchain. The `.wasm` is reproducible and therefore git-ignored.

## Usage

```ts
import { loadADL } from "./web/adl-runtime/index.mjs";

const adl = await loadADL();           // defaults to ./adl.wasm
const { model, diagnostics } = adl.compile(source);

for (const d of diagnostics) {
  console.log(`${d.severity} ${d.pos.line}:${d.pos.column} ${d.message}`);
}
```

- `compile(source)` — lex, parse, **and** resolve `[Entity]` references (the full
  pipeline).
- `parse(source)` — syntax only, no reference resolution.

Both return a `Result` (`{ model, diagnostics }`); see `index.d.ts`. Works in
Node (>=18) and the browser. In a Langium language server, call `compile` and map
each `Diagnostic` (which carries 1-based `line`/`column` plus a 0-based `offset`)
onto LSP diagnostics, replacing Langium's generated parser pass.
