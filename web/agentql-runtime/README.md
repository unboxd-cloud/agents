# AgentQL runtime (WASM)

This directory packages the **canonical AgentQL runtime** for JavaScript/TypeScript
consumers. The runtime itself is the Go implementation in
[`pkg/agentql`](../../pkg/agentql); it is compiled to WebAssembly so the editor
tooling and the Go backend share **one** parser and **one** set of semantics
instead of maintaining two.

## Contents

| File | Source | Committed |
| --- | --- | --- |
| `index.mjs` | loader that instantiates the WASM and exposes `compile`/`parse` | yes |
| `index.d.ts` | TypeScript types for the AST and diagnostics | yes |
| `wasm_exec.js` | Go's WASM support shim (from the Go toolchain) | yes |
| `agentql.wasm` | the compiled runtime | no — build it |

## Building

```sh
make agentql-wasm
```

This runs `GOOS=js GOARCH=wasm go build` and refreshes `wasm_exec.js` from the
local Go toolchain. The `.wasm` is reproducible and therefore git-ignored.

## Usage

```ts
import { loadAgentQL } from "./web/agentql-runtime/index.mjs";

const agentql = await loadAgentQL();           // defaults to ./agentql.wasm
const { model, diagnostics } = agentql.compile(source);

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
