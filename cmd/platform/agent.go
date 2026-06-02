package main

import (
	"fmt"
	"os"

	"github.com/unboxd-cloud/platform/pkg/adl"
)

// agentCmd implements `platform agent <check|show|deploy> <file>` — the agent
// language, runnable from the command line through the same runtime the Go
// backend and the editor tooling (via WASM) use. Agent definitions are .agent
// files (e.g. platform.agent, product.agent, team.agent).
func agentCmd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: platform agent <check|show|deploy> <file>")
	}
	sub, path := args[0], args[1]
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	res := adl.Compile(string(src))

	printDiags := func() {
		for _, d := range res.Diagnostics {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n",
				path, d.Pos.Line, d.Pos.Column, d.Severity, d.Message)
		}
	}

	switch sub {
	case "check":
		printDiags()
		if res.HasErrors() {
			return fmt.Errorf("%s: validation failed", path)
		}
		fmt.Printf("%s: ok (%d declarations, %d diagnostics)\n",
			path, len(res.Model.Declarations), len(res.Diagnostics))
		return nil
	case "show":
		return printJSON(res)
	case "deploy":
		printDiags()
		if res.HasErrors() {
			return fmt.Errorf("%s: cannot deploy, validation failed", path)
		}
		ag := adl.NewAgent(res.Model)
		fmt.Fprintf(os.Stderr,
			"deploying %s: %d entities, %d relations, %d brains, %d minds, %d beliefs, %d policies, %d apis, %d functions\n",
			path, len(ag.Entities), len(ag.Relations), len(ag.Brains), len(ag.Minds),
			len(ag.Beliefs), len(ag.Policies), len(ag.Apis), len(ag.Functions))
		return printJSON(ag)
	default:
		return fmt.Errorf("agent: unknown subcommand %q (use check|show|deploy)", sub)
	}
}
