package cloudstack

import "testing"

func TestDefaultMappingAnchorsOnCloudStack(t *testing.T) {
	m := DefaultMapping()
	if m.Anchor != "Apache CloudStack" {
		t.Fatalf("unexpected anchor: %s", m.Anchor)
	}
	if m.Provider != "cloudstack" {
		t.Fatalf("unexpected provider: %s", m.Provider)
	}
	if len(m.Primitives) == 0 {
		t.Fatal("expected primitive mappings")
	}
}

func TestDefaultMappingIncludesTenantAndMeteringSources(t *testing.T) {
	m := DefaultMapping()
	seenAccount := false
	seenEvent := false
	for _, p := range m.Primitives {
		if p.CloudStack == "account" && p.Unboxd == "tenant account" {
			seenAccount = true
		}
		if p.CloudStack == "event" && p.Unboxd == "audit and metering source" {
			seenEvent = true
		}
	}
	if !seenAccount {
		t.Fatal("expected account -> tenant account mapping")
	}
	if !seenEvent {
		t.Fatal("expected event -> audit and metering source mapping")
	}
}
