// Command platform is the unified CLI. It composes the local stack and drives
// the control plane through the same SDK applications use, so the command, the
// SDK, and agents all share one client surface.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/pkg/sdk"
)

const sandboxManifest = "deploy/sandbox/pod.yaml"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(cmd string, args []string) error {
	ctx := context.Background()
	c := sdk.New()
	c.Tenant = os.Getenv("TENANT_ID")

	switch cmd {
	case "compose":
		return compose(args)
	case "agent", "adl":
		return agentCmd(args)
	case "catalog":
		profile := ""
		if len(args) > 0 {
			profile = args[0]
		}
		offers, err := c.ListOfferings(ctx, profile)
		if err != nil {
			return err
		}
		return printJSON(offers)
	case "pricebook":
		pb, err := c.PriceBook(ctx)
		if err != nil {
			return err
		}
		return printJSON(pb)
	case "frameworks":
		fw, err := c.Frameworks(ctx)
		if err != nil {
			return err
		}
		return printJSON(fw)
	case "rate":
		var req api.RateRequest
		if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
			return fmt.Errorf("read rate request from stdin: %w", err)
		}
		resp, err := c.Rate(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(resp)
	case "evaluate":
		var req api.EvalRequest
		if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
			return fmt.Errorf("read eval request from stdin: %w", err)
		}
		rep, err := c.Evaluate(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(rep)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

// compose drives the local sandbox via a container manager (podman by default).
func compose(args []string) error {
	mgr := os.Getenv("CONTAINER")
	if mgr == "" {
		mgr = "podman"
	}
	sub := "up"
	if len(args) > 0 {
		sub = args[0]
	}
	var cmdArgs []string
	switch sub {
	case "up":
		cmdArgs = []string{"play", "kube", sandboxManifest}
	case "down":
		cmdArgs = []string{"play", "kube", "--down", sandboxManifest}
	default:
		return fmt.Errorf("compose: unknown subcommand %q (use up|down)", sub)
	}
	c := exec.Command(mgr, cmdArgs...)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func usage() {
	fmt.Fprint(os.Stderr, `platform - Unboxd platform CLI

Usage:
  platform compose up|down        run/stop the local sandbox (CONTAINER=podman|docker)
  platform agent check  <file>    parse + validate an agent (.agent) document (exit 1 on errors)
  platform agent show   <file>    print the compiled model + diagnostics as JSON
  platform agent deploy <file>    validate, then emit the deployable resolved agent as JSON
  platform agent bench  <file>    blueprint conformance benchmark (JSON-LD report)
  platform agent export <file>... export the combined data model as JSON
  platform catalog [profile]      list catalog offerings
  platform pricebook              show the active price book
  platform frameworks             list compliance frameworks
  platform rate   < req.json      rate usage (RateRequest on stdin)
  platform evaluate < req.json    evaluate compliance (EvalRequest on stdin)

Env:
  TENANT_ID   tenant for requests (X-Tenant-ID)
  CONTAINER   container manager for 'compose' (default: podman)
`)
}
