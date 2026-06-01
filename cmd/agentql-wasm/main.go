//go:build js && wasm

// Command agentql-wasm exposes the canonical AgentQL runtime (pkg/agentql) to
// JavaScript/TypeScript via WebAssembly. The Langium-based tooling loads the
// resulting agentql.wasm and calls the globals registered here instead of
// shipping its own parser, so the Go backend and the editor share one runtime
// and one set of semantics.
//
// Build with:
//
//	GOOS=js GOARCH=wasm go build -o web/agentql-runtime/agentql.wasm ./cmd/agentql-wasm
package main

import (
	"encoding/json"
	"syscall/js"

	agentql "github.com/unboxd-cloud/platform/pkg/agentql"
)

func main() {
	js.Global().Set("agentqlCompile", js.FuncOf(compile))
	js.Global().Set("agentqlParse", js.FuncOf(parse))
	// Block forever so the registered functions stay callable.
	select {}
}

// compile(source) -> JSON string of agentql.Result {model, diagnostics}.
func compile(_ js.Value, args []js.Value) any {
	src, err := requireSource(args)
	if err != "" {
		return errorResult(err)
	}
	res := agentql.Compile(src)
	return marshal(res)
}

// parse(source) -> JSON string of {model, diagnostics} without reference
// resolution (syntax only).
func parse(_ js.Value, args []js.Value) any {
	src, err := requireSource(args)
	if err != "" {
		return errorResult(err)
	}
	model, diags := agentql.Parse(src)
	return marshal(agentql.Result{Model: model, Diagnostics: diags})
}

func requireSource(args []js.Value) (string, string) {
	if len(args) < 1 || args[0].Type() != js.TypeString {
		return "", "agentql: expected a single source string argument"
	}
	return args[0].String(), ""
}

func marshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return errorResult("agentql: failed to encode result: " + err.Error())
	}
	return string(b)
}

func errorResult(msg string) string {
	b, _ := json.Marshal(agentql.Result{
		Diagnostics: []agentql.Diagnostic{{Severity: agentql.SeverityError, Message: msg}},
	})
	return string(b)
}
