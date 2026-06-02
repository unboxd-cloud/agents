package main

import (
	"fmt"
	"os"

	"github.com/unboxd-cloud/platform/pkg/adl"
)

// adlCmd implements `platform adl <check|show> <file>` — the agent language,
// runnable from the command line through the same runtime the Go backend and the
// editor tooling (via WASM) use.
func adlCmd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: platform adl <check|show> <file>")
	}
	sub, path := args[0], args[1]
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	res := adl.Compile(string(src))

	switch sub {
	case "check":
		for _, d := range res.Diagnostics {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n",
				path, d.Pos.Line, d.Pos.Column, d.Severity, d.Message)
		}
		if res.HasErrors() {
			return fmt.Errorf("%s: validation failed", path)
		}
		fmt.Printf("%s: ok (%d declarations, %d diagnostics)\n",
			path, len(res.Model.Declarations), len(res.Diagnostics))
		return nil
	case "show":
		return printJSON(res)
	default:
		return fmt.Errorf("adl: unknown subcommand %q (use check|show)", sub)
	}
}
