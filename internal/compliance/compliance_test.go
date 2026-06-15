package compliance

import (
	"strings"
	"testing"
)

const sampleSpecs = `[
  {"framework":"GDPR","name":"GDPR","category":"privacy","authority":"EU","regions":["EU-DE","EU-FR"],"requiresEncryption":true},
  {"framework":"SOC2","name":"SOC2","category":"security","authority":"AICPA","requiresEncryption":false}
]`

func loadedRegistry(t *testing.T) *Registry {
	t.Helper()
	r := NewRegistry()
	n, err := r.Load(strings.NewReader(sampleSpecs))
	if err != nil || n != 2 {
		t.Fatalf("load specs: n=%d err=%v", n, err)
	}
	return r
}

func TestEvaluate_Compliant(t *testing.T) {
	reg := loadedRegistry(t)
	p := Profile{TenantID: "t1", Frameworks: []string{"GDPR"}, DataResidency: []string{"EU-DE", "EU-FR"}}
	pl := Placement{Region: "EU-DE", OfferingID: "object-storage", Certifications: []string{"GDPR"}, Encrypted: true}
	rep := Evaluate(p, pl, reg)
	if !rep.Compliant {
		t.Fatalf("expected compliant, got findings: %+v", rep.Findings)
	}
}

func TestEvaluate_ResidencyViolation(t *testing.T) {
	reg := loadedRegistry(t)
	p := Profile{TenantID: "t1", Frameworks: []string{"GDPR"}, DataResidency: []string{"EU-DE"}}
	pl := Placement{Region: "US-CA", Certifications: []string{"GDPR"}, Encrypted: true}
	rep := Evaluate(p, pl, reg)
	if rep.Compliant {
		t.Error("expected residency violation")
	}
}

func TestEvaluate_MissingCertification(t *testing.T) {
	reg := loadedRegistry(t)
	p := Profile{TenantID: "t1", Frameworks: []string{"GDPR"}}
	pl := Placement{Region: "EU-DE", Certifications: []string{"SOC2"}, Encrypted: true}
	rep := Evaluate(p, pl, reg)
	if rep.Compliant {
		t.Error("expected missing-certification violation")
	}
}

func TestEvaluate_EncryptionRequired(t *testing.T) {
	reg := loadedRegistry(t)
	p := Profile{TenantID: "t1", Frameworks: []string{"GDPR"}}
	pl := Placement{Region: "EU-DE", Certifications: []string{"GDPR"}, Encrypted: false}
	rep := Evaluate(p, pl, reg)
	if rep.Compliant {
		t.Error("GDPR requires encryption-at-rest")
	}
}

func TestStore(t *testing.T) {
	s := NewMemStore()
	if err := s.Set(Profile{TenantID: "t1", Jurisdiction: "EU-DE"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(Profile{}); err == nil {
		t.Error("expected error for missing tenantId")
	}
	if _, ok := s.Get("t1"); !ok {
		t.Error("profile not stored")
	}
}
