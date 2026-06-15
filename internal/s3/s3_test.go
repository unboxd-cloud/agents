package s3

import "testing"

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		s    Settings
		ok   bool
	}{
		{"ok", Settings{Scope: "platform", Bucket: "b", Region: "us-east-1"}, true},
		{"no scope", Settings{Bucket: "b", Region: "r"}, false},
		{"no bucket", Settings{Scope: "platform", Region: "r"}, false},
		{"no region", Settings{Scope: "platform", Bucket: "b"}, false},
		{"negative retention", Settings{Scope: "platform", Bucket: "b", Region: "r", RetentionDays: -1}, false},
	}
	for _, c := range cases {
		err := c.s.Validate()
		if (err == nil) != c.ok {
			t.Errorf("%s: Validate() err=%v, want ok=%v", c.name, err, c.ok)
		}
	}
}

func TestSetGetList(t *testing.T) {
	st := NewMemStore()
	if err := st.Set(Settings{Scope: "platform", Bucket: "snaps", Region: "eu-central-1", SecretKey: "shh"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := st.Set(Settings{Scope: "acme", Bucket: "acme-snaps", Region: "eu-west-1"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := st.Get("platform")
	if !ok || got.Bucket != "snaps" {
		t.Fatalf("Get(platform) = %+v ok=%v", got, ok)
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be defaulted")
	}
	if _, ok := st.Get("missing"); ok {
		t.Error("Get(missing) should be false")
	}
	if l := st.List(); len(l) != 2 || l[0].Scope != "acme" || l[1].Scope != "platform" {
		t.Errorf("List() not sorted by scope: %+v", l)
	}
}

func TestSetRejectsInvalid(t *testing.T) {
	st := NewMemStore()
	if err := st.Set(Settings{Scope: "platform"}); err == nil {
		t.Fatal("Set should reject settings without bucket/region")
	}
}

func TestRedacted(t *testing.T) {
	s := Settings{Scope: "platform", Bucket: "b", Region: "r", SecretKey: "topsecret"}
	if r := s.Redacted(); r.SecretKey != secretMask {
		t.Errorf("Redacted secret = %q, want mask", r.SecretKey)
	}
	if s.SecretKey != "topsecret" {
		t.Error("Redacted must not mutate the original")
	}
	empty := Settings{Scope: "platform", Bucket: "b", Region: "r"}
	if empty.Redacted().SecretKey != "" {
		t.Error("empty secret should stay empty when redacted")
	}
}

func TestSecretOrKeep(t *testing.T) {
	if got := SecretOrKeep("", "existing"); got != "existing" {
		t.Errorf("blank should keep existing, got %q", got)
	}
	if got := SecretOrKeep(secretMask, "existing"); got != "existing" {
		t.Errorf("mask should keep existing, got %q", got)
	}
	if got := SecretOrKeep("new", "existing"); got != "new" {
		t.Errorf("new value should replace, got %q", got)
	}
}

func TestFromEnv(t *testing.T) {
	if _, set := FromEnv("platform"); set {
		t.Error("no env should report unset")
	}
	t.Setenv("S3_BUCKET", "env-bucket")
	t.Setenv("S3_REGION", "us-west-2")
	t.Setenv("S3_RETENTION_DAYS", "14")
	t.Setenv("S3_SNAPSHOT_SCHEDULE", "0 */6 * * *")
	s, set := FromEnv("platform")
	if !set {
		t.Fatal("S3 env vars set should report set=true")
	}
	if s.Bucket != "env-bucket" || s.Region != "us-west-2" || s.RetentionDays != 14 || s.Schedule != "0 */6 * * *" {
		t.Errorf("FromEnv = %+v", s)
	}
	if err := s.Validate(); err != nil {
		t.Errorf("env settings should validate: %v", err)
	}
}
