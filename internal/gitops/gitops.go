// Package gitops provides a GitOps agent: it treats a directory of desired-state
// artifacts (datasets + chart values, reconciled from Git) as the source of
// truth, validating them every cycle and optionally applying them. This realizes
// "GitOps as an agent" using the shared agent runtime.
package gitops

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/catalog"
	"github.com/unboxd-cloud/platform/internal/compliance"
)

// Applier applies validated desired state to a target (e.g. helm/kubectl). When
// nil, the reconciler runs dry (validate-only) — safe by default.
type Applier interface {
	Apply(ctx context.Context, dir string) error
}

// Reconciler validates (and optionally applies) the dataset artifacts under Dir.
type Reconciler struct {
	Dir     string
	Applier Applier
}

// Name implements agent.Agent.
func (r *Reconciler) Name() string { return "gitops" }

// Reconcile validates every dataset artifact present, then applies if configured.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	if err := ValidateDatasets(r.Dir); err != nil {
		return fmt.Errorf("desired state invalid, refusing to apply: %w", err)
	}
	if r.Applier == nil {
		log.Printf("gitops: validated desired state in %s (dry-run)", r.Dir)
		return nil
	}
	return r.Applier.Apply(ctx, r.Dir)
}

// ValidateDatasets parses each known dataset file (if present) with its loader,
// so malformed artifacts are caught before they reach a cluster.
func ValidateDatasets(dir string) error {
	checks := []struct {
		file string
		load func(*os.File) error
	}{
		{"offerings.json", func(f *os.File) error { _, err := catalog.Load(f); return err }},
		{"pricebook.json", func(f *os.File) error { _, err := billing.LoadPriceBook(f); return err }},
		{"tax-rules.json", func(f *os.File) error { _, err := billing.LoadTaxTable(f); return err }},
		{"compliance-frameworks.json", func(f *os.File) error {
			_, err := compliance.NewRegistry().Load(f)
			return err
		}},
	}
	validated := 0
	for _, c := range checks {
		path := filepath.Join(dir, c.file)
		f, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("open %s: %w", c.file, err)
		}
		err = c.load(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("validate %s: %w", c.file, err)
		}
		validated++
	}
	if validated == 0 {
		return fmt.Errorf("no dataset artifacts found in %s", dir)
	}
	log.Printf("gitops: validated %d dataset artifact(s) in %s", validated, dir)
	return nil
}
