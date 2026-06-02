package main

import (
	"fmt"
	"os"
	"time"

	"github.com/unboxd-cloud/platform/pkg/adl"
)

// agentCmd implements `platform agent <check|show|deploy|bench|export> <file>` —
// the agent language, runnable from the command line through the same runtime the
// Go backend and the editor tooling (via WASM) use. Agent definitions are .agent
// files (e.g. platform.agent, product.agent, team.agent).
func agentCmd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: platform agent <check|show|deploy|bench|export> <file>...")
	}
	sub := args[0]
	if sub == "export" {
		return exportModel(args[1:])
	}
	path := args[1]
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
	case "bench":
		printDiags()
		if res.HasErrors() {
			return fmt.Errorf("%s: cannot benchmark, validation failed", path)
		}
		return benchAgent(path, adl.NewAgent(res.Model))
	default:
		return fmt.Errorf("agent: unknown subcommand %q (use check|show|deploy|bench)", sub)
	}
}

// exportModel compiles each .agent file and emits the combined data model as a
// single JSON document, so other repositories and users can consume the model.
func exportModel(paths []string) error {
	type fileModel struct {
		Path        string     `json:"path"`
		Valid       bool       `json:"valid"`
		Diagnostics int        `json:"diagnostics"`
		Model       *adl.Model `json:"model"`
	}
	type doc struct {
		GeneratedAt time.Time   `json:"generatedAt"`
		Count       int         `json:"count"`
		Models      []fileModel `json:"models"`
	}
	d := doc{GeneratedAt: time.Now().UTC()}
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		res := adl.Compile(string(src))
		d.Models = append(d.Models, fileModel{
			Path:        p,
			Valid:       !res.HasErrors(),
			Diagnostics: len(res.Diagnostics),
			Model:       res.Model,
		})
	}
	d.Count = len(d.Models)
	return printJSON(d)
}

// benchAgent runs a blueprint conformance benchmark: it scores the agent
// definition against the governance/decision dimensions it should declare, and
// emits a JSON-LD (schema.org) report. This is a structural, reproducible
// conformance check of the blueprint — not a task or cross-platform benchmark.
func benchAgent(path string, ag *adl.Agent) error {
	have := map[string]bool{}
	for _, c := range ag.Constitutions {
		for _, r := range c.Rules {
			have[r.Name] = true
		}
	}
	// The governance/decision dimensions a certified platform blueprint declares.
	required := []string{
		"governedAll", "auditable", "completeContext", "accountability",
		"explainable", "firstPrinciples", "innovation", "conservation",
		"socialTrust", "trustFirst", "deliverable", "northStar", "contextualAI",
	}
	covered := 0
	missing := []string{}
	for _, r := range required {
		if have[r] {
			covered++
		} else {
			missing = append(missing, r)
		}
	}
	score := float64(covered) / float64(len(required))

	fmt.Fprintf(os.Stderr, "benchmark %s (blueprint conformance):\n", path)
	fmt.Fprintf(os.Stderr, "  shape: declarations=%d entities=%d relations=%d brains=%d policies=%d functions=%d vocabularies=%d\n",
		len(ag.Model.Declarations), len(ag.Entities), len(ag.Relations), len(ag.Brains),
		len(ag.Policies), len(ag.Functions), len(ag.Vocabularies))
	fmt.Fprintf(os.Stderr, "  conformance: %d/%d dimensions (%.0f%%)\n", covered, len(required), score*100)
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "  missing: %v\n", missing)
	}

	report := map[string]any{
		"@context":             "https://schema.org",
		"@type":                "Dataset",
		"name":                 "blueprint conformance benchmark: " + path,
		"measurementTechnique": "ADL static conformance (structural); not a task or cross-platform benchmark",
		"variableMeasured": []map[string]any{
			{"@type": "PropertyValue", "name": "conformanceScore", "value": score},
			{"@type": "PropertyValue", "name": "dimensionsCovered", "value": covered, "maxValue": len(required)},
			{"@type": "PropertyValue", "name": "entities", "value": len(ag.Entities)},
			{"@type": "PropertyValue", "name": "relations", "value": len(ag.Relations)},
			{"@type": "PropertyValue", "name": "policies", "value": len(ag.Policies)},
			{"@type": "PropertyValue", "name": "functions", "value": len(ag.Functions)},
			{"@type": "PropertyValue", "name": "constitutionRules", "value": len(have)},
		},
	}
	return printJSON(report)
}
