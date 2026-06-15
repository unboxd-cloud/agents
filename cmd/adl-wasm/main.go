//go:build js && wasm

// Command adl-wasm exposes the canonical ADL runtime (pkg/adl) to
// JavaScript/TypeScript via WebAssembly. The Langium-based tooling loads the
// resulting adl.wasm and calls the globals registered here instead of
// shipping its own parser, so the Go backend and the editor share one runtime
// and one set of semantics.
//
// Build with:
//
//	GOOS=js GOARCH=wasm go build -o web/adl-runtime/adl.wasm ./cmd/adl-wasm
package main

import (
	"encoding/json"
	"syscall/js"

	adl "github.com/unboxd-cloud/platform/pkg/adl"
)

func main() {
	js.Global().Set("adlCompile", js.FuncOf(compile))
	js.Global().Set("adlParse", js.FuncOf(parse))
	// Block forever so the registered functions stay callable.
	select {}
}

// compile(source) -> JSON string of adl.Result {model, diagnostics}.
func compile(_ js.Value, args []js.Value) any {
	src, err := requireSource(args)
	if err != "" {
		return errorResult(err)
	}
	res := adl.Compile(src)
	return marshal(res)
}

// parse(source) -> JSON string of {model, diagnostics} without reference
// resolution (syntax only).
func parse(_ js.Value, args []js.Value) any {
	src, err := requireSource(args)
	if err != "" {
		return errorResult(err)
	}
	model, diags := adl.Parse(src)
	return marshal(adl.Result{Model: model, Diagnostics: diags})
}

func requireSource(args []js.Value) (string, string) {
	if len(args) < 1 || args[0].Type() != js.TypeString {
		return "", "adl: expected a single source string argument"
	}
	return args[0].String(), ""
}

func marshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return errorResult("adl: failed to encode result: " + err.Error())
	}
	return string(b)
}

func errorResult(msg string) string {
	b, _ := json.Marshal(adl.Result{
		Diagnostics: []adl.Diagnostic{{Severity: adl.SeverityError, Message: msg}},
	})
	return string(b)
}
