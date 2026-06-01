package gitops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDatasets_Valid(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pricebook.json", `{"version":"v1","prices":{}}`)
	write(t, dir, "offerings.json", `[{"id":"x","name":"X"}]`)
	if err := ValidateDatasets(dir); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateDatasets_Malformed(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pricebook.json", `{"prices":{}}`) // missing version
	if err := ValidateDatasets(dir); err == nil {
		t.Fatal("expected validation error for missing version")
	}
}

func TestValidateDatasets_Empty(t *testing.T) {
	if err := ValidateDatasets(t.TempDir()); err == nil {
		t.Fatal("expected error when no artifacts present")
	}
}

func TestReconcile_DryRun(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "offerings.json", `[{"id":"x","name":"X"}]`)
	r := &Reconciler{Dir: dir}
	if err := r.Reconcile(context.Background()); err != nil {
		t.Fatalf("dry-run reconcile failed: %v", err)
	}
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
